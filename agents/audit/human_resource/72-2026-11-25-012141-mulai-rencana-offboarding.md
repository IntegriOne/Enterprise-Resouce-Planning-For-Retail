# Mulai rencana offboarding

RPC privat StartOffboardingPlan menambahkan izin start dengan step-up. Transisi approved ke in_progress mempertahankan metadata persetujuan dan berlaku pada hari UTC antara tanggal pemberitahuan dan tanggal efektif, termasuk kedua batas. Validator mengunci dan memeriksa pekerja, mengecek persetujuan independen dan mensyaratkan setidaknya satu tugas. Outbox dicatat dalam transaksi yang sama.

Tes gRPC ke PostgreSQL sementara lulus untuk penolakan mulai sebelum approval, perubahan status setelah approval, penolakan pengulangan, perlindungan tugas setelah mulai, dan jumlah event outbox. Policy IAM menggunakan stub. Penyelesaian atau waiver tugas dan penutupan belum tersedia.
