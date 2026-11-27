# gRPC persetujuan sampai terminasi

Commit `8a51178` terverifikasi di `origin/main`. gRPC mendukung persetujuan offboarding oleh reviewer independen, mulai rencana, dan terminasi beserta lifecycle serta outbox IAM. Update umum tidak dapat mengganti status atau menyuntikkan bukti persetujuan.

Tes PostgreSQL disposable lulus untuk alur gRPC, verifikasi grant pada katalog, penolakan persetujuan sendiri dan update status langsung, proteksi terminasi sebelum rencana berjalan, serta event IAM. Buf lint, tes race seluruh modul, vet, dan build lulus. Kandidat menggunakan snapshot main dan pengujian ini tidak menggunakan database aktif. Wiring Gateway dan E2E HTTPS dari main masih perlu diverifikasi.
