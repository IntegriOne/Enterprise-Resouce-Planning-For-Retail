# Pengujian integrasi Mutation pengajuan jurnal payroll

Endpoint GraphQL Gateway memakai sesi employee dan Payroll gRPC fixture.

| Kasus | Hasil |
| --- | --- |
| Run disetujui dengan izin global, MFA, dan CSRF | Tanda terima dengan waktu permintaan dan status `approved` |
| Run tidak memenuhi syarat di HRM | `CONFLICT` |
| Sesi tanpa izin pengajuan | `FORBIDDEN` |
| Sesi tanpa MFA | `FORBIDDEN` |
| Mutasi tanpa token CSRF | `FORBIDDEN` |

Kelima kasus lulus dalam `TestHRMRequestPayrollPostingThroughGateway` dengan race detector. Uji ini memakai fixture. Status publikasi, jurnal, pembayaran, dan rekonsiliasi Finance berada di luar cakupan pengujian.
