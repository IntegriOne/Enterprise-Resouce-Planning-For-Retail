# Aktivasi akun setelah penautan HRM

Stack disposable memakai IAM dari checkout dan simulasi respons penautan pada Gateway seperti pada [uji pemulihan](137-iam-mutation-pemulihan-penautan-akun.md). Grant superadmin dipasang melalui GraphQL.

| Skenario | Hasil |
| --- | --- |
| HRM sudah menyimpan tautan tetapi Gateway belum dapat mengonfirmasinya ke IAM | Mutation provisioning mengembalikan `UNAVAILABLE` dan penerimaan IAM tetap `processing`. |
| Operator memanggil `activateUser` saat penerimaan masih `processing` | IAM menolak dengan `CONFLICT`. |
| Operator mengulang provisioning, kemudian memanggil `activateUser` | IAM memakai akun yang sama, menandai penerimaan `linked`, dan mengizinkan aktivasi akun. |
| PostgreSQL dan Query HRM | Satu akun, satu pengajuan, dan status HRM `applied`. |

Jalur verifikasi identitas menggunakan pemeriksaan domain IAM yang sama sebelum aktivasi. Skenario HTTPS ini menguji mutation aktivasi manual.

## Rollout development

Image IAM terbaru dibangun dan container lulus healthcheck. Uji HTTPS development memakai superadmin dan grant sementara melalui GraphQL. Login, empat bacaan daftar pekerja, pencabutan grant dan logout berhasil. Simulasi penerimaan `processing` tetap dijalankan di stack disposable.
