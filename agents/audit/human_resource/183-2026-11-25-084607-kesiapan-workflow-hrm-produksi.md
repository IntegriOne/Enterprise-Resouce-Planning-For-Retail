# Kesiapan workflow HRM produksi

| Cakupan | Bukti tersedia | Pekerjaan tersisa |
| --- | --- | --- |
| Kontrak dan operasi Gateway | Pemeriksa repo mencocokkan 1.010 RPC di 22 service, 669 kandidat Query, 323 kandidat Mutation dan 18 CRUD nonaktif. Tes policy menolak operasi nonaktif walaupun fixture berisi grant historis. | Wiring resolver dan grant IAM resmi dilakukan setelah persetujuan katalog. |
| Runtime lokal | Regresi 68 paket dengan PostgreSQL dan race detector serta startup mTLS membuktikan baseline pada laporan terakhir. Perubahan payroll dan pengaitan akun setelah laporan itu memiliki regresi terarah. | Uji lintas Gateway–IAM–HRM memakai grant resmi. |
| Kompensasi | RPC CRUD dapat mencatat proposal terbatas, dengan referensi assignment, pay component, periode dan band. | Persetujuan kompensasi memerlukan provenance pengusul dan pemisahan pengusul dari penyetuju. Kontrak saat ini belum menyediakan command persetujuan. |
| Learning dan Career | Learning dapat mencatat mulai, selesai administratif dan withdrawal. Career menyelesaikan coaching atau mentoring yang dikonfirmasi mentor. | Penyelesaian tindakan course membutuhkan referensi enrollment yang dapat ditelusuri. Nilai kelulusan dan kredensial memerlukan sumber bukti terpisah. |
| Payroll dan slip pribadi | Pengajuan jurnal, renderer PDF, handler event privat dan pemeriksaan kepemilikan slip telah diuji dengan fixture. | Finance perlu menyediakan fakta jurnal dan penerbitan. DRM perlu menyediakan penyimpanan dan verifikasi akses PDF. Gateway perlu menghubungkan Query dan akses cetak. |
| Pengaitan pekerja dan akun | HRM menolak tautan dengan riwayat orang berbeda dan mencatat lifecycle event untuk tautan yang diterima. | IAM perlu memberikan bukti provisioning yang mengikat principal, pekerja dan party orang. |

[Inventaris Gateway](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md) · [Regresi runtime](182-2026-11-25-083520-regresi-dan-startup-hrm-terkini.md) · [Handoff Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [Handoff IAM](../../notes/iam_management/1-human_resource-permission-handoff.md)
