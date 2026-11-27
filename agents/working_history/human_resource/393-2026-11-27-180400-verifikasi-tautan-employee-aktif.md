# Verifikasi tautan employee aktif

Commit `ad45203` di `origin/main` menambahkan VerifyWorkerAccountLink pada listener mTLS IAM. HRM mencocokkan pekerja aktif, principal yang tertaut, dan badan hukum sebelum memberikan konfirmasi. Buf lint, tes race, dan vet lulus pada kandidat. Tes mencakup sertifikat yang tidak terverifikasi, workload lain, akun dan badan hukum berbeda, serta pekerja tidak aktif. RPC privat ini tidak menambahkan operasi GraphQL.
