# Uji HTTPS scope self dan prasyarat slip gaji

Dua akun fixture memakai stack disposable. Permission slip dipasang pada role lewat GraphQL IAM. Superadmin menerima assignment global, employee kedua menerima assignment `SELF` dengan `scope_id` yang sama dengan ID akunnya. Kedua hasil payroll sintetis dan tautan employee ke satu pekerja disiapkan melalui SQL fixture pada database HRM uji. Kedua hasil belum diterbitkan dan tidak memiliki versi dokumen.

| Kasus | Hasil |
| --- | --- |
| Assignment `SELF` tanpa `scope_id` | IAM menolak dengan `ARGUMENT` |
| `hrmMyPayslip` dengan grant global saja | `FORBIDDEN` |
| Employee memakai grant self untuk hasil sendiri yang belum terbit | `NOT_FOUND` |
| Employee meminta hasil pekerja lain | `NOT_FOUND` |
| `hrmMyPayslips` dengan grant self ketika verifier belum terpasang | `UNAVAILABLE` |
| PostgreSQL | Dua hasil sintetis dan nol referensi dokumen slip |

Database dan Redis disposable dihapus dan layanan development dipulihkan. Uji ini membuktikan perilaku akses sebelum publikasi. Akses PDF setelah penerbitan tetap memerlukan Finance dan DRM. [Handoff DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md).
