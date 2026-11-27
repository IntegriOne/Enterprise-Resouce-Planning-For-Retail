# Pengujian integrasi review dan persetujuan payroll

| Kasus GraphQL dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Ajukan review dengan izin submit_review, MFA dan CSRF | Status review |
| Setujui dengan izin approve, MFA dan CSRF | Status approved |
| Setujui dengan izin submit_review saja | FORBIDDEN |

Semua kasus lulus dengan race detector. HRM memeriksa pemisahan penginisiasi dan penyetuju melalui workflow gRPC. Uji positif HTTPS dengan grant IAM resmi masih terbuka.
