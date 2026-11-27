# Pengujian Query hak cuti pekerja

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Detail dengan izin global baca hak cuti | Saldo snapshot desimal 12.50 |
| Daftar menurut ID pekerja | Item sesuai dan total 1 |
| Izin baca kebijakan saja | FORBIDDEN |
| Respons gRPC memuat hak cuti pekerja lain | UNAVAILABLE |

Tes integrasi menggunakan race detector. Verifikasi dengan grant resmi melalui HTTPS masih terbuka.
