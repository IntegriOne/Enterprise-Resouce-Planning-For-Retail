# Uji HTTPS transisi persetujuan dan pembukaan requisition

Stack disposable menjalankan Gateway, IAM, Enterprise Foundation, dan HRM. Superadmin masuk dengan MFA dan memasang role serta grant melalui GraphQL IAM. HRM memeriksa mandat melalui listener mTLS IAM. Dua unit dibuat melalui GraphQL Enterprise Foundation. Record awal submitted dan approved disiapkan di PostgreSQL disposable.

Binary HRM uji memasang verifier anggaran fixture saat build terisolasi. Fixture memeriksa korelasi snapshot dan mengembalikan satu UUID anggaran yang stabil. Konfigurasi deployment tidak berubah.

| Kasus | Hasil |
| --- | --- |
| Approval dengan grant global saja atau grant unit lain | Ditolak. |
| Approval dengan grant unit yang cocok dan bukti anggaran fixture | Status berubah dari submitted menjadi approved. UUID anggaran fixture tersimpan. |
| Pembukaan dengan mandat aktif dan UUID anggaran yang sama | Status berubah menjadi open. |
| Ulangi pembukaan record yang sudah open | Konflik. |
| Buka record milik peninjau lain yang grant-nya dicabut | Ditolak dan status tetap approved. |
| Periksa outbox | Record yang berhasil ditransisikan memiliki tepat satu event approve dan satu event open. |

Assignment dan role uji dicabut melalui GraphQL. Kontainer dan database disposable dibersihkan. Bukti ini mencakup jalur positif HRM terhadap fixture verifier. Jalur dengan layanan Finance masih perlu diuji.
