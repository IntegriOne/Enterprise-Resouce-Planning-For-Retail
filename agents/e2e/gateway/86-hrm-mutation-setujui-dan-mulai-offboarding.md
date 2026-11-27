# Pengujian Mutation persetujuan dan mulai offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Grant approve dan MFA | approved |
| RPC menolak prasyarat persetujuan | CONFLICT |
| Grant start dan MFA | in_progress |
| Grant approve dipakai untuk start | FORBIDDEN |
| Sesi tanpa MFA | FORBIDDEN |

Uji memakai sesi employee, CSRF, dan fixture gRPC. Verifikasi guard HTTPS tercatat di bawah.

## HTTPS lokal

Kedua replika sehat setelah rollout image `sha256:13e0f4475b380330f1ea357de07a7d16ad22355c6cbd999163f98fd62a5efa66`. `hrmApproveOffboardingPlan` dan `hrmStartOffboardingPlan` mengembalikan `FORBIDDEN` dengan pesan token CSRF absen melalui Caddy HTTPS. Suite race dan vet Gateway lulus. Uji positif masih menunggu fixture IAM berizin.
