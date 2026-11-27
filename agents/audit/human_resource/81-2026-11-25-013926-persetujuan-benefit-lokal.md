# Persetujuan benefit lokal

ApproveBenefitEnrollment memakai izin step-up untuk actor manusia. Dalam transaksi, validator memeriksa pekerja aktif, masa berlaku enrollment dan plan, serta plan lokal tanpa aturan kontribusi atau provider. Approval menyimpan status approved dan verified_at dari jam server. Actor tercatat dalam outbox. Nilai kontribusi, referensi provider, dan status aktivasi provider tidak diisi oleh command.

Integrasi gRPC ke PostgreSQL mencakup approval melalui RPC khusus, penolakan approval ulang, perlindungan dari delete dan jumlah event outbox. Tes policy masih menggunakan stub. Pengakhiran enrollment dan aktivasi atau kontribusi provider masih terbuka.
