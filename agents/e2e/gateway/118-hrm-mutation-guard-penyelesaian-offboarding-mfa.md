# Uji HTTPS guard penyelesaian tugas dan penutupan offboarding

Dua akun fixture memakai MFA dan assignment IAM melalui GraphQL. Stack disposable memakai konfigurasi HRM standar tanpa verifier konfirmasi tugas dan clearance. Pekerja aktif dan verifikasi akun kedua disiapkan melalui fixture SQL khusus `*_test`. Pembuatan tugas, persetujuan independen, mulai offboarding, waiver, dan terminasi menggunakan API.

| Kasus | Hasil |
| --- | --- |
| `hrmCompleteOffboardingTask` untuk tugas pending dengan ID bukti sintetis | `UNAVAILABLE`, detail tugas tetap pending sebelum waiver |
| `hrmCompleteOffboardingPlan` setelah terminasi dan waiver dengan referensi sintetis | `UNAVAILABLE`, detail rencana tetap `in_progress` |
| PostgreSQL | Waktu penyelesaian tugas dan rencana tetap kosong. Tugas hanya berubah karena waiver terpisah. |

Database dan Redis uji dibersihkan setelah runtime development kembali sehat. Jalur sukses memakai verifier konfirmasi dan clearance. Regresi PostgreSQL HRM telah membuktikan jalur dengan verifier fixture. [Bukti gRPC](../human_resource/54-query-mutation-regresi-grpc-postgresql.md) · [Handoff offboarding](../../grpc/hrm/13-operasi-offboarding.md).
