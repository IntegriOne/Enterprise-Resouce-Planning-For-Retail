# Pengujian integrasi pembatalan payroll

| Kasus GraphQL dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin cancel, MFA, CSRF, dan alasan dengan spasi tepi | RPC menerima alasan yang sudah dirapikan. Respons memuat status cancelled dan waktu pembatalan |
| Alasan hanya berisi spasi | INVALID_ARGUMENT |
| Izin approve tanpa izin cancel | FORBIDDEN |

Ketiga kasus lulus dengan race detector. HRM memeriksa bahwa run sedang dalam review dan pembatal berbeda dari penginisiasi. Verifikasi positif melalui HTTPS dengan grant IAM resmi masih terbuka.
