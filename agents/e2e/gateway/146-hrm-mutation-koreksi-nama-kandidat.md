# Koreksi nama pilihan kandidat

Stack HTTPS terisolasi memakai IAM, Enterprise Foundation, HRM, Gateway, dan PostgreSQL. Superadmin memasang izin buat, baca, dan koreksi kandidat melalui GraphQL. Kandidat dibuat dengan individu dan kontak aktif dari Enterprise Foundation.

| Skenario | Hasil |
| --- | --- |
| Mengganti nama pilihan | Detail menampilkan nama baru |
| Mengirim nama `null` | Detail menampilkan nama kosong |
| Mengirim input kosong | Payload mengembalikan `ARGUMENT` |
| Mutation setelah role dicabut | Ditolak karena izin |
| Pemeriksaan PostgreSQL | Satu kandidat dengan individu dan kontak yang sama |

Runner selesai dengan status 0 dan menghapus database uji.
