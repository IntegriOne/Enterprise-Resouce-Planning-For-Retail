# List operasi

Alur HRM → IAM mencakup pengajuan dari detail pekerja, antrean IAM, pembuatan akun, dan pembacaan hasil dari HRM. Gateway meneruskan setiap operasi ke service pemilik.

## Penjelasan pengajuan akun

Petugas HR memilih pekerja di halaman HRM dan mengajukan akun saat pekerja memenuhi syarat onboarding. HRM mengambil `worker_id` dari record pekerja yang dipilih, memvalidasi lifecycle dan scope petugas, lalu menyimpan permintaan `provision` di `hr_access_requests`. Permintaan mengacu pada pekerja dan event lifecycle yang menjadi dasarnya. Pengiriman ke IAM menggunakan ID pengajuan yang sama pada setiap percobaan.

## Penjelasan antrean IAM

IAM menerima pengajuan melalui kontrak internal resmi dan menyimpan penerimaan agar operator bisa melihatnya di antrean provisioning. Antrean menampilkan nama atau nomor pekerja, badan hukum, pengaju, waktu, dan status. Operator memilih pengajuan tersebut. Sistem mengambil `worker_id` dari pengajuan yang sudah diverifikasi server. Notifikasi dapat mengarahkan operator ke antrean, sementara status pengajuan tetap tersedia di sana.

## Penjelasan createEmployeeUser

Pertahankan `worker_id: ID!` dalam kontrak yang sudah disepakati dan tambahkan referensi pengajuan untuk jalur ini. UI mengisi `worker_id` dari baris antrean. Saat menerima Mutation, server memuat pengajuan berdasarkan ID dan memeriksa bahwa worker dalam input sama dengan worker pada pengajuan, status masih dapat diproses, dan scope operator mencakup pengajuan. Pemrosesan mengklaim pengajuan secara atomik untuk mencegah dua operator membuat dua akun. Percobaan ulang memakai ID pengajuan yang sama dan melanjutkan langkah yang belum selesai.

IAM membuat akun `PENDING`, lalu HRM memverifikasi bukti provisioning melalui gRPC privat dan menyimpan tautan ke worker. IAM menandai pengajuan `linked` setelah tautan dikonfirmasi. Query HRM membaca hasil dari IAM dan memproyeksikannya sebagai `applied`. Bila IAM tidak dapat memeriksa status, Query HRM tetap menampilkan pengajuan lokal dengan status `verification_unavailable`. Petugas dapat memperbarui status. Penolakan izin tetap dikembalikan sebagai error. IAM menahan aktivasi manual dan aktivasi melalui verifikasi identitas selama penerimaan masih diproses. Setelah status `linked`, akun mengikuti verifikasi identitas dan kebijakan aktivasi IAM. [Uji aktivasi](../../e2e/gateway/139-iam-mutation-aktivasi-akun-setelah-penautan-hrm.md). Jika hasil RPC belum pasti, pengajuan tetap dapat direkonsiliasi menggunakan ID pengajuan dan referensi akun yang tersimpan.

## Butuh GraphQL Query/Mutation

| Pemilik | Jenis | Kebutuhan |
| --- | --- | --- |
| HRM | Mutation | `requestEmployeeAccount` menyimpan pengajuan. |
| HRM | Query | `employeeAccountRequest` membaca status berdasarkan ID pengajuan. |
| HRM | Query | `workerEmployeeAccountRequest(worker_id: ID!)` membaca pengajuan dari halaman pekerja. |
| IAM | Query | `employeeProvisioningRequests` dan `employeeProvisioningRequest`. |
| IAM | Mutation | `createEmployeeUser` memerlukan `worker_id` dan `provisioning_request_id`. |

## RBAC

HRM menentukan siapa yang boleh mengajukan untuk pekerja dalam scope-nya. IAM menentukan siapa yang boleh membaca antrean dan memproses akun dengan izin provisioning serta MFA. Pemberian role dan permission mengikuti kebijakan IAM. HRM menyimpan relasi sebagai bagian dari data pekerja dan menerima status pemrosesan pengajuan.

