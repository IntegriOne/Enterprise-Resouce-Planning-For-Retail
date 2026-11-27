# Uji HTTPS slip gaji dengan grant SELF dan global

Stack disposable menggunakan build IAM dan Gateway dari checkout. Katalog uji menambahkan izin slip. Superadmin membuat role dan memberikan permission serta assignment melalui GraphQL. Employee dibuat melalui pengajuan HRM dan `createEmployeeUser`, sehingga principal sudah tertaut ke pekerja. Verifikasi identitas login memakai fixture di database IAM terisolasi. Seluruh database uji dihapus setelah selesai.

| Skenario | Hasil |
| --- | --- |
| Pemberian assignment SELF kepada akun dengan ID scope milik principal lain | IAM menolak penugasan. |
| Superadmin dengan grant global saja memanggil `hrmMyPayslips` | Gateway mengembalikan `FORBIDDEN`. |
| Employee tertaut dengan grant global dan SELF yang cocok memanggil `hrmMyPayslips` | Otorisasi lulus. HRM mengembalikan `UNAVAILABLE` karena verifier belum terpasang. |

Runner lokal `/tmp/erp-cross-service.UUJx1n/run-payslip-scope.py` selesai dengan status 0. Hasil ini mencakup enforcement scope sampai HRM. Penerbitan dan akses PDF memerlukan integrasi Finance dan DRM.

Regresi diulang dengan build IAM yang memakai manifest HRM dari checkout. Hasil penolakan global dan penerimaan grant SELF eksplisit tetap sama.
