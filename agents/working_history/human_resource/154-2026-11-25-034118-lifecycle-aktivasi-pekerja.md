# Catatan lifecycle pada aktivasi pekerja

ActivateWorker menyimpan event hire bersama perubahan preboarding menjadi active dan outbox dalam transaksi yang sama. Waktu pencatatan berasal dari server, tanggal efektif dari hire_date, dan principal dari actor yang lolos otorisasi. Aktivasi ulang ditolak dan tidak menambah event.

Tes gRPC dengan PostgreSQL sementara dan Redis TLS membuktikan satu event hire setelah rangkaian aktivasi yang gagal, berhasil, dan diulang. Tes service dengan race detector dan go vet bertag integration lulus.