## Status dan kriteria penerimaan

Uji HTTPS membuktikan pengajuan dari browser, penolakan saat MFA belum terpenuhi atau pekerja pada input berbeda, penerbitan satu akun saat mutation diulang, dan pembacaan status akhir di HRM. [Uji lintas service](../../e2e/gateway/134-hrm-mutation-pengajuan-katalog-iam-otomatis.md) · [Uji browser](../../frontend/e2e/human_resource/1-mutation-pengajuan-akun-pekerja.md).

Uji HTTPS juga membuktikan command `activateWorker` mencatat peristiwa perekrutan yang diterima saat pengajuan akun. Pengujian memakai fixture kontrak bertanda tangan dan katalog izin aktivasi khusus uji. [Uji aktivasi dan provisioning](../../e2e/gateway/135-hrm-mutation-aktivasi-dan-provisioning-akun.md). Dua permintaan HTTPS bersamaan dari satu sesi MFA dengan username dan email berbeda menghasilkan satu akun. Klaim serentak oleh dua principal MFA berbeda menghasilkan satu akun dan satu penolakan `CONFLICT`. [Uji dua operator](../../e2e/gateway/142-iam-mutation-provisioning-dua-operator.md). Pengujian dengan simulasi respons penautan yang tidak dapat dikonfirmasi membuktikan operator yang sama dapat mengulang dan menyelesaikan pengajuan dengan akun yang sudah tersimpan. [Uji pemulihan](../../e2e/gateway/137-iam-mutation-pemulihan-penautan-akun.md). [Uji permintaan bersamaan](../../e2e/gateway/136-iam-mutation-create-employee-serentak.md).

Dialog IAM menampilkan pesan lokal ketika pemuatan antrean atau pengiriman akun employee gagal. Untuk konflik, operator diminta memeriksa status pengajuan serta email dan nama pengguna yang dipakai. Pesan sumber yang dapat memuat ID teknis tetap berada di lapisan API.

Runner repositori juga menguji pengulangan `createEmployeeUser` dan refetch `EmployeeAccountRequest` via Relay. [Laporan HTTPS](../../e2e/gateway/162-iam-mutation-replay-provisioning-dan-node.md).

Respons privat IAM untuk penerimaan dan status mengembalikan ID pengajuan HRM, ID pekerja, dan ID badan hukum. Gateway mencocokkan ketiganya dengan record HRM sebelum memakai status IAM.

## Pemulihan pengajuan yang masih diproses

Klaim dan asosiasi akun tersimpan di IAM. Pengajuan `processing` dengan akun tersimpan dapat dilanjutkan oleh principal pengklaim menggunakan receipt yang sama. Jika pengklaim tidak lagi berwenang atau tidak tersedia, pemindahan klaim memerlukan keputusan IAM tersendiri. Pemeriksaan saat ini tetap menolak operator berbeda dan menjaga akun yang belum tertaut dalam status pending.

Regresi HTTPS pada build checkout aktif membuktikan satu akun dan satu receipt setelah pengulangan, serta status akhir yang cocok di IAM dan HRM. [Hasil regresi](../../e2e/gateway/178-iam-mutation-regresi-provisioning-employee-https.md).

Setelah respons penautan tidak pasti, Gateway memeriksa ulang record pekerja dengan tenggat tiga detik. Pemeriksaan ini dapat berjalan saat permintaan browser sudah dibatalkan. Jika konfirmasi belum didapat, status pengajuan tetap dapat dipulihkan melalui retry dengan receipt yang sama.

