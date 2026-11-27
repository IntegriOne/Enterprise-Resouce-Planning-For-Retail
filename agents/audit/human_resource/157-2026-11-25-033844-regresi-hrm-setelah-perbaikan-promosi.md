# Regresi HRM setelah perbaikan Personnel dan Promotion

Seluruh modul HRM diuji dengan `go test -v -p 1 -tags=integration -race -count=1 ./...` pada PostgreSQL sementara dan Redis TLS terisolasi. Sebanyak 67 paket menjalankan tes dan lulus. Log memuat 276 tes tingkat atas yang lulus tanpa skip. Build seluruh modul dan go vet dengan tag integration juga lulus.

Cakupan mencakup registrasi RPC, keamanan sesi, dan alur promosi yang menutup penempatan lama secara atomik. Transfer umum, aksi yang bergantung pada DRM/Finance, serta E2E melalui IAM dan Gateway aktual masih terbuka. Pemeriksaan diff seluruh HRM masih melaporkan baris kosong akhir pada perubahan .env.example yang sudah ada.
