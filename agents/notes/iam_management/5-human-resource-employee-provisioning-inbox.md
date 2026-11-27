# Antrean IAM untuk pengajuan akun dari HRM

HRM memakai `hr_access_requests` jenis `provision` sebagai sumber pengajuan dan menyimpan referensi pekerja serta event lifecycle. IAM perlu menerima pengajuan lewat kontrak privat dengan ID sumber unik, menyimpan status pemrosesan dan ID akun, serta mengembalikan keputusan dan hasil penautan melalui kontrak terverifikasi.

Gateway perlu memuat ID pekerja dari pengajuan yang dipilih di antrean IAM, lalu memvalidasi pasangan ID pengajuan dan pekerja sebelum memanggil `createEmployeeUser`. Pengajuan hanya bisa diklaim satu operator pada satu waktu. Percobaan ulang perlu melanjutkan pengajuan yang sama, termasuk jika akun sudah dibuat tetapi tautan HRM belum terkonfirmasi.

Lihat [alur pengajuan akun employee](../../grpc/hrm/33-alur-pengajuan-akun-employee.md) untuk pembagian tanggung jawab dan kriteria penerimaan.

## Status implementasi penyimpanan

Baseline IAM `000001`, DBML, dan repository sudah memuat `employee_provisioning_requests`. Penerimaan ulang dengan ID HRM dan isi identik mengembalikan ID penerimaan yang sama. Pekerja hanya dapat memiliki satu penerimaan terbuka. Klaim pending memakai update bersyarat dan hanya satu klaim konkuren yang berhasil.

Uji repository dengan PostgreSQL IAM `*_test` baru lulus untuk migrasi, penerimaan ulang, penolakan perubahan payload dengan ID yang sama, penolakan pengajuan kedua, dan klaim konkuren. Ingress gRPC, query antrean, penyimpanan ID akun dalam transaksi pembuatan user, dan penyampaian hasil ke HRM masih perlu disambungkan.

Ingress gRPC IAM dan pengiriman dari `hrmRequestEmployeeAccount` sudah lulus uji HTTPS. IAM memeriksa ulang snapshot izin pengaju serta badan hukum melalui EF. Setelah penerimaan, payload Mutation membawa ID IAM; status dan ID tersebut belum diproyeksikan ke penyimpanan HRM. Pengiriman ulang dari operator yang sama memerlukan izin baca pengajuan. [Uji pengiriman](../../e2e/gateway/127-hrm-iam-mutation-pengiriman-pengajuan-employee.md).

`createEmployeeUser` sekarang menerima `provisioning_request_id` opsional. Bila diberikan, IAM mencocokkan worker dan menyimpan klaim serta akun dalam transaksi yang sama. Operator yang sama dapat mengulang pemrosesan dan melanjutkan akun yang tersimpan. Gateway menandai penerimaan `linked` setelah tautan HRM terkonfirmasi. [Uji pemrosesan](../../e2e/gateway/129-iam-mutation-create-employee-dari-antrean.md).

Status pemrosesan kini tersedia bagi HRM melalui baca langsung ke IAM. `hrmEmployeeAccountRequest` memeriksa cakupan baca pengajuan dan pekerja pada Gateway, lalu IAM memeriksa lagi scope badan hukum sebelum mengembalikan status. IAM `pending/processing` tampil sebagai `submitted` dan IAM `linked` tampil sebagai `applied`. Kolom status lokal tidak digandakan selama pembacaan. [Uji status lintas layanan](../../e2e/gateway/130-hrm-query-status-provisioning-dari-iam.md).

Kontrak publik sekarang mewajibkan `provisioning_request_id` dan `worker_id`. Form IAM mengambil keduanya dari pengajuan yang dipilih. Tes HTTPS membuktikan penolakan tanpa ID pengajuan, tanpa MFA, dan untuk pasangan worker yang salah. [Uji kontrak wajib](../../e2e/gateway/131-iam-mutation-create-employee-wajib-pengajuan.md).

Katalog resmi memuat empat permission untuk `workers.read`, `workers.create`, `access_requests.read`, dan `access_requests.create` tanpa role default. IAM menyimpan nomor pekerja yang diterima dari HRM untuk membantu operator memilih pengajuan. [Uji end-to-end API memakai katalog resmi](../../e2e/gateway/132-iam-hrm-query-mutation-katalog-resmi-provisioning.md).
