# Penyelesaian tugas offboarding

RPC CompleteOffboardingTask memakai izin khusus dan step-up. Konfirmasi harus divalidasi melalui port yang menerima ID bukti, tugas, rencana, pemilik, dan actor. Hanya tugas pending milik HRM pada rencana in_progress yang didukung. Status, waktu server, referensi konfirmasi, dan outbox dicatat dalam satu transaksi.

Integrasi gRPC ke PostgreSQL sementara lulus untuk penolakan konfirmasi yang tidak cocok dan penyelesaian dengan konfirmasi yang diterima verifier tiruan. Bootstrap produksi belum menyediakan verifier otoritatif. RPC mengembalikan Unavailable selama port tersebut belum terhubung. Clearance dan penutupan rencana tetap terbuka.
