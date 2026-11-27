# Pengujian integrasi pembuatan draf payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Input sah dengan grant create, MFA dan CSRF | Status draft. RPC menerima digest SHA-256 biner dan kunci idempotensi yang sama |
| Duplikasi dari HRM | CONFLICT |
| Digest bukan SHA-256 heksadesimal | INVALID_ARGUMENT |

Ketiga kasus lulus dengan race detector. Pengujian positif dengan grant IAM resmi melalui HTTPS masih terbuka.
