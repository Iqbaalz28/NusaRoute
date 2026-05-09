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