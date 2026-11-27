# Pengujian integrasi Query status run payroll

Gateway menerima sesi employee fixture dan memakai Payroll gRPC melalui bufconn.

| Kasus | Hasil |
| --- | --- |
| Detail run diajukan dengan izin global | Status approved, waktu pengajuan terisi, jurnal dan settlement kosong |
| Daftar run diposting dengan izin global | Status posted, referensi jurnal, dan total dari HRM |
| Akses daftar dengan scope badan hukum | FORBIDDEN |

Ketiga kasus lulus dalam `TestHRMPayrollRunsThroughGateway` dengan race detector. Uji positif HTTPS dengan IAM resmi dan Finance masih terbuka.
