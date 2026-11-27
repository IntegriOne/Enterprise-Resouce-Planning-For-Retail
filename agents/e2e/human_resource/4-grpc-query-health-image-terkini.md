# Uji startup HRM gRPC pada image terkini

| Kasus | Hasil aktual |
| --- | --- |
| Build dari checkout | `docker compose build human-resource` lulus. |
| Startup dengan database kosong | Container sementara berjalan menggunakan image `87b5ee28c5c7`, database PostgreSQL khusus berakhiran `_test`, dan outbox mode standby. |
| `grpc.health.v1.Health/Check` | Probe mTLS bawaan image mengembalikan kode proses 0 setelah gRPC melaporkan `SERVING`. |
| Penghentian proses | Pada eksekusi terpisah, `docker stop -t 20` mengirim SIGTERM setelah probe lulus. Proses berhenti dengan exit 0 dan `OOMKilled=false`. |
| Isolasi | Container HRM development tetap berjalan memakai image `33bd4519a2e8`. Container dan database probe dibersihkan setelah tes. |

Tes memeriksa startup dan health gRPC. Sesi IAM, grant, serta kontrak DRM/Finance tidak tercakup.
