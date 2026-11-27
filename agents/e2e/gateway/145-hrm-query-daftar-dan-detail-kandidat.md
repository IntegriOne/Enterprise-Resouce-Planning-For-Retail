# Daftar dan detail kandidat

Gateway, IAM, Enterprise Foundation, HRM, dan PostgreSQL diuji melalui HTTPS pada stack terisolasi. Superadmin memasang izin baca dan buat lewat GraphQL. Kandidat dibuat lewat GraphQL dengan individu dan kontak aktif dari Enterprise Foundation.

| Kasus | Hasil |
| --- | --- |
| Detail dengan ID kandidat hasil mutasi | Nomor, nama pilihan, dan status cocok |
| Daftar tanpa filter dan dengan nomor persis | Satu baris dan total satu |
| Nomor yang sama pada offset satu | Daftar kosong dan total tetap satu |
| Nomor yang tidak ditemukan | Daftar kosong dan total nol |
| Daftar setelah role dicabut | `FORBIDDEN` |

Runner selesai dengan status 0. Data dan stack uji dibersihkan.
