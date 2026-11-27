# List operasi

| Pemilik | Operasi | Pemanggil |
| --- | --- | --- |
| HRM | `TerminateWorker` | Operator HR lewat jalur terminasi Personnel yang telah diotorisasi |
| IAM | `ApplyEmployeeTermination` | Relay outbox HRM melalui mTLS |
| HRM | `VerifyTerminationAccess` | Workload IAM melalui mTLS |

## Penjelasan ApplyEmployeeTermination

`TerminateWorker` mencatat terminasi dan event outbox dalam satu transaksi HRM. Relay mengirim worker, rencana offboarding, dan tanggal efektif. IAM membaca bukti terminasi dan hubungan kerja terkini dari HRM, lalu hanya menonaktifkan akun employee yang terhubung apabila bukti mengizinkan. Akun dengan hubungan kerja lain yang masih berlaku dipertahankan. Deaktivasi dan pencabutan sesi yang persisten menggunakan operasi IAM yang sama dengan jalur manual. Sinkronisasi sesi Redis mengikuti rekonsiliasi keamanan IAM.

IAM mengembalikan hasil deaktivasi, akun dipertahankan, atau akun belum terhubung. HRM mencocokkan worker, rencana, tanggal, dan hasil sebelum menandai event sebagai terkirim. Kegagalan komunikasi masuk antrean retry sesuai pengaturan outbox. Mode log/webhook HRM tidak mengambil event terminasi, termasuk saat integrasi IAM belum aktif.

## Butuh GraphQL

`ApplyEmployeeTermination` dan `VerifyTerminationAccess` adalah RPC privat antar-workload. Keduanya tidak memerlukan Query atau Mutation GraphQL tersendiri.

## RBAC

HRM tetap memeriksa grant dan scope terhadap operator pada operasi bisnis terminasi. Receiver IAM menerima sertifikat workload dengan identitas `spiffe://integrione/human-resource`. Verifier HRM menerima `spiffe://integrione/iam`. Principal akun employee diputuskan dari relasi HRM yang terverifikasi.

## Konfigurasi dan verifikasi

Untuk Docker development, jalankan dengan overlay [mandat IAM](../../../backend/deploy/docker-compose.hrm-mandate.yml) dan [terminasi HRM](../../../backend/deploy/docker-compose.hrm-termination.yml). `HRM_MANDATE_CERT_DIR` mengikuti konfigurasi mandat yang ada. `HRM_TERMINATION_CERT_DIR` berisi `hrm-server.crt`, `hrm-server.key`, `iam-client.crt`, `iam-client.key`, dan `ca.crt`. Sertifikat server HRM membutuhkan SAN DNS `human-resource`. Kunci privat dipasang read-only dan dapat dibaca UID service, 10001. Kedua listener privat; Compose tidak membuka port ke host.

Tes Go mencakup penolakan caller yang salah, korelasi hasil, retry, keputusan mempertahankan akun, dan idempotensi deaktivasi. Docker development membuktikan kedua listener mTLS hidup serta penolakan saat sertifikat klien atau bukti terminasi tidak tersedia. Uji HTTPS terminasi otomatis lulus, termasuk penolakan sesi lama dan login ulang. [Uji HTTPS](../../e2e/gateway/182-hrm-mutation-terminasi-pencabutan-sesi-iam.md). Bukti deaktivasi ini juga belum menjadi adapter clearance penutupan offboarding; kebutuhan Finance tetap di catatan handoff. [Clearance Finance](38-kebutuhan-finance-clearance-offboarding.md)

Regresi PostgreSQL disposable dengan race detector memeriksa empat konfigurasi relay. Mode umum hanya mengambil event umum. Mode terminasi hanya mengambil event untuk IAM. Mode gabungan mengambil keduanya, dan standby mempertahankan semua event. [Tes integrasi](../../../backend/services/human_resource/infrastructure/outbox/integration_test.go)

Saat relay terminasi aktif, konfigurasi memerlukan minimal dua koneksi PostgreSQL HRM. Relay menahan transaksi outbox selama IAM memanggil balik verifier HRM, yang membutuhkan koneksi kedua untuk membaca bukti. Event yang mencapai `HRM_OUTBOX_MAX_ATTEMPTS` tetap tersimpan dengan `published_at` kosong dan kode kegagalan terakhir. Batas percobaan mengikuti konfigurasi outbox.

