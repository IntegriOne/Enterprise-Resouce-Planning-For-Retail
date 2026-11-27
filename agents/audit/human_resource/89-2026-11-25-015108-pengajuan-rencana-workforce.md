# Pengajuan rencana workforce

SubmitWorkforcePlan memakai izin step-up. Pembuat draf dengan actor manusia dapat mengirim rencana yang belum melewati periode akhir dan memiliki setidaknya satu baris. Validator mengunci rencana, memeriksa badan hukum ke Foundation dan memastikan periode mencakup seluruh baris. Status review dan outbox dicatat atomik. Referensi anggaran atau persetujuan tetap kosong pada pengajuan.

Integrasi gRPC ke PostgreSQL membuktikan draf kosong ditolak, rencana berisi baris berhasil diajukan, pengajuan ulang ditolak, dan baris rencana review tidak dapat dihapus. Verifikasi Foundation dan IAM dalam tes memakai stub. Approval dengan bukti Finance, supersede dan cancel masih terbuka.
