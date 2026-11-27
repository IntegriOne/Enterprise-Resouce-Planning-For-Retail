# Uji HTTPS draf offboarding dan pemisahan persetujuan

Stack uji memakai database disposable IAM, Foundation, HRM, dan Redis TLS sementara. Badan hukum, party, dan pekerja dibuat melalui API. Status pekerja disiapkan sebagai aktif melalui fixture SQL yang memeriksa nama database `*_test`, karena skenario ini menguji offboarding setelah prasyarat aktivasi terpenuhi. Sesi superadmin fixture memakai TOTP, CSRF, dan role IAM sementara.

| Kasus | Hasil |
| --- | --- |
| `hrmCreateOffboardingPlan` untuk pekerja aktif | Berhasil, status `draft` |
| `hrmCreateOffboardingTask` untuk tugas wajib HRM | Berhasil, status `pending` |
| `hrmApproveOffboardingPlan` oleh pembuat rencana | `FORBIDDEN` |
| `hrmOffboardingPlan`, `hrmOffboardingTask`, `hrmOffboardingTasks` | Status draf, tugas pending, dan satu tugas muncul pada daftar |
| Pemeriksaan PostgreSQL | Satu rencana `draft` dan satu tugas `pending` |

Runtime development kembali sehat setelah database dan Redis sementara dibersihkan. Persetujuan berhasil memerlukan principal kedua dengan izin sesuai. [Aturan offboarding](../../grpc/hrm/13-operasi-offboarding.md).

## Daftar rencana per pekerja

Pengulangan pada stack disposable juga menjalankan `hrmOffboardingPlans`. GID pekerja yang memiliki draf mengembalikan satu rencana dan total satu. UUID pekerja yang sama dengan limit satu dan offset satu mengembalikan daftar kosong dengan total tetap satu. UUID pekerja yang tidak ada mengembalikan daftar kosong dengan total nol. Runtime semula dipulihkan dan database serta Redis uji dihapus.
