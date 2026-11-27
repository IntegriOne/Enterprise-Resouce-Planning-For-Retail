# Regresi lintas domain workforce dan rekrutmen

Integrasi menjalankan handler Workforce Planning dan Recruitment pada gRPC yang sama dengan PostgreSQL sementara. Setelah SupersedeWorkforcePlan mengganti rencana, SubmitRequisition menolak draf yang masih merujuk versi lama. Pembacaan ulang membuktikan status tetap draf dan outbox hanya berisi event pembuatan. Requisition yang sudah dibuka tetap melewati alur posting pada tes. Versi pengganti memakai fixture SQL. Policy IAM, bukti Finance, dan verifikasi DRM menggunakan stub.

## Cakupan dan pekerjaan tersisa

Pemeriksaan nama method sumber mencocokkan 982 deklarasi RPC dengan handler pada 22 domain. Dukungan perilaku dan kesiapan produksi memerlukan bukti terpisah. Perbandingan matriks rancangan dengan katalog runtime menghasilkan 36 kode command yang masih belum dipetakan.

| Domain | Command tersisa pada matriks |
| --- | --- |
| Personnel | Link principal, suspend, terminate, transfer assignment, approval/aktivasi/akhir kontrak, verifikasi dan penolakan dokumen. |
| Recruitment | Approval, penerbitan, penarikan, penerimaan dan penolakan offer. |
| Compensation | Submit, approve, reject dan activate. |
| Payroll | Calculate, submit review dan request posting. |
| Learning | Penyelesaian enrollment dan verifikasi kualifikasi. |
| Succession | Approval rencana, nominate dan withdraw. |
| Employee Relations | Resolusi kasus. |
| Policy | Publikasi kebijakan. |
| HR Service | Assign, resolve dan close. |
| Engagement | Publish dan close. |
| Access Request | Submit dan recommend. |
| Talent | Penutupan sesi kalibrasi. |

Kode runtime juga belum membuktikan bahwa semua adapter produksi tersedia. Clearance Offboarding, anggaran Workforce Planning dan Recruitment, serta verifikasi bukti DRM masih memerlukan wiring dan pengujian owner. Grant IAM nyata dan jalur gRPC mTLS belum dibuktikan untuk seluruh workflow.

Suite integrasi serial dengan race detector selesai dengan exit 0, mencakup 67 paket yang lulus pada PostgreSQL disposable. Database tes telah dibersihkan.
