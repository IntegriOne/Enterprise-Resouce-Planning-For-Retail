# List operasi

| Operasi usulan | Jenis | Pengguna |
| --- | --- | --- |
| `reissueEmployeePayslip` | Mutation | Petugas HR dengan izin cetak ulang slip gaji |

## Penjelasan reissueEmployeePayslip

Setelah terminasi, akses portal dan sesi kerja ditutup oleh IAM. Mantan karyawan meminta slip melalui HR. Petugas memilih identitas pekerja di HRM dan slip yang diminta. HRM mencocokkan hasil payroll dengan pekerja tersebut dan memverifikasi bahwa Finance sudah menerbitkannya.

Jika versi PDF privat tersedia di DRM, gunakan versi yang sudah ada. Jika DRM mengonfirmasi bahwa PDF belum tersedia, render dari snapshot payroll terbit yang dibekukan, simpan PDF secara privat dan idempoten di DRM, lalu kembalikan versi yang dapat dicetak oleh petugas. Gangguan DRM atau penolakan akses menghentikan proses. Jika referensi versi sudah tersimpan tetapi DRM mengonfirmasi versi itu hilang, gunakan alur pemulihan yang menjaga jejak versi lama. Kontrak pemulihan perlu disepakati karena referensi versi yang sudah tersimpan bersifat immutable.

## Butuh GraphQL Mutation

Usulan `reissueEmployeePayslip(worker_id: ID!, payroll_result_id: ID!)`. ID pekerja memilih record HRM dan ID hasil menentukan satu slip ketika ada beberapa periode atau koreksi. UI menampilkan nomor dan nama pekerja, periode, serta aksi cetak. Pembentukan dan penyimpanan PDF merupakan perubahan state. Mutation tersedia di schema Gateway.

## RBAC

Usulan izin sensitif khusus petugas HR dengan MFA dan cakupan badan hukum pekerja. Scope SELF tetap untuk portal pekerja aktif. Audit minimal mencatat petugas, pekerja, hasil payroll, waktu, serta apakah versi dipakai kembali atau dibuat. Respons PDF mengikuti batas ukuran dan kebijakan tanpa cache yang berlaku untuk slip gaji.

## Kriteria penerimaan

- Mantan karyawan tidak dapat login setelah terminasi dan sesi lama ditolak.
- Petugas HR di luar cakupan badan hukum ditolak.
- PDF yang tersedia dapat dicetak tanpa membuat versi baru.
- PDF yang belum tersedia dibuat sekali dari snapshot terbit, termasuk saat permintaan diulang atau berjalan bersamaan.
- Hasil pekerja lain, payroll yang belum terbit, dan kegagalan DRM ditolak.

