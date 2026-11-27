# Konsistensi inventaris RPC Gateway

Pemeriksa menemukan pemetaan kosong pada query detail requisition serta detail dan riwayat transfer. Detail requisition melewati metode agregasi detail sebelum pemeriksaan baca. Kedua metode transfer memakai otorisasi eksplisit. Skrip sekarang memeriksa bentuk otorisasi itu dan mencatat pemetaan detail requisition.

Handoff pengaitan akun kini mencatat `iam.users.write` dari katalog produksi serta verifier yang sudah terpasang di startup HRM. Pengaitan berjalan dalam provisioning IAM `createEmployeeUser`.

Pemeriksaan ulang mencocokkan 1.014 RPC pada 22 service dengan nol selisih. Regresi sintetis yang mengganti izin kembali ke kode lama dideteksi.
