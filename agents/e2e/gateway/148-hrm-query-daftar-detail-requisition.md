# Daftar dan detail requisition melalui HTTPS

Stack terisolasi memakai Gateway, IAM, HRM, Enterprise Foundation, dan PostgreSQL. Superadmin memasang grant baca GLOBAL melalui GraphQL. Requisition draf disiapkan sebagai fixture PostgreSQL.

| Skenario | Hasil |
| --- | --- |
| Detail requisition | Nomor, status, kebutuhan, job, unit, dan kategori pekerja cocok |
| Daftar maupun filter nomor yang cocok | Satu hasil dengan total satu |
| Offset satu pada nomor yang cocok | Item kosong dengan total satu |
| Nomor yang tidak ada | Item kosong dengan total nol |
| Setelah permission baca dicabut | `FORBIDDEN` |

Runner selesai dengan status 0 dan menghapus database uji.
