# Verifikasi HTTPS Query dan Mutation cuti

Kedua replika Gateway lokal dibangun ulang dan sehat. Tes Caddy HTTPS memakai CA development terverifikasi.

| Skenario | Hasil |
| --- | --- |
| Detail dan daftar kebijakan tanpa sesi | UNAUTHENTICATED |
| Detail dan daftar hak cuti tanpa sesi | UNAUTHENTICATED |
| Daftar ledger dan koreksi tanpa sesi | UNAUTHENTICATED |
| Buat, setujui dan tolak koreksi tanpa CSRF | FORBIDDEN |
| Pembukuan koreksi dan penutupan hak cuti tanpa CSRF | FORBIDDEN |

Seluruh 11 skenario lulus dalam `TestHRMLeaveGuardsOverHTTPS` dengan race detector. Dua belas kasus payroll dan dua kasus filter pekerja juga lulus pada build yang sama. Uji positif dengan grant resmi masih terbuka.
