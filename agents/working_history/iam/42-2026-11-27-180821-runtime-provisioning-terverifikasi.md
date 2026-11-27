# Runtime provisioning terverifikasi

Commit `3827b06` terverifikasi di `origin/main`. IAM mendaftarkan handler provisioning, mengklaim pengajuan dan mengaitkan akun pending dalam satu transaksi, memverifikasi badan hukum lewat EF, serta membaca konfirmasi tautan HRM melalui mTLS. Aktivasi admin dan verifikasi email tertahan selama pengajuan belum tertaut.

Race suite IAM, vet, build command, dan build image kandidat terisolasi lulus. Penulisan tautan HRM, Gateway, dan E2E HTTPS dari main masih dalam pekerjaan.