[Handoff IAM](28-izin-iam-untuk-wiring-gateway.md#bukti-iam-untuk-clearance-offboarding) · [Kontrak DRM](24-kebutuhan-integrasi-drm.md).

## Status implementasi

HRM sudah menolak ketiga jalur slip mandiri jika principal tidak memiliki hubungan kerja yang masih berjalan, termasuk saat sesi lama masih memiliki grant SELF. Jika pekerja aktif kembali dengan principal dan identitas orang yang sama, slip riwayat tetap dapat dibaca. Uji gRPC dengan PostgreSQL membuktikan penolakan untuk daftar, detail, dan PDF setelah seluruh hubungan kerja berakhir. Pencabutan sesi dan penolakan login di IAM serta mutation cetak ulang oleh HR masih perlu diimplementasikan.

### Jalur penutupan di IAM

HRM sudah mencatat event `hr_worker_terminations.recorded` dalam transaksi terminasi. Event membawa ID pekerja, rencana, operator dan tanggal efektif terminasi. Event lama dapat belum memuat tanggal, sehingga consumer perlu membaca keadaan sumber untuk rekonsiliasi. Untuk otomatisasi, IAM memerlukan pengiriman privat yang terautentikasi dan pembacaan keadaan hubungan kerja terkini dari HRM sebelum menonaktifkan akun. Pengiriman ulang atau event lama setelah rehire tidak boleh menonaktifkan akun dengan hubungan kerja yang masih berjalan.

IAM sudah memiliki `disableUser` untuk operator IAM berwenang dan service `Disable` yang mencabut sesi. Jalur otomasi perlu menggunakan semantik deaktivasi dan pencabutan sesi yang sama dengan otorisasi workload tersendiri. Regresi lifecycle IAM dengan race detector lulus. Pengiriman event dan konfirmasi penutupan kembali ke HRM masih terbuka. [Bukti HTTPS disableUser](../../e2e/gateway/2-iam-22-query-dan-45-mutation-operations-e2e.md).

Build development HRM sudah memuat pembatasan akses slip mandiri. Rollout mempertahankan overlay mandat IAM mTLS. Guard HTTPS lulus pada 52 Query dan 60 Mutation. Bukti skenario mantan pekerja berasal dari uji gRPC PostgreSQL hingga integrasi Finance, DRM, dan pencabutan sesi offboarding tersedia.

### Bukti yang dibutuhkan sebelum membuat ulang

Pemroses HRM yang ada dapat membuat PDF secara idempoten jika sudah memperoleh ID penerbitan Finance. Bila PDF belum pernah tertaut, hasil payroll juga belum menyimpan ID penerbitan tersebut. Jalur cetak ulang memerlukan cara menemukan penerbitan yang sah berdasarkan run, jurnal, badan hukum dan digest snapshot di Finance. Identitas penerbitan harus diperoleh dari service pemilik atau receipt yang sudah diverifikasi. Status `posted` sendiri hanya membuktikan tahap jurnal. [Catatan penerbitan Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md).

Adapter DRM untuk petugas HR perlu mengotorisasi peran petugas dan cakupan badan hukum serta mengikat versi PDF pada hasil payroll yang dipilih. Pembaca SELF saat ini memakai principal pemilik slip. Pisahkan otorisasi terhadap petugas dari verifikasi identitas pekerja yang menjadi subjek dokumen.

Outbox HRM menyimpan event saat mode `standby`. Jalur provisioning IAM memakai RPC khusus untuk pengajuan akun dan belum menerima event terminasi. Integrasi terminasi perlu delivery privat dan konfirmasi keberhasilan sebelum outbox mengakui pengiriman. Tautan principal, tanggal efektif, dan hubungan kerja aktif tetap diperiksa dari sumber HRM. Regresi outbox dan konfigurasi lulus dengan race detector.

Regresi service IAM kini secara eksplisit menjalankan deaktivasi akun employee, memeriksa pencabutan token sesi aktif, lalu mencoba login baru dengan kredensial yang benar. Login ditolak dan sesi baru tidak diterbitkan. Pembuktian ini mencakup perilaku `Disable` yang akan dipakai receiver offboarding.

Penerbit sesi IAM juga memeriksa status akun saat dipanggil langsung. Regresi membuktikan `Issue` menolak akun employee yang sudah disabled dan tidak mengembalikan sesi.

Uji HTTPS dengan akun employee hasil provisioning belum mencapai `disableUser`. Fixture mendapat respons sukses dari `activateUser`, lalu login awal ditolak dengan `invalid login credentials` pada percobaan email dan username. Pembersihan grant via GraphQL dan penghapusan stack disposable berhasil. Hasil ini tidak membuktikan adanya regresi deaktivasi. Verifikasi pada service IAM tetap lulus untuk pencabutan sesi dan penolakan login setelah deaktivasi.

### Verifikasi deaktivasi melalui HTTPS

Gagal login pada percobaan awal berasal dari dua pembuat akun pada receipt yang sama. Dengan satu pembuat fixture, akun employee berhasil login sebelum deaktivasi. Setelah `disableUser`, sesi lama ditolak dan login baru dengan kredensial yang sama gagal. [Uji HTTPS IAM](../../e2e/gateway/180-iam-mutation-disable-employee-provisioning-hrm.md).

Katalog IAM dan policy HRM sudah memuat izin sensitif `human_resource.payslips.reissue`. Policy mengharuskan MFA dan grant `LEGAL_ENTITY` yang cocok dengan badan hukum pemberi kerja pada hasil payroll. Grant GLOBAL saja ditolak. Command perlu melakukan pemeriksaan dengan atribut record sebelum memulai pembacaan atau pembuatan PDF. Izin belum dipasang pada role.

Command internal HRM `ReissueEmployeePayslip` sudah mencocokkan pekerja dan hasil, memeriksa scope badan hukum payroll, serta menggunakan pemroses penerbitan yang sama untuk membuat PDF bila belum tertaut. Command memverifikasi versi, digest, format dan ukuran dokumen sebelum mengembalikan byte, memeriksa ulang otorisasi, dan mencatat event audit. Uji PostgreSQL mencakup penggunaan ulang dokumen tanpa panggilan simpan kedua dan satu event audit. Adapter Finance/DRM dan pemulihan versi hilang masih terbuka.

RPC privat `PayrollService.ReissueEmployeePayslip` menerima Worker ID dan Payroll Result ID, lalu mengembalikan versi dokumen dan byte PDF. Tes registrasi server dan regresi gRPC PostgreSQL lulus. Inventory bertambah satu RPC. Integrasi provider Finance/DRM masih terbuka.

Gateway mengembalikan `payslip_pdf.content_base64` dengan batas satu PDF per operasi. Uji GraphQL dengan fixture gRPC lulus untuk petugas ber-MFA dengan grant badan hukum, penolakan grant GLOBAL atau MFA absen, dan penolakan dua PDF dalam satu operasi. Verifikasi jalur sukses dengan Finance dan DRM runtime masih terbuka.

Regresi gRPC juga mencakup PDF yang belum tertaut, respons penyimpanan hilang, dan retry yang mengembalikan versi awal. [Hasil uji](../../e2e/human_resource/61-payroll-mutation-cetak-ulang-pdf-grpc.md).

HRM mendukung pemulihan blob versi tertaut ketika DRM mengonfirmasi bahwa metadata versi masih ada. Adapter DRM harus membandingkan digest dan asosiasi hasil pada metadata immutable sebelum mengembalikan blob ke versi lama. Setelah pemulihan, HRM membaca kembali versi itu dan memvalidasi isinya. Verifikasi dengan DRM runtime masih terbuka. [Uji pemulihan gRPC](../../e2e/human_resource/61-payroll-mutation-cetak-ulang-pdf-grpc.md).

## Pemilihan slip oleh petugas HR

Workspace HR sudah memiliki halaman pekerja. Mutation cetak ulang memerlukan hasil payroll spesifik, sementara `payrollResults` saat ini memerlukan ID payroll run dan izin baca nominal. Untuk memilih slip dari detail pekerja, diperlukan Query daftar ringkasan per pekerja yang memeriksa grant khusus cetak ulang dan menjalankan filter badan hukum sebelum menghitung total atau memaginasi.

Query perlu menyediakan ID hasil untuk aksi cetak, kode dan tanggal periode untuk tampilan, serta penanda hasil koreksi jika ada. Nama dan nomor pekerja berasal dari detail pekerja yang telah dipilih. Filter harus memakai badan hukum pemberi kerja pada masing-masing hasil payroll agar perpindahan badan hukum tidak membuka slip historis di luar cakupan. Nominal, GID, dan referensi DRM tidak perlu ditampilkan pada daftar. Jika PDF belum ada, hasil yang sudah diterbitkan Finance tetap bisa dipilih.

Koneksi Relay mengikuti pola `nodes`, `edges`, `pageInfo`, dan `total_count`. Panggilan cetak tetap memeriksa ulang scope, MFA, dan validitas hasil.

### Korelasi event terminasi

Penerima perlu membedakan identitas operator yang tercatat pada event dari principal akun pekerja yang akan ditutup. Bukti HRM perlu mengikat ID pekerja, sumber lifecycle event berupa ID rencana, dan tanggal efektif. Untuk event lama, tanggal dapat diperoleh dari lifecycle event yang tersimpan. Sebelum deaktivasi, periksa tautan principal terkini dan seluruh hubungan kerja milik identitas yang sama. Jika tidak ada akun tertaut atau hubungan kerja yang lain masih berjalan, hasil pemrosesan perlu membedakan kedua kondisi tersebut dari kegagalan teknis.

IAM perlu menggunakan ID event untuk deduplikasi dan mengembalikan hasil yang dapat direkonsiliasi bila respons hilang setelah akun ditutup. Bukti penutupan untuk offboarding harus berasal dari IAM setelah status akun dan pencabutan sesi efektif. [Lifecycle terminasi HRM](../../../backend/services/human_resource/internal/service/personnel/termination.go) · [Deaktivasi IAM](../../../backend/services/iam/services/user/service.go).

### Batas transport verifikasi terminasi

Listener HRM bisnis saat ini mengandalkan proof sesi Gateway. Verifier status kerja untuk IAM memerlukan identitas workload IAM yang diverifikasi terpisah dari sesi browser. Pola listener mTLS terbatas untuk mandat requisition tersedia sebagai rujukan. Kontrak untuk IAM cukup mengembalikan keputusan penutupan, principal akun yang terikat, serta korelasi event, pekerja, rencana, dan tanggal. Detail profil, kontrak, serta slip tetap di HRM. Jalur internal ini tidak membutuhkan field GraphQL atau endpoint HTTP publik. [Interceptor HRM](../../../backend/services/human_resource/transport/grpc/interceptor/unary.go) · [Listener mTLS IAM](../../../backend/services/iam/cmd/requisition_mandate.go).

HRM kini memiliki `VerifyTerminationAccess` di service Personnel. Pemeriksaan mencocokkan lifecycle terminasi dengan pekerja, rencana, dan tanggal. Hasil membedakan pekerja tanpa akun tertaut, akun yang masih punya hubungan kerja, dan akun yang memenuhi syarat deaktivasi. Regresi PostgreSQL mencakup tanggal salah, pekerja tanpa akun, keputusan positif, dan event lama ketika ada record kerja aktif. Keputusan ini membaca keadaan HRM saat pemeriksaan. Transport mTLS dan koordinasi saat perubahan kerja bersamaan dengan deaktivasi masih perlu disambungkan.

Kontrak `TerminationAccessService.VerifyTerminationAccess` dan handler mTLS telah dibuat untuk listener workload terpisah. Handler menolak pemanggil tanpa sertifikat terverifikasi atau dengan URI di luar `spiffe://integrione/iam`. Tes handler mencakup penolakan sertifikat, workload lain, format tanggal, dan korelasi respons. Listener startup HRM dan klien IAM belum disambungkan. RPC workload ini berada di luar inventaris RPC bisnis dan tidak memiliki pemetaan GraphQL.

Startup HRM sekarang dapat membuka listener workload untuk `VerifyTerminationAccess` dengan alamat, sertifikat server, private key, dan CA klien tersendiri. Listener memerlukan mTLS dan hanya mendaftarkan service verifikasi terminasi. Tes registrasi memastikan service ini tidak terdaftar pada listener bisnis. Listener opsional belum diaktifkan pada Compose development, dan klien IAM serta uji lintas container masih terbuka.

Klien IAM untuk `VerifyTerminationAccess` sudah tersedia. Klien memakai mTLS dengan verifikasi hostname dan batas waktu, serta menolak respons dengan pekerja, rencana, atau tanggal yang berbeda. Keputusan positif tanpa principal valid ditolak. Tes klien IAM dan handler HRM lulus dengan race detector. Klien belum dipanggil oleh receiver event.

Use case IAM `ApplyEmployeeTermination` sudah memakai verifier HRM dan mekanisme deaktivasi akun employee. IAM memeriksa ulang keputusan sebelum mengubah akun. Akun yang masih memiliki hubungan kerja dipertahankan, perubahan bukti menghentikan deaktivasi, dan pengulangan pada akun disabled tidak mengubah status lagi. Tes service IAM membuktikan sesi aktif dicabut pada keputusan sah. Hasil service menyatakan perubahan akun tersimpan. Verifikasi publikasi pencabutan sesi, deduplikasi event, dan wiring receiver antarlayanan masih terbuka.
