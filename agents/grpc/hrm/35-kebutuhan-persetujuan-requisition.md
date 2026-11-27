# List operasi

- `ApproveRequisition` memutuskan requisition yang sudah diajukan.
- `OpenRequisition` membuka requisition yang sudah disetujui setelah bukti diperiksa ulang.

## Penjelasan ApproveRequisition dan OpenRequisition

HRM memiliki transisi privat dan mencatat keputusan. Verifier runtime harus menyatukan mandat organisasi, kebutuhan tenaga kerja, dan otorisasi anggaran. Saat membuka requisition, HRM membandingkan referensi otorisasi terkini dengan referensi yang disimpan saat persetujuan.

## Butuh GraphQL Mutation

Gateway menyediakan `approveRequisition(id: ID!)` dan `openRequisition(id: ID!)`. Kedua mutation tetap membutuhkan verifier anggaran Finance untuk berhasil. Detail memakai nomor requisition, judul jabatan, dan nama unit. ID internal digunakan untuk mengirim operasi.

## RBAC

Aktor persetujuan harus manusia dan berbeda dari pembuat serta pengaju. Permission sensitif `human_resource.requisitions.approve` memerlukan MFA dan pemeriksaan mandat spesifik unit. Permission `human_resource.requisitions.open` mengatur pembukaan. IAM sudah mendukung kaitan grant dengan scope `ORGANIZATION_UNIT` serta memeriksa target scope ke Enterprise Foundation saat assignment.

## Enterprise Foundation, IAM, dan Finance

- Mandat peninjau berasal dari grant IAM `human_resource.requisitions.approve` pada scope `ORGANIZATION_UNIT` yang tepat. Keputusan owner 27 November 2026 memilih grant khusus unit. Jabatan kepala unit di Enterprise Foundation tidak memberikan mandat otomatis. Grant GLOBAL saja tidak memenuhi mandat ini.
- Enterprise Foundation memvalidasi unit aktif dan badan hukumnya. IAM memvalidasi unit saat assignment. Verifikasi runtime harus memakai grant efektif peninjau yang masih berlaku untuk unit record HRM, termasuk saat pembukaan diperiksa ulang dengan principal peninjau yang tersimpan. Identitas pelaku pembukaan dan izin `human_resource.requisitions.open` diperiksa terpisah.
- HRM perlu mengikat bukti terhadap snapshot requisition, job, unit, posisi atau lini rencana bila ada, kategori kerja, target tanggal, dan headcount.
- Finance perlu menyediakan referensi otorisasi anggaran stabil yang terikat pada requisition dan snapshot tersebut. Verifikasi perlu membedakan otorisasi aktif, dicabut, kedaluwarsa, atau tidak ditemukan. Kegagalan komunikasi harus mempertahankan status requisition. Integrasi Finance dicatat untuk sinkronisasi berikutnya.

[Uji referensi pekerja pada kepala unit](../../e2e/gateway/159-ef-mutation-kepala-unit-id-pekerja.md).

## Pemeriksaan mandat pada RPC privat

Policy HRM sekarang hanya menerima grant `ORGANIZATION_UNIT` untuk `ApproveRequisition`. Otorisasi awal memerlukan grant unit yang sah dan MFA sebelum transaksi. Setelah mengunci record, command mencocokkan grant dengan unit requisition yang tersimpan dan memeriksa ulang sebelum commit. Grant GLOBAL saja, grant unit lain, dan ketiadaan MFA ditolak. Pengujian policy dan service rekrutmen dengan race detector lulus.

Verifikasi ulang mandat peninjau dari IAM saat `OpenRequisition` sudah tersedia sebagai dependency terpisah. Verifier anggaran Finance masih memerlukan wiring. [Policy HRM](../../../backend/services/human_resource/infrastructure/security/policy.go) · [Transaksi command](../../../backend/services/human_resource/internal/service/operation/command.go).

Bukti yang diverifikasi saat approval dan pembukaan juga membawa lokasi kerja bila terisi. Pada pembukaan, hasil penolakan eksplisit verifier diteruskan sebagai akses ditolak. Gangguan dependency tetap menghasilkan `Unavailable`. Uji service menyimulasikan mandat yang dicabut dan memastikan status serta outbox tetap utuh.

## Riwayat implementasi verifier IAM

IAM perlu menyediakan verifikasi privat yang menjawab apakah principal peninjau yang tersimpan masih memiliki grant approval aktif pada unit tertentu. Jawaban cukup keputusan dan identitas unit serta principal yang diperiksa. `GetUserAccessSnapshot` yang sudah ada mengambil snapshot izin penuh dan tidak menyediakan kontrak verifikasi terbatas ini. Identitas pemanggil verifier harus berasal dari workload HRM terautentikasi. Implementasi dan uji lintas IAM masih terbuka.

Application service IAM kini memiliki `HasRequisitionApprovalMandate`. Fungsi membaca grant efektif dari repository IAM saat diperiksa, lalu mengembalikan keputusan untuk grant unit yang tepat pada akun employee aktif. Pengujian menolak GLOBAL, unit lain, grant yang hilang, dan akun nonaktif atau nonemployee. Entry point gRPC privat serta autentikasi workload masih perlu disambungkan. [Service IAM](../../../backend/services/iam/services/user/requisition_mandate.go).

Kontrak IAM privat `RequisitionMandateService.VerifyRequisitionMandate` menerima principal dan unit, lalu mengembalikan keputusan serta referensi korelasi. Handler mensyaratkan rantai sertifikat klien terverifikasi dengan URI persis `spiffe://integrione/human-resource`. Uji handler mencakup sertifikat tidak terverifikasi, workload lain, dan HRM. Registrasi listener mTLS terpisah, penerbitan sertifikat workload, dan klien HRM masih terbuka. [Proto IAM](../../../backend/services/iam/contract/proto/iam/v1/requisition_mandate.proto).

