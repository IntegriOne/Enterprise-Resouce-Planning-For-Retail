# Uji koneksi IAM, Enterprise Foundation, dan HRM pada stack development

Basis Git `e28491c` dengan perubahan worktree. Build Docker Gateway dan HRM berhasil setelah daftar dependency HRM menambahkan `github.com/google/uuid` yang telah digunakan oleh bukti provisioning. Rollout memakai `HRM_GATEWAY_ENABLED=true docker compose up -d --no-deps --no-build --wait` untuk HRM, Gateway 1, lalu Gateway 2. Kedua replika memiliki `HRM_GRPC_ENABLED=true` dan status healthy. IAM dan Enterprise Foundation development juga healthy.

| Kasus HTTPS di `erp.localhost:8443` | Hasil |
| --- | --- |
| `userLogin` sebagai superadmin | PASS |
| Query `legalEntities` | PASS |
| Pembuatan role, pemberian permission Payroll Rule Sets dan assignment ke akun uji | PASS lewat GraphQL IAM |
| Query `hrmPayrollRuleSets` empat kali setelah login ulang | PASS. Log HRM mencatat empat RPC `GetAllPayrollRuleSets`. |
| Pencabutan role dan permission, hapus role, dan logout | PASS lewat GraphQL IAM |

Pengujian normal memakai katalog IAM development yang tersedia dan operasi baca. Uji tulis lintas service dan skenario penolakan dijalankan dengan database dan Redis sementara. [Workforce Planning](124-iam-ef-hrm-query-mutation-workforce-cross-service.md) · [Provisioning akun employee](125-iam-ef-hrm-mutation-employee-provisioning-scope.md).

Flag aktivasi diberikan pada perintah Compose. Default Compose masih `false`, sehingga pengulangan rollout memerlukan flag tersebut. Katalog resmi HRM pada source dan workflow antrean pengajuan IAM berada di luar hasil uji konektivitas ini.
