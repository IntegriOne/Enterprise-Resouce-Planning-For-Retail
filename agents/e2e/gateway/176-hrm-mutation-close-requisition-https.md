# Uji HTTPS penutupan requisition

Superadmin masuk dengan MFA melalui HTTPS dan memperoleh role serta permission melalui GraphQL IAM. Stack disposable menjalankan Gateway, HRM, IAM, dan Enterprise Foundation. Requisition awal disiapkan pada database HRM disposable. Approval dan pembukaan memakai verifikasi mandat IAM mTLS dan bukti anggaran fixture.

| Kasus | Hasil |
| --- | --- |
| Tutup requisition yang masih submitted | `CONFLICT`. |
| Tutup setelah approval dan pembukaan | `closeRequisition` berhasil. Status tersimpan menjadi closed. |
| Ulangi penutupan | `CONFLICT`. |
| Periksa outbox | Tepat satu event untuk masing-masing approval, pembukaan, dan penutupan. |

Regresi gRPC dengan PostgreSQL juga lulus, termasuk penolakan HRM ketika posting atau lamaran masih aktif. Race test Gateway, katalog RBAC, sinkronisasi IAM, dan go vet Gateway lulus.
