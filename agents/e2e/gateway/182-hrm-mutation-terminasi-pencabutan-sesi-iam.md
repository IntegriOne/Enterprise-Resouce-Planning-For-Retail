# Uji HTTPS terminasi pekerja dan pencabutan sesi IAM

Jalankan `python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py --termination` setelah mengaktifkan overlay workload mandat dan terminasi. Uji RPC privat memakai `grpcurl` dan memerlukan akses lokal ke jaringan bridge Docker. Runner mengarahkan kedua koneksi mTLS ke service disposable, memakai database `*_test` dan Redis terpisah, serta menghapus resource uji saat selesai.

| Kasus | Hasil |
| --- | --- |
| Superadmin login dan memberi grant terminasi melalui GraphQL IAM, lalu menyelesaikan MFA | Lulus |
| Provisioning employee melalui bukti aktivasi HRM, replay, dan korelasi akun | Lulus |
| Employee login sebelum terminasi dan memiliki sesi HTTPS | Lulus |
| Superadmin menjalankan `terminateWorker` pada rencana offboarding berjalan | Status pekerja `terminated` |
| Relay HRM memublikasikan event sesudah IAM mengonfirmasi | Satu event terkirim, status IAM `disabled` |
| RPC privat IAM mengembalikan clearance untuk worker, rencana, tanggal, dan principal yang cocok | Lulus melalui mTLS |
| Sesi lama employee mengakses `me` | `UNAUTHENTICATED` |
| Login baru dengan kredensial employee | Ditolak |
| Superadmin mencabut assignment/permission dan menghapus role uji melalui GraphQL | Lulus |

Rencana offboarding yang telah disetujui independen disiapkan sebagai fixture basis data HRM. Skenario dimulai pada langkah terminasi dan tidak mencakup pembuatan, persetujuan, atau clearance rencana. Role dan grant IAM seluruhnya dikelola melalui GraphQL. [Kontrak gRPC dan konfigurasi](../../grpc/hrm/40-deaktivasi-akun-setelah-terminasi.md)
