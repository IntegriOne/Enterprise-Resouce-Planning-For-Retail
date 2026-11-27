# Uji HTTPS penolakan requisition dengan dua aktor

Uji berjalan melalui Caddy dan `/api/graphql` pada stack terisolasi. Superadmin membuat role, menambahkan permission, dan memberi role kepada kedua akun melalui GraphQL IAM. Akun kedua berasal dari provisioning IAM–HRM. Verifikasi alamat akun uji memakai fixture database IAM terisolasi.

| Kasus | Hasil |
| --- | --- |
| Pembuat draf mencoba menolak pengajuannya sendiri dengan MFA | Ditolak dengan `AUTHORIZATION` |
| Peninjau mencoba menolak tanpa MFA | Ditolak dengan `AUTHORIZATION` |
| Peninjau dengan MFA memberi alasan kosong | Ditolak |
| Peninjau dengan MFA memberikan kode alasan | Status menjadi `rejected` dan payload memuat kode alasan |
| Query detail setelah penolakan | Status dan kode alasan sama dengan payload |
| Penolakan ulang | Ditolak dengan `CONFLICT` |
| Pemeriksaan database tes | Timestamp terisi dan principal peninjau berbeda dari pembuat |
