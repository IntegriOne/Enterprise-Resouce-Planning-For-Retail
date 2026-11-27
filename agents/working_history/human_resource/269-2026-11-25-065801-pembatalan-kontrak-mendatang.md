# Pembatalan kontrak mendatang

`TerminateWorker` menetapkan status cancelled pada kontrak bertanda tangan yang dijadwalkan mulai setelah hari kerja terakhir. Jadwal dan bukti tanda tangan tetap tersimpan. Kontrak yang sudah berjalan tetap ditutup pada hari kerja terakhir. Tes gRPC PostgreSQL dengan race detector dan go vet lulus. Penempatan mendatang masih memblokir terminasi. [Laporan pengujian](../../e2e/human_resource/41-personnel-mutation-terminate-worker.md)
