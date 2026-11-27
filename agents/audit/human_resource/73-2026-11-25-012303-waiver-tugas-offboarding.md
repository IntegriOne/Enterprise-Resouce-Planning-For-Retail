# Waiver tugas offboarding

RPC privat WaiveOffboardingTask memakai izin khusus dan step-up. Pelaku manusia dapat memberi waiver beralasan atas tugas pending milik HRM pada rencana in_progress. Validator mengunci rencana dan pekerja, serta mempertahankan seluruh atribut tugas di luar status, alasan dan pelaku waiver. Perubahan dan outbox dicatat secara atomik. Jalur ini mencakup tugas lokal. Konfirmasi penyelesaian oleh owner dan penutupan plan masih terbuka.

Integrasi gRPC ke PostgreSQL sementara membuktikan alasan kosong ditolak, waiver tercatat atas actor yang sah, waiver ulang ditolak, dan hanya mutasi yang berhasil menambah satu event outbox. Policy IAM memakai stub.