Regresi adapter Gateway dengan race detector membatalkan context permintaan saat respons link gagal. Pembacaan ulang masih dapat mengonfirmasi tautan yang sudah tersimpan. Jika pembacaan ulang tertahan, tenggat tiga detik berakhir dan adapter mengembalikan `ErrLinkUnconfirmed`. [Tes Gateway](../../../backend/gateway/internal/adapter/human_resource/worker_provisioning_test.go)

Jika penautan serentak menghasilkan konflik, Gateway membaca ulang pekerja dengan tenggat yang sama. Permintaan dapat dilanjutkan hanya jika HRM mengonfirmasi principal akun yang persis sama. Tautan ke akun lain tetap menghasilkan `CONFLICT`. Regresi adapter menguji kedua hasil tersebut.

## Penyimpanan dan konfirmasi penautan

IAM menyimpan penerimaan, klaim operator, akun hasil provisioning, dan status penautan melalui commit `c4442a2`. Pengiriman dengan ID pengajuan yang sama hanya dianggap idempoten jika identitas, scope, dan rincian penerimaan cocok. Satu pekerja hanya dapat memiliki satu pengajuan terbuka. Baseline IAM `000001` memasang antrean pada database baru dan melengkapi baseline IAM yang sudah terpasang tanpa mereset data. Uji ini memakai PostgreSQL disposable.

Penyelesaian provisioning perlu memverifikasi bahwa HRM mengaitkan worker dan akun yang sama dengan isi pengajuan. Periksa jalur aktivasi admin dan verifikasi email karena keduanya dapat mengaktifkan akun. Wiring service dan uji HTTPS masih perlu diselesaikan dari `main`.

### Verifikasi akun sebelum penyelesaian provisioning

`VerifyWorkerAccountLink` adalah RPC baca privat tanpa Query atau Mutation GraphQL tambahan. IAM meminta konfirmasi pekerja, principal akun, dan badan hukum dari HRM. HRM memeriksa bahwa pekerja aktif, belum diterminasi, serta memiliki principal dan badan hukum yang cocok. Listener mTLS hanya menerima sertifikat beridentitas `spiffe://integrione/iam`.

Server HRM dan client IAM tersedia di `origin/main` pada commit `ad45203` dan `d6d36eb`. Wiring lokal pada `CompleteEmployeeProvisioningRequest` mempertahankan status pengajuan ketika verifier hilang, HRM tidak tersedia, atau tautan tidak cocok. Uji race dan vet lulus untuk service lokal. Runtime provisioning lengkap dan E2E HTTPS dari `main` masih belum tersedia.

Runtime IAM sudah dipublikasikan pada `3827b06`. Server mendaftarkan layanan provisioning, mengklaim pengajuan dan menyimpan akun pending dalam satu transaksi, serta memeriksa badan hukum ke Enterprise Foundation. Penyelesaian pengajuan memerlukan konfirmasi mTLS dari HRM. Aktivasi admin dan verifikasi email menolak akun provisioning yang belum tertaut. Konfigurasi Enterprise Foundation menggunakan `IAM_EF_GRPC_ENABLED`, `IAM_EF_GRPC_TARGET`, dan `IAM_EF_GRPC_TIMEOUT_SECONDS`. Konfirmasi akun menggunakan konfigurasi mTLS listener HRM yang juga dipakai untuk akses terminasi. Tes race IAM, vet, build Go, serta build image kandidat lulus. Penulisan tautan dari HRM, publikasi Gateway, dan uji HTTPS dari main masih dalam pekerjaan.

Commit `a337d67` mempublikasikan `LinkWorkerPrincipal` sebagai RPC privat. Bukti mengikat pekerja, party, akun employee, dan operator IAM. Database membatasi satu hubungan kerja yang masih berlaku per principal, sambil mempertahankan riwayat pekerja yang telah berakhir. Update pekerja generik ditolak jika mencoba mengganti principal. PostgreSQL disposable membuktikan bukti sah, penolakan bukti palsu, penolakan update generik, dan pembacaan kembali tautan yang tersimpan. Operasi ini tidak menambah mutation publik untuk HR.
