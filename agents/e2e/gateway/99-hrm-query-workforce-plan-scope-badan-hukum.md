# Uji HTTPS scope badan hukum pada workforce plan

Superadmin membuat dua badan hukum aktif melalui API Foundation dan masing-masing satu draf workforce plan melalui Gateway. Kedua service memakai database disposable. Setelah setup, assignment global dan izin membuat draf dicabut. Role uji yang hanya memiliki `human_resource.workforce_plans.read` diberikan lagi lewat IAM untuk badan hukum A. Sesi diperbarui dengan login ulang.

| Kasus | Hasil |
| --- | --- |
| `hrmWorkforcePlan` dengan ID badan hukum A | Detail memuat ID dan badan hukum yang sesuai |
| Detail rencana badan hukum B | `NOT_FOUND` |
| `hrmWorkforcePlans` tanpa offset | Satu rencana yang diizinkan dan total 1 |
| Daftar dengan limit 1 dan offset 1 | Item kosong dan total tetap 1 |
| Pemeriksaan PostgreSQL | Dua badan hukum berbeda dan dua rencana tersimpan |

Assignment, role, dan sesi uji dibersihkan. Kedua service dipulihkan ke database development dan kembali sehat. Database disposable dihapus. [Pengujian perubahan draf](97-hrm-mutation-referensi-badan-hukum-workforce.md).

## Regresi rencana dan rincian dengan katalog resmi

Superadmin membuat dua badan hukum aktif dan dua unit organisasi melalui EF, lalu membuat satu rencana dan satu rincian per badan hukum. Job merupakan fixture dalam database HRM disposable. Gateway memakai payload GraphQL terkini dan IAM memuat manifest resmi HRM.

Assignment global dan izin pembuatan dicabut melalui GraphQL sebelum role baca ditugaskan dengan scope badan hukum A. Setelah login ulang, detail rencana dan rincian A terbaca. Detail B mengembalikan `NOT_FOUND`. Daftar global mengembalikan hanya rincian A dengan total satu. Filter memakai ID rencana B mengembalikan daftar kosong dan total nol. Pengujian keluar dengan status 0, dan stack disposable dibersihkan.

Regresi lulus kembali di stack HTTPS terisolasi dengan image IAM, EF, HRM, dan Gateway yang aktif di Compose development. Runner memakai binary dari image masing-masing. Pembuatan badan hukum, unit, rencana, dan rincian berhasil. Setelah assignment global dicabut, scope badan hukum tetap menyaring detail, daftar, total, dan offset untuk rencana serta rincian.