## Penjelasan VerifyEmployeeTerminationClearance

RPC privat pada listener IAM yang sama memverifikasi worker, rencana, tanggal, dan principal dengan bukti HRM terkini. Jika akun harus dinonaktifkan, status persisten IAM wajib `disabled`. IAM dapat mengonfirmasi pekerja tanpa akun tertaut atau akun employee yang dipertahankan karena hubungan kerja masih berlaku. Pemeriksaan hanya membaca data. Jika proyektor Redis IAM aktif, clearance juga mewajibkan gate keamanan terbuka, recovery epoch sesuai, status Redis `disabled`, dan epoch autentikasi sudah mencapai revisi persisten. Proyeksi yang belum siap mengembalikan `UNAVAILABLE`. Client HRM mencocokkan seluruh identitas pada balasan. Penyelesaian rencana tetap memerlukan verifier Finance. [Handoff clearance](38-kebutuhan-finance-clearance-offboarding.md)

Saat batas percobaan tercapai atau digest payload rusak, relay menulis peringatan setelah transaksi outbox berhasil commit. Peringatan membawa ID event, jenis, dan kode kegagalan untuk rekonsiliasi. Isi event, ID pekerja di payload, dan pesan layanan upstream tidak dicatat. Event tetap berstatus belum terkirim.

## Prasyarat skema untuk TerminateWorker

Pembatalan penugasan yang dijadwalkan setelah hari kerja terakhir menggunakan `hr_worker_assignments.cancelled_at`. Penugasan yang sudah berjalan ditutup pada hari kerja terakhir. Riwayat keduanya tetap tersimpan. Perubahan ini membutuhkan baseline HRM `000001` yang sudah dikonsolidasikan beserta perubahan entity dan repository sebelum operasi bisnis terminasi dipublikasikan.

Runner memeriksa versi dan checksum riwayat migrasi. Database yang masih memakai migrasi HRM lama atau checksum baseline lama akan ditolak saat startup. Uji PostgreSQL disposable dengan race detector telah memverifikasi pemasangan ulang baseline, penolakan versi asing, dan penolakan checksum salah. Verifikasi ini tidak mengubah database development aktif. Untuk database berisi data yang harus dipertahankan, siapkan prosedur pemindahan data dan verifikasi sebelum rollout baseline.

Verifikasi SQL terpisah berhasil memasang baseline, menjalankan skenario integritas HRM dan talent, membatalkan baseline tanpa meninggalkan tabel HRM, lalu memasangnya kembali.

Konsolidasi skema yang setara sudah dipublikasikan pada commit `1325fe4`. Database baru mencatat satu migrasi. Database dengan tujuh checksum historis yang persis cocok tetap dapat berjalan dengan ledger lama. Pengecualian ini terikat pada checksum baseline yang telah dibuktikan setara melalui perbandingan `pg_dump --schema-only`. Baseline yang menambah perubahan bisnis, termasuk kolom `cancelled_at` milik penugasan, memerlukan verifikasi kompatibilitas tersendiri.

Commit `12fb870` menambahkan kompatibilitas kolom `cancelled_at` untuk dua riwayat yang diverifikasi, yaitu ledger tujuh versi historis yang lengkap dan baseline gabungan `1325fe4`. Runner menerapkan perubahan skema serta mencatat checksum baru dalam satu transaksi. Kedua jalur mempertahankan baris bisnis. Binary lama memerlukan checksum lama saat startup, sehingga setelah ledger diperbarui arah rollout adalah ke binary baru. Riwayat lain tetap ditolak.

Commit `8a51178` menambahkan `ApproveOffboardingPlan`, `StartOffboardingPlan`, dan `TerminateWorker` pada gRPC HRM. Persetujuan memerlukan principal berbeda dari pembuat rencana. Status persetujuan dan identitas reviewer berasal dari operasi khusus dengan grant terpisah. Pengujian PostgreSQL membuktikan alur positif, penolakan persetujuan sendiri, penolakan perubahan status lewat update umum, serta event terminasi yang tercatat satu kali. Wiring Gateway masih menunggu publikasi.
