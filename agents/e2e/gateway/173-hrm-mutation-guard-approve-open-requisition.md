# Uji HTTPS guard persetujuan dan pembukaan requisition

Stack disposable memakai binary Gateway, IAM, HRM, dan Enterprise Foundation dari checkout lokal. Superadmin masuk dengan MFA dan memperoleh role serta grant melalui GraphQL. Dua unit dibuat melalui Enterprise Foundation. Requisition submitted dan approved disiapkan sebagai fixture di PostgreSQL HRM disposable.

| Kasus | Hasil |
| --- | --- |
| `approveRequisition` dengan grant GLOBAL saja | `AUTHORIZATION`. |
| Approval dengan grant untuk unit berbeda | `AUTHORIZATION`. |
| Approval dengan grant untuk unit yang tepat | `UNAVAILABLE` karena verifier runtime belum lengkap. |
| `openRequisition` pada record approved dengan grant GLOBAL pembukaan | `UNAVAILABLE` karena verifier runtime belum lengkap. |
| Baca ulang setelah semua permintaan | Status masing-masing tetap submitted dan approved. |

Race test adapter Gateway dan integrasi gRPC–PostgreSQL juga lulus. Uji HTTPS ini membuktikan guard dan status yang dipertahankan. Uji mandat IAM mTLS dilakukan terpisah. Jalur transisi positif masih membutuhkan verifier anggaran Finance. [Uji mandat IAM](172-iam-mutation-grant-mandat-requisition-mtls.md).
