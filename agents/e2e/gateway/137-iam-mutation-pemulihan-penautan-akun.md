# Pemulihan penautan akun yang belum dapat dikonfirmasi

Uji memakai stack HTTPS terisolasi dan pengaturan role superadmin melalui GraphQL seperti pada [uji aktivasi](135-hrm-mutation-aktivasi-dan-provisioning-akun.md). Gateway uji dibangun dengan overlay yang mengembalikan `UNAVAILABLE` satu kali setelah HRM menyimpan tautan, sebelum Gateway memberi tanda selesai ke IAM. Overlay hanya dipakai pada build uji.

| Skenario | Hasil |
| --- | --- |
| Percobaan pertama `createEmployeeUser` | GraphQL mengembalikan `UNAVAILABLE`. Pengajuan IAM tetap `processing`. PostgreSQL HRM sudah menyimpan principal. |
| Operator yang sama mengulang mutation | Gateway menggunakan akun yang sudah tersimpan dan mengonfirmasi tautan. |
| Percobaan ulang berikutnya | IAM kembali memberikan ID akun yang sama. |
| Status lintas service dan pemeriksaan PostgreSQL | IAM `linked`, HRM `applied`, satu akun dan satu penerimaan. |

Percobaan pertama menjalankan stack dengan target hostname internal dan sempat menerima `UNAVAILABLE` pada login. Percobaan ulang memakai IP container sebagai target gRPC pada lingkungan uji. Runner lokal `/tmp/erp-cross-service.UUJx1n/run-activation-fault.py` selesai dengan exit code 0.
