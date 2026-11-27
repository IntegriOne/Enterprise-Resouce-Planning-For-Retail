# Aktivasi pekerja dan provisioning akun melalui HTTPS

Stack terisolasi memakai Gateway, IAM, HRM, Enterprise Foundation, Redis, dan database PostgreSQL `*_test`. Build IAM khusus pengujian menambahkan kode sensitif `human_resource.workers.activate` lewat overlay katalog. Role uji dibuat, diberi izin, dan dipasang ke superadmin melalui GraphQL. Role dan penugasan kemudian dicabut melalui GraphQL.

| Skenario | Hasil |
| --- | --- |
| Pekerja dibuat melalui `hrmCreateWorker` dengan referensi badan hukum dan individu EF | Pekerja berstatus `preboarding`. |
| Pengajuan akun saat pekerja masih preboarding | `CONFLICT`. |
| Aktivasi sebelum MFA | `AUTHORIZATION` pada payload. |
| Aktivasi setelah MFA, sebelum kontrak ditandatangani | `CONFLICT`. |
| Aktivasi melalui `hrmActivateWorker` setelah kontrak bertanda tangan tersedia | Pekerja menjadi `active`. HRM menerima peristiwa perekrutan yang dibuat oleh command saat pengajuan berikutnya. |
| Pengajuan akun dan percobaan ulang | Menghasilkan satu penerimaan IAM. |
| `createEmployeeUser` untuk pekerja berbeda | Ditolak. |
| Penerbitan akun dari antrean dan percobaan ulang | IAM menyimpan satu akun. Principal pada pekerja cocok dan status HRM menjadi `applied`. |

Kontrak bertanda tangan berasal dari fixture di database HRM terisolasi karena verifier DRM produksi belum tersedia. Status pekerja dan peristiwa perekrutan dihasilkan oleh command HRM. Runner lokal `/tmp/erp-cross-service.UUJx1n/run-activation-command.py` selesai dengan exit code 0.
