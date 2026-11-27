# Sinkronisasi inventaris RPC

Pemeriksa handoff mengenali otorisasi langsung pada pembacaan transfer dan mencatat pemetaan detail requisition. Inventaris mengoreksi izin provisioning `iam.users.write` dan status verifier runtime. Pemeriksaan ulang menghasilkan 1.014 RPC dengan nol selisih. Pengujian negatif menolak kode izin lama.
