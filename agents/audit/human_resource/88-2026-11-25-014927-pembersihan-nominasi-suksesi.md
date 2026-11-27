# Pembersihan nominasi suksesi

Perpindahan legal entity kandidat sebelumnya membuat nominasi draf tidak dapat diperbarui maupun dihapus. Validator sekarang memisahkan pemeriksaan kelayakan penulisan dari penghapusan draf. Penghapusan tetap mengunci dan memeriksa rencana, mempertahankan pembatasan rencana disetujui dan otorisasi dari resource. Perubahan nominasi tetap ditolak jika legal entity pekerja berbeda dari posisi.

Integrasi gRPC ke PostgreSQL mereproduksi kegagalan penghapusan sebelum perbaikan dan membuktikan penghapusan serta satu event outbox berhasil setelah perbaikan. Tes unit dengan race detector dan go vet juga lulus.
