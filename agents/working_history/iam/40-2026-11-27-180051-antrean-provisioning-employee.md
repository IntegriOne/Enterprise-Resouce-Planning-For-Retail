# Antrean provisioning employee

Commit `c4442a2` terverifikasi pada `origin/main`. Penyimpanan pengajuan IAM mendukung penerimaan idempoten, satu klaim atomik, asosiasi akun, dan jejak penautan. Baseline `000001` melengkapi database IAM dengan transaksi di bawah advisory lock.

Tes race PostgreSQL disposable pada kandidat dan worktree lulus untuk instalasi baru, rekonsiliasi ledger lama, retensi user, konflik permintaan ganda, dan klaim bersamaan. Vet dan tes paket lulus. Handler provisioning, verifikasi tautan HRM, dan E2E melalui HTTPS masih perlu dipublikasikan.
