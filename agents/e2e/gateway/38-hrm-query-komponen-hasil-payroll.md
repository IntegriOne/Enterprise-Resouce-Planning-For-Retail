# Pengujian Query komponen hasil payroll

| Kasus Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin global baca komponen dan MFA | Kode komponen, nominal presisi, dan rate opsional kosong |
| Izin baca ringkasan hasil saja | FORBIDDEN |
| Izin baca komponen tanpa MFA | FORBIDDEN |
| Induk hasil berbeda dari permintaan | UNAVAILABLE |

Keempat kasus lulus dengan race detector. Uji positif melalui HTTPS dengan akun resmi masih terbuka.
