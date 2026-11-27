# Repository permintaan transfer penempatan

Model dan registry HRM kini mencakup permintaan transfer. Repository menyimpan pengajuan, mencari berdasarkan idempotency key, mengunci permintaan dalam transaksi, dan memperbarui hanya keputusan dari status submitted. Integrasi PostgreSQL membuktikan penguncian dalam transaksi, penolakan perubahan keputusan terminal, dan rollback. Go test dan vet modul lulus. Generator penuh masih terhenti pada ekspresi SQL historis penutupan saldo cuti, sehingga keluaran khusus transfer dihasilkan pada salinan sementara.
