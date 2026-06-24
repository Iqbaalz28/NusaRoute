# ML Analytics — Lapisan Prediktif Data Warehouse

Tiga model machine-learning ringan berjalan **di atas star schema yang sama**
(`fact_pengiriman` + dimensi) di dalam `analytics-service`. Semuanya **pure-Go,
tanpa dependensi baru** (tidak ada Python/TensorFlow/layanan eksternal), dan
**deterministik** — snapshot warehouse yang sama selalu menghasilkan output sama.

Kode:
- Algoritma: `services/analytics-service/internal/ml/ml.go` (+ `ml_test.go`)
- Query + orkestrasi: `services/analytics-service/internal/repository/analytics_ml.go`
- Endpoint: `GET /api/v1/analytics/ml/{forecast,segments,risk}` (ADMIN, via gateway)
- UI: bagian "🤖 ML Analytics" di `front-end/components/AnalyticsPage.tsx`

Alurnya identik untuk ketiganya: SQL mengagregasi fakta → hasilnya jadi input
model Go → model mengembalikan prediksi/segmen/skor → JSON → kartu dashboard.

---

## 1. Peramalan Permintaan & Pendapatan — Regresi Linear (OLS)

**Untuk apa.** Memperkirakan jumlah paket & pendapatan **3 bulan ke depan** agar
admin bisa merencanakan kapasitas (kurir, armada, kapasitas hub) sebelum
permintaan datang. Juga menandai bulan **anomali** (lonjakan/anjlok tak wajar).

**Cara kerja.**
1. SQL mengambil deret bulanan `(paket, pendapatan)` (Jan–Jun).
2. `FitLine` mencari garis `y = a·x + b` dengan **Ordinary Least Squares** —
   meminimalkan jumlah kuadrat selisih titik aktual ke garis. Menghasilkan:
   - **slope** `a` = arah & kecepatan tren (naik/turun per bulan),
   - **intercept** `b`,
   - **R²** = seberapa baik garis menjelaskan data (0–1; ditampilkan "Akurasi Model"),
   - **σ residual** = sebaran titik terhadap garis.
3. **Proyeksi**: garis dievaluasi di bulan ke-7, 8, 9 → batang "Prediksi".
4. **Deteksi anomali**: bulan yang jaraknya **> 2σ** dari garis tren ditandai
   merah (metode *control chart* / residual outlier).

**Baca hasilnya.** Tren Naik/Turun/Stabil, prediksi paket & pendapatan bulan
depan, pertumbuhan MoM, dan R%. R² rendah ⇒ permintaan fluktuatif, prediksi
kurang andal (jujur ditampilkan apa adanya).

---

## 2. Segmentasi Wilayah — K-Means Clustering (k=3)

**Untuk apa.** Mengelompokkan provinsi tujuan ke **3 tier** secara otomatis
("Bintang/Prioritas", "Berkembang", "Perlu Perhatian") supaya strategi berbeda
per tier: jaga SLA di tier Bintang, dorong promosi di Berkembang, perbaiki
operasional di Perlu Perhatian.

**Cara kerja.**
1. SQL menghitung per provinsi: `[volume paket, pendapatan, success-rate]`.
2. **Standardisasi (z-score)** tiap fitur — penting karena pendapatan (ratusan
   juta) dan success-rate (0–100) beda skala; tanpa ini pendapatan akan
   mendominasi jarak.
3. **K-Means (algoritma Lloyd)**: tiap titik ditempel ke centroid terdekat
   (jarak Euclidean), centroid dihitung ulang = rata-rata anggotanya, diulang
   sampai stabil. **Seeding deterministik** (titik diurutkan lalu diambil merata)
   menggantikan inisialisasi acak → hasil reprodusibel.
4. Klaster diberi label tier berdasarkan **peringkat rata-rata pendapatannya**.

**Baca hasilnya.** Kartu ringkasan tiap tier (jumlah provinsi, rata-rata
pendapatan & success-rate) + daftar provinsi berlabel warna tier.

---

## 3. Skor Risiko Kegagalan per Rute — Empirical-Bayes Smoothing

**Untuk apa.** Mengurutkan rute (**layanan × provinsi tujuan**) berdasarkan
**probabilitas paket gagal** (CANCELLED/DELIVERY_FAILED/RETURN), agar admin
fokus memperbaiki rute paling berisiko lebih dulu.

**Cara kerja.**
1. SQL menghitung per rute: total paket & paket berstatus kategori "Gagal".
2. Rasio mentah `gagal/total` **menyesatkan** untuk sampel kecil (rute 2 paket
   dengan 1 gagal = 50% padahal tak signifikan). Maka dipakai **penghalusan
   empirical-Bayes (Beta–Binomial)**:

   ```
   skor = (gagal + m · rata_global) / (total + m)
   ```

   `rata_global` = tingkat gagal seluruh jaringan (prior), `m = 80` = bobot
   prior. Rute bervolume kecil **ditarik mendekati rata-rata jaringan**; rute
   bervolume besar tetap dekat rasio aslinya. Inilah inti *shrinkage estimator*.
3. Tiap rute dilabeli **Tinggi / Sedang / Rendah** relatif terhadap rata-rata
   global (≥1.4× = Tinggi, ≥0.8× = Sedang), lalu diurutkan dari paling berisiko.

**Baca hasilnya.** `raw_rate` (mentah) vs `risk_score` (dihaluskan) — selisihnya
menunjukkan koreksi sampel kecil. Bar + badge level di dashboard.

---

## Catatan

- Model dihitung **on-the-fly** saat endpoint dipanggil (data warehouse kecil,
  cukup cepat). Bila fakta tumbuh besar, tinggal cache hasil atau materialize.
- Kualitas prediksi mengikuti kualitas data: deret bulanan saat ini pendek
  (6 titik), jadi forecast bersifat indikatif. Menambah granularitas waktu
  (mingguan/harian) di seed/ETL langsung meningkatkan ketajaman model #1 & #3.
- Semua perhitungan transparan & dapat diaudit (bukan black box) — cocok untuk
  konteks edukatif/portfolio.
