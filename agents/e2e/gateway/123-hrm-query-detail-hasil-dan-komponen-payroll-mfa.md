# Uji HTTPS detail hasil payroll dan komponennya

Fixture payroll sintetis memakai database disposable. Review, persetujuan independen, dan pengajuan jurnal dijalankan melalui HTTPS. Kedua akun uji mendapat grant lewat mutasi GraphQL IAM dan memakai TOTP.

| Query | Hasil |
| --- | --- |
| `hrmPayrollResult` | ID hasil, run dan pekerja cocok. Mata uang IDR, gross dan net masing-masing 100. |
| `hrmPayrollResultLine` | ID hasil induk cocok. Komponen `BASE` bertipe `earning`, nilai 100 dan sequence 1. |
| `hrmPayrollResults` dan `hrmPayrollResultLines` | Tepat satu item pada filter yang sesuai. |
| PostgreSQL | Nilai gross, net dan komponen masing-masing `100.000000`. |

Nilai GraphQL dikonfirmasi sebagai string desimal. Database dan Redis uji dihapus setelah seluruh layanan development kembali sehat. [Uji alur pengajuan jurnal](108-hrm-query-mutation-review-approval-dan-pengajuan-payroll.md).
