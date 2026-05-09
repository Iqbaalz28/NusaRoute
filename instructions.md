Untuk tubes tahap dua, yang perlu dibuat adalah unit test  (xxx_test.go) dan functional test. Unit test tidak mengakses database atau web service lain (bisa  disimulasikan dengan library seperti gomuck). Sedangkan functional test bisa mengakses database. Tentu yang dikumpulkan saat dijalankan testnya akan failed semua karena isi codenya belum selesai. 

Posisi unit test dan functional test nanti dalam pipeline adalah sbb (nanti bisa dijalankan Jenkins):
1. Checkout repo
2. Unit Tests  (go test)
3. Lint/Vet (go vet)
4. Build Image  (lokal)
5. Functional Tests (test app di lokal atau staging)
6. Push image
7. Deploy di kubernetes
8. Verify

working list:
Viewed instructions.md:1-11
Listed directory back-end
Listed directory services
Listed directory order-service
Listed directory internal
Listed directory service
Viewed order_service_test.go:1-221
Listed directory deployments
Listed directory scripts
Viewed Jenkinsfile:1-250
Listed directory service
Viewed order_service.go:1-199
Searched for "functional"
Searched for "//go:build functional"
Searched for "functional"
Listed directory service
Viewed user_service_test.go:1-173

Berdasarkan analisis saya terhadap file `instructions.md` dan kondisi kode saat ini, berikut adalah hal-hal yang **masih perlu Anda kerjakan** atau lengkapi untuk memenuhi syarat tugas:

### 1. Pembuatan Functional Tests (Prioritas Tinggi)
Dalam `instructions.md` disebutkan bahwa *functional test* harus bisa mengakses database. Sementara itu:
*   Saya menemukan banyak **Unit Tests** (dengan tag `//go:build unit`), tetapi saya **tidak menemukan satu pun Functional Tests** (dengan tag `//go:build functional`).
*   `Jenkinsfile` Anda dikonfigurasi untuk menjalankan `go test -tags=functional`. Jika file ini tidak ada, tahap testing di pipeline Jenkins kemungkinan besar akan "pass" tanpa benar-benar mengetes apa pun atau justru gagal jika script mengharuskan ada test yang dijalankan.
*   **Tugas Anda**: Buat file `xxx_functional_test.go` di layanan utama (seperti `order-service` atau `user-service`) yang melakukan koneksi asli ke database (bukan mock).

### 2. Penyesuaian Path di Jenkinsfile
Di dalam `Jenkinsfile`, terdapat banyak perintah `dir('Tubes/NusaRoute/back-end')`. 
*   Jika struktur repository Anda di Jenkins nanti tidak memiliki folder `Tubes/NusaRoute/`, maka pipeline akan gagal karena tidak menemukan direktori tersebut.
*   **Tugas Anda**: Pastikan path di `Jenkinsfile` sesuai dengan struktur folder di repository git Anda (misalnya langsung `dir('back-end')` jika folder `back-end` berada di root repo).

### 3. Implementasi Kode yang "Belum Selesai"
Instruksi menyebutkan: *"Tentu yang dikumpulkan saat dijalankan testnya akan failed semua karena isi codenya belum selesai."*
*   Ini adalah instruksi yang unik. Artinya, Anda diminta mengumpulkan kode dalam kondisi "setengah jadi" di mana test-nya sudah ada tapi gagal dijalankan. 
*   Jika saat ini test Anda (Unit Test) sudah **PASS**, Anda mungkin perlu memastikan logika bisnis yang kompleks di dalam `internal/service` memang sengaja dikosongkan atau dibuat agar gagal sesuai permintaan asisten/dosen (untuk tujuan demonstrasi pipeline).

### 4. Pengecekan Lint & Vet
Pipeline Jenkins Anda memiliki tahap `go vet`.
*   **Tugas Anda**: Jalankan `go vet ./...` secara lokal di setiap folder service untuk memastikan tidak ada kesalahan penulisan kode yang akan menghentikan pipeline di tengah jalan.

### 5. Deployment Kubernetes
Anda memiliki file `kantin-k8s.yaml` (mungkin dari project lain) dan folder `back-end/deployments/k8s`. 
*   **Tugas Anda**: Pastikan manifest Kubernetes di dalam folder `back-end/deployments/k8s` sudah menggunakan nama image yang sesuai dengan yang di-build oleh Jenkins (`nusaroute/order-service:latest`, dsb).

**Ringkasan**: Fokus utama Anda sekarang adalah membuat **Functional Tests** di backend Go, karena itulah bagian yang paling krusial dan belum terlihat implementasinya di folder `services/`.

