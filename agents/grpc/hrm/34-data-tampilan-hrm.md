# List operasi

## Penjelasan workers dan worker

Daftar dan detail pekerja memakai nomor pekerja, nama pekerja, nama badan hukum, status kerja, dan tanggal masuk sesuai konteks tampilan. HRM memberikan referensinya dan Enterprise Foundation memberikan labelnya. Jika nama belum tersedia, nomor pekerja tetap membantu operator mengenali record. Gunakan nama jabatan dan nama unit dari sumber pemilik ketika ditampilkan pada layar penempatan.

## Butuh GraphQL Query

- `workers` memuat daftar dan pencarian nomor pekerja. Satu permintaan `partiesByIds` dan satu permintaan `legalEntitiesByIds` mengambil label yang berbeda untuk satu halaman secara paralel, sesuai izin. Jika lookup gagal, daftar tetap tampil dan label yang hilang diberi keterangan.
- `myPayslips` dan `myPayslip` menyediakan kode dan rentang periode, tanggal terbit, tanggal bayar terjadwal, dan status pembayaran. Tampilkan tanggal terbit dan tanggal bayar dengan label terpisah. Aksi unduh dan cetak membutuhkan akses dokumen DRM privat.

## Butuh GraphQL Mutation

`requestEmployeeAccount` memakai ID pekerja internal. Tombol memakai label tindakan dan detail menampilkan status pengajuan serta waktu pengajuan.

## RBAC

Katalog izin HRM menentukan akses baca pekerja dan pengajuan akun. Query nama mengikuti izin Enterprise Foundation. Slip gaji memerlukan `human_resource.payslips.read_self` dan pemeriksaan kepemilikan, termasuk saat mengakses PDF privat dari DRM.

## Penyajian data

Simpan ID berbentuk `gid://` dalam variabel GraphQL dan state aplikasi. Untuk tabel, judul, filter, dan pesan kesalahan, gunakan nomor bisnis, nama, tanggal, atau label status. Gunakan pesan UI yang ditentukan per tindakan ketika permintaan gagal. Jika label belum tersedia, gunakan keterangan seperti “Belum tercatat”.

PDF cetak memakai kode periode bisnis beserta rentang tanggal, misalnya “Periode NOV-2026 (2026-11-01 hingga 2026-11-30)”. Tanggal terbit tetap dicetak pada baris terpisah.

Akses slip mandiri memerlukan jalur yang menerima izin slip milik sendiri. Guard halaman struktur perusahaan saat ini memerlukan izin operator HRM atau Enterprise Foundation. Izin baca struktur tidak perlu diberikan kepada pekerja hanya untuk membuka slipnya.

Rute `/account` sekarang memuat daftar slip melalui `myPayslips`. Rute ini memakai sesi login dan pemeriksaan akses mandiri pada Gateway. Daftar menampilkan status pembayaran dan tanggal berdasarkan metadata, serta mendukung paginasi. Tombol Siapkan PDF memuat konten melalui `myPayslipPdf`, lalu menyediakan tautan untuk mencetak dari penampil browser. Pengambilan dokumen riil menunggu adapter DRM produksi.

Akun baru memperoleh role dasar `iam_selfservice`. Role itu dipakai bersama oleh karyawan, vendor, dan supplier. Untuk mengaktifkan daftar slip bagi karyawan, administrator memberikan role yang memuat `human_resource.payslips.read_self` melalui `assignRolePermission`, lalu menugaskan role lewat `assignUserRole` dengan `scope_type: SELF` dan `scope_id` berupa ID akun karyawan tersebut. Gateway juga memeriksa jenis akun `employee` dan HRM memeriksa tautan pekerjanya. Gunakan alur yang sama bagi akun yang sudah ada.

Gateway memvalidasi tanggal kalender, urutan tanggal periode, status pembayaran, dan keunikan hasil per halaman. Respons upstream yang melanggar kontrak menghasilkan `UNAVAILABLE`, sehingga UI dapat menawarkan percobaan ulang tanpa menampilkan metadata yang tidak konsisten.

Portal menampilkan waktu terbit dengan jam dan zona UTC yang eksplisit. Tanggal bayar terjadwal dan rentang periode tetap berupa tanggal kalender.

## Daftar kandidat

Ruang kerja SDM menyediakan `/human-resource/candidates`. Daftar memakai `candidates` dengan pencarian nomor kandidat persis dan paginasi. Kolom memakai nomor, nama pilihan, dan label status. Tautan navigasi tampil untuk grant baca GLOBAL. Gateway tetap mengevaluasi akses ketika URL dibuka langsung.

Operator dengan grant ubah GLOBAL dapat mengoreksi nama pilihan kandidat aktif dari daftar. Nama kosong dikirim sebagai `null`. UI memakai nomor kandidat pada judul formulir dan pesan yang ditentukan per tindakan ketika penyimpanan gagal.

Pembuatan kandidat dari daftar memerlukan grant buat GLOBAL. Operator mengisi nomor, versi pemberitahuan privasi, waktu pengakuan, dan batas retensi. Dengan grant baca Enterprise Foundation, operator dapat mencari individu aktif dengan kode bisnis lalu memilih kontak aktif dari nama. Formulir mengirim referensi yang dipilih melalui variabel GraphQL dan tidak menampilkan ID tersebut sebagai label.

Aksi hapus kandidat aktif memerlukan grant hapus GLOBAL. Konfirmasi menyebut nomor kandidat. Jika Gateway meminta MFA, UI membuka verifikasi dan meminta operator mengonfirmasi kembali setelah sesi terverifikasi. Setelah berhasil, daftar dimuat ulang.

Field input HRM yang mengacu ke party, badan hukum, unit atau lokasi menerima Global ID bertipe sesuai dari pilihan Enterprise Foundation serta UUID. Label pilihan tetap memakai nama dan kode bisnis. [Uji HTTPS worker dan requisition](../../e2e/gateway/160-hrm-mutation-referensi-typed-id-foundation.md).
