# Kompatibilitas rollout gRPC HRM

Image checkout berhasil dibangun tetapi startup pada DB development menolak `unsupported HRM migration version`. DB mencatat tujuh migrasi. Checkout menggunakan baseline gabungan `000001`. Ketujuh checksum DB cocok dengan file dari commit `e979089c646939572c05fefbcc7ccc1f5f270843`.

Image sebelumnya tidak tersedia melalui Docker setelah container diganti. Layanan dipulihkan dengan membangun commit tersebut dari arsip Git di luar worktree dan menjalankan Compose kembali. Healthcheck gRPC lulus, kedua Gateway sehat, dan tracker migrasi tetap memakai tujuh versi.

Panggilan mTLS ke `GetOffboardingTaskByID` tanpa sesi mencapai guard `Unauthenticated`. `CancelOffboardingPlan` mengembalikan `Unimplemented` pada runtime pemulihan. Schema Gateway terkini mengenali mutation tersebut, tetapi alur berakun memakai fixture sampai backend terbaru bisa dijalankan dengan DB yang kompatibel. Image checkout tersimpan dengan tag lokal `integrione-retail-human-resource:checkout-verification`.

## Startup image checkout pada database sementara

Image `sha256:d9347627289789e49fd9aca729ee3b56045c984d95ccfe36def57e9e8c5f9f9d` berhasil start dengan database kosong pada PostgreSQL Docker. Healthcheck gRPC lulus dan tracker mencatat baseline versi 1. Dua panggilan mTLS tanpa sesi, `GetOffboardingTaskByID` dan `CancelOffboardingPlan`, mencapai guard dan mengembalikan `Unauthenticated`. Container dan database sementara dihapus. Layanan development tetap memakai image pemulihan yang cocok dengan tujuh migrasi yang sudah tercatat.

## Pengalihan database development

Pemeriksaan baca-saja menemukan nol baris pada 86 tabel bisnis `hr_*` di database lama. Development dialihkan melalui konfigurasi lokal `HRM_DB_NAME` ke `human_resource_baseline_dev`. Database lama `human_resource` tetap tersedia dengan tujuh migrasi. Image pemulihan disimpan sebagai `integrione-retail-human-resource:legacy-seven`.

Image checkout sehat pada baseline versi 1. RPC detail tugas dan pembatalan mencapai guard sesi pada listener development melalui mTLS. Kedua replika Gateway sehat. Pemeriksaan Caddy HTTPS lulus untuk 42 Query dan 47 Mutation tanpa sesi atau token CSRF. Jalur berakun masih memerlukan grant IAM resmi. Pengalihan ini berlaku untuk database development tanpa baris bisnis. Upgrade database lama yang berisi data masih perlu migrasi yang teruji.

## Registrasi seluruh kontrak pada listener development

Pemeriksaan mTLS memakai deklarasi proto lokal untuk memanggil seluruh 1.011 RPC di 22 service pada listener development. Setiap panggilan menggunakan payload kosong tanpa sesi. Seluruh RPC mengembalikan `Unauthenticated`, termasuk 18 RPC yang digolongkan nonaktif pada inventaris. Hasil membuktikan method terdaftar dan mencapai guard sesi. Perilaku bisnis setelah sesi valid tetap mengikuti status dan tes masing-masing operasi. [Inventaris operasi](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md).
