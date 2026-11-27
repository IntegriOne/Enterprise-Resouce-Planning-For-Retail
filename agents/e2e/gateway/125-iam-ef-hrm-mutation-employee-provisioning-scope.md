# Uji HTTPS createEmployeeUser dan validasi scope IAM → Enterprise Foundation

Basis Git `e28491c` dengan perubahan worktree. Gateway, IAM, Enterprise Foundation, dan HRM dibangun dari source aktif. Build IAM uji menambahkan katalog permission HRM fixture melalui Go overlay. Ketiga layanan bisnis memakai database PostgreSQL `*_test` dan Redis uji. Gateway memakai HRM gRPC mTLS serta HTTPS terverifikasi di port 18443.

| Kasus | Hasil |
| --- | --- |
| Login superadmin, cari permission, buat role, beri permission dan assignment | PASS lewat GraphQL IAM |
| `assignUserRole` dengan scope badan hukum acak | PASS, IAM menolak target yang tidak ditemukan di EF |
| `assignUserRole` dengan scope badan hukum aktif hasil `createLegalEntity` | PASS, IAM menerima target EF |
| `createParty` dan `hrmCreateWorker` | PASS, worker HRM dibuat sebagai preboarding menggunakan referensi EF |
| Enrollment, konfirmasi, dan login MFA superadmin | PASS lewat HTTPS |
| `createEmployeeUser` menerima Global ID worker | PASS, akun IAM berjenis `EMPLOYEE` dengan status `PENDING` |
| Pemeriksaan data HRM | PASS, `iam_principal_reference_id` sama dengan `_id` akun yang dikembalikan IAM |
| Pencabutan assignment dan permission, penghapusan role, logout | PASS lewat GraphQL IAM |

Worker fixture diubah menjadi aktif melalui SQL pada PostgreSQL HRM `*_test` setelah pembuatan preboarding. Aktivasi formal pekerja berada di luar cakupan uji. Penulisan dan cleanup role/grant memakai GraphQL IAM. Database dan container sementara dibersihkan setelah uji. Antrean pengajuan akun IAM serta deployment normal belum tercakup.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-employee.py`.

## Pengujian ulang kasus negatif

| Kasus | Hasil |
| --- | --- |
| Operator memanggil `createEmployeeUser` tanpa MFA | PASS, payload `AUTHORIZATION`. |
| Operator mencoba membuat akun kedua untuk worker yang sudah tertaut | PASS, payload `CONFLICT`. |
| PostgreSQL IAM setelah percobaan penolakan | PASS, hanya satu akun fixture tersimpan. |
