# Pengujian Query posisi

| Kasus Workforce Planning gRPC fixture | Hasil |
| --- | --- |
| Detail dengan izin global baca posisi | Kode posisi dan headcount desimal 1.25 |
| Daftar dengan izin baca posisi | Satu posisi dan total 1 |
| Izin baca jabatan saja | FORBIDDEN |

Uji positif HTTPS dengan grant resmi masih terbuka.

| Kasus filter badan hukum | Hasil |
| --- | --- |
| ID badan hukum valid | Posisi yang cocok dan total dari HRM |
| ID badan hukum tidak valid | INVALID_ARGUMENT |
| RPC mengembalikan posisi dari badan hukum lain | UNAVAILABLE |

Izin global baca posisi tetap diperlukan untuk daftar terfilter.
