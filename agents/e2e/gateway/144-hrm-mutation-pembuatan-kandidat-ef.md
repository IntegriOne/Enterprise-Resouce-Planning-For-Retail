# Pembuatan kandidat dengan kontak Enterprise Foundation

Stack HTTPS terisolasi menjalankan Gateway, IAM, Enterprise Foundation, HRM, Redis, dan PostgreSQL. Superadmin membuat role, memasang izin, dan menugaskan role ke akun melalui GraphQL. Individu dan kontak dibuat melalui GraphQL Enterprise Foundation.

| Kasus | Hasil |
| --- | --- |
| Kontak individu lain | `CONFLICT` |
| Kontak tanpa individu | `CONFLICT` |
| Individu aktif dengan kontak miliknya | Kandidat aktif tersimpan |
| Nomor kandidat duplikat | `CONFLICT` |
| Permintaan setelah assignment dicabut | Ditolak karena izin |
| Pemeriksaan PostgreSQL | Satu kandidat dengan referensi yang cocok |

Runner berhasil dan menghapus database uji.

IAM dan kedua replika Gateway development sudah memakai build terbaru dan sehat. Uji baca pekerja menggunakan superadmin juga lulus setelah rollout.
