# List operasi

`createEmployeeUser`, `approveOffboardingPlan`, `startOffboardingPlan`, `terminateWorker`, operasi payroll dan slip, serta operasi HRM lain yang membutuhkan kontrak server terbaru.

## Penjelasan pemeriksaan build Gateway

Kandidat memakai snapshot `main` untuk backend dan salinan Gateway lokal. Build mengidentifikasi kontrak dan modul yang harus dipublikasikan sebelum Gateway. Kontrak provisioning IAM dan penandatanganan bukti penautan HRM telah dipublikasikan. Runtime provisioning IAM masih perlu menerima permintaan, memvalidasi pekerja, serta menyelesaikan aktivasi setelah penautan HRM.

Kontrak lain yang masih dibutuhkan kandidat meliputi `CheckUsernameAvailability` IAM, `preserve_head_person` Enterprise Foundation, dan operasi HRM untuk kontrak bertanda tangan, katalog jabatan, penyesuaian cuti, offboarding lanjutan, payroll, slip, mutasi penugasan, dan workforce plan. Paket katalog RBAC bersama juga belum ada pada snapshot yang diuji. Pemeriksaan berhenti pada tahap kompilasi sehingga hasilnya belum membuktikan perilaku runtime.

## Butuh GraphQL

Operasi domain mengikuti Query atau Mutation pada katalog masing-masing. Pemeriksaan ini belum mempublikasikan SDL atau menjalankan E2E HTTPS dari snapshot `main`.

## RBAC

Izin `human_resource.payslips.read_self` kini mensyaratkan grant `self` serta snapshot milik principal yang sama. Grant global, snapshot tanpa principal, dan permintaan untuk principal lain ditolak.

Runtime Enterprise Foundation, pembaruan unit organisasi dengan `preserve_head_person`, migrasi, dan pengaturan client IAM sudah dipublikasikan di `db79aea`. Kandidat dari main lulus 88 tes dengan race detector pada PostgreSQL disposable, buf lint, vet, build Go, serta build image. Koneksi dari Gateway dan E2E HTTPS lintas domain masih perlu diverifikasi.

IAM `CheckUsernameAvailability` dan katalog deklaratif `shared/rbac` sudah tersedia di `origin/main` pada `ffe8f7d` dan `87956c3`. Kandidat Gateway berhasil dikompilasi setelah kedua commit tersebut dipadukan dengan runtime HRM lokal di snapshot terisolasi. Regresi HRM dan Gateway serta prosedur migrasi baseline HRM masih diverifikasi sebelum publikasi.

Runtime HRM yang dibutuhkan Gateway sudah tersedia pada `34ee90e`. Tes PostgreSQL disposable merekam 301 tes utama lulus. Jalur upgrade dari empat riwayat yang dikenali mempertahankan data contoh dan menghasilkan kolom, constraint, indeks, fungsi, dan trigger setara dengan pemasangan baru. Uji konflik data membuktikan perubahan skema dan ledger dibatalkan bersama. Gateway dan E2E HTTPS dari main masih dalam verifikasi. Integrasi Finance dan DRM tetap bergantung pada layanan pemilik.

Gateway sudah dipublikasikan pada `cefcf5f`. Regenerasi SDL, vet, tes race Gateway, dan build image lulus. Stack HTTPS disposable dari commit tersebut menolak akses anonim ke antrean provisioning, lalu pengujian berhenti ketika login superadmin mengembalikan `AUTHENTICATION`. Main masih memakai konfigurasi bootstrap lama, sedangkan fixture memuat konfigurasi superadmin terbaru. Sinkronisasi katalog startup dan pemberian role melalui GraphQL perlu diverifikasi setelah bootstrap diselaraskan. Database aktif tidak digunakan untuk skenario ini.

Bootstrap IAM yang baru tersedia pada `08fdec6`. Pengujian HTTPS dari arsip main berhasil login dan MFA sebagai superadmin, mengelola role melalui GraphQL, menautkan employee secara otomatis, serta mencabut sesi setelah terminasi. Uji guard juga lulus untuk 52 Query dan 61 Mutation. Verifikasi slip dengan provider Finance dan DRM masih perlu dilakukan. [Laporan E2E](../../e2e/gateway/183-hrm-mutation-provisioning-dan-terminasi-main.md)
