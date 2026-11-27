# Pengujian Mutation assignment primer dan pengaitan principal

Suite `TestHRMWorkerAssignmentAndPrincipalThroughGateway` menggunakan sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Assignment primer valid | active |
| Assignment tidak cocok dengan pekerja | CONFLICT |
| Bukti pengaitan diterima RPC | active |
| Verifier pengaitan tidak tersedia | UNAVAILABLE |
| Format ID bukti salah | INVALID_ARGUMENT |
| Grant assignment dipakai untuk pengaitan | FORBIDDEN |
| Pengaitan tanpa MFA | FORBIDDEN |

Pengaitan runtime masih memerlukan verifier bukti provisioning. Uji HTTPS dengan grant IAM resmi masih terbuka.
