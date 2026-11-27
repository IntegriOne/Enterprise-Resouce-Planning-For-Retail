# Verifikasi referensi pengajuan workforce

SubmitWorkforcePlan kini memeriksa ulang unit organisasi, lokasi dan badan hukum setiap baris melalui kontrak Foundation. Pemeriksaan menggunakan aturan yang sama dengan create/update baris dan berjalan sebelum status review serta outbox. Konflik referensi dibedakan dari dependency yang tidak tersedia.

Tes regresi gRPC ke PostgreSQL dengan Foundation tiruan menemukan pengajuan sebelumnya berhasil setelah referensi baris tidak valid. Setelah perbaikan, pengajuan ditolak, status tetap draft, dan pengajuan berhasil setelah referensi kembali valid. Tes modul dengan race detector dan go vet lulus.