Startup IAM mendukung listener khusus dengan `IAM_REQUISITION_MANDATE_ADDR` dan berkas server certificate, private key, serta CA klien melalui `IAM_REQUISITION_MANDATE_CERT_FILE`, `IAM_REQUISITION_MANDATE_KEY_FILE`, dan `IAM_REQUISITION_MANDATE_CLIENT_CA_FILE`. Jika semua kosong, listener tidak dibuka. Listener hanya mendaftarkan layanan mandat, memerlukan TLS 1.3 dan sertifikat klien yang diverifikasi. Uji TCP gRPC memakai sertifikat sementara dan membuktikan permintaan tanpa sertifikat atau dengan URI workload lain ditolak. Klien HRM dan konfigurasi sertifikat deployment masih terbuka.

HRM kini menginisialisasi klien mTLS IAM saat target serta berkas TLS dikonfigurasi. Service memerlukan mandat IAM dan bukti anggaran sebagai dua dependency terpisah. Saat membuka requisition, HRM meminta mandat untuk principal peninjau yang tersimpan. Keputusan yang salah principal atau unit ditolak sebelum pemeriksaan anggaran. Konfigurasi tersedia pada `.env.example` HRM. Tes adapter dan regresi gRPC dengan PostgreSQL lulus. Wiring sertifikat deployment dan uji IAM lintas kontainer masih terbuka.

Katalog IAM sudah memuat izin sensitif `human_resource.requisitions.approve`. Pemberian grant unit tetap dilakukan melalui GraphQL. Pada stack uji, klien HRM menerima mandat hanya untuk unit yang tepat dan kembali ditolak setelah assignment dicabut. [Bukti lintas IAM–HRM](../../e2e/gateway/172-iam-mutation-grant-mandat-requisition-mtls.md).

## Pengoperasian development

Saat membuat ulang container IAM atau HRM, sertakan overlay mandat. Rollout dengan Compose dasar dapat menghapus konfigurasi listener dan klien mTLS. Overlay telah diterapkan kembali pada stack development. Kedua service sehat dan log IAM menunjukkan listener privat aktif pada port 9011. Guard HTTPS tetap lulus untuk 52 Query dan 60 Mutation.

Untuk mengaktifkan jalur development, buat sertifikat dan gunakan overlay Compose. Simpan berkas di lokasi lokal yang diabaikan Git. CA privat dihapus oleh generator setelah menerbitkan kedua sertifikat.

```sh
export HRM_MANDATE_CERT_DIR="$PWD/backend/services/human_resource/.local/iam-mandate"
backend/deploy/scripts/gen-hrm-mandate-dev-certs.sh "$HRM_MANDATE_CERT_DIR"
docker compose -f docker-compose.yml -f backend/deploy/docker-compose.hrm-mandate.yml up -d --build --wait iam human-resource
```

Pemeriksaan mandat berjalan sebelum pengecekan ketersediaan verifier anggaran. Uji HTTPS dengan IAM mTLS membuktikan `openRequisition` memakai peninjau tercatat walaupun operator pembukaan berbeda. Pemberian grant memungkinkan alur mencapai pemeriksaan Finance, dan pencabutan grant langsung memblokir pembukaan tanpa perubahan status. [Uji HTTPS pembukaan dan pencabutan mandat](../../e2e/gateway/174-hrm-mutation-open-requisition-mandat-dicabut.md).

Pembukaan kini memberikan UUID anggaran yang tersimpan kepada verifier agar Finance dapat memeriksa otorisasi yang sama secara langsung. [Kontrak dan kriteria uji Finance](37-kebutuhan-finance-anggaran-requisition.md).

Kedua mutation juga lulus jalur positif HTTPS dengan IAM mTLS, EF, PostgreSQL, dan verifier anggaran fixture yang hanya dipasang pada binary uji. HRM menyimpan UUID anggaran yang sama pada dua transisi dan menulis masing-masing satu event outbox. [Uji jalur positif dengan fixture anggaran](../../e2e/gateway/175-hrm-mutation-approve-open-requisition-fixture-anggaran.md).

### Pemeriksaan mTLS setelah rollout

Sertifikat development berlaku 30 hari dan generator membutuhkan direktori tujuan baru. Saat rotasi, arahkan `HRM_MANDATE_CERT_DIR` ke direktori baru lalu buat ulang kedua service dengan overlay yang sama agar keduanya memakai CA yang cocok.

```sh
openssl verify -CAfile "$HRM_MANDATE_CERT_DIR/ca.crt" -purpose sslserver -verify_hostname iam "$HRM_MANDATE_CERT_DIR/iam-mandate.crt"
openssl verify -CAfile "$HRM_MANDATE_CERT_DIR/ca.crt" -purpose sslclient "$HRM_MANDATE_CERT_DIR/hrm-mandate.crt"
openssl x509 -in "$HRM_MANDATE_CERT_DIR/hrm-mandate.crt" -noout -ext subjectAltName -enddate
```

Sertifikat client harus memuat URI `spiffe://integrione/human-resource`. Port mandat IAM 9011 hanya tersedia pada jaringan privat container. Pemeriksaan sertifikat development saat ini lulus, port 9011 tidak dipublikasikan ke host, dan [regresi HTTPS](../../e2e/gateway/174-hrm-mutation-open-requisition-mandat-dicabut.md) membuktikan pencabutan grant unit.
