# Pengujian Query jabatan

| Kasus Workforce Planning gRPC fixture | Hasil |
| --- | --- |
| Detail dengan izin baca jabatan | Kode jabatan dan level opsional kosong |
| Daftar dengan izin baca jabatan | Satu jabatan dan total 1 |
| Izin baca rencana saja | FORBIDDEN |

Akses positif dengan grant resmi melalui HTTPS masih terbuka.

| Kasus filter kelompok | Hasil |
| --- | --- |
| ID kelompok valid | Jabatan dari kelompok tersebut dan total 1 |
| Format ID kelompok tidak valid | INVALID_ARGUMENT |
| RPC mengembalikan jabatan kelompok lain | UNAVAILABLE |
