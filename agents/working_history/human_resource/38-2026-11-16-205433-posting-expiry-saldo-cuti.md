# Posting expiry saldo cuti HRM

Menambah `expiry` pada command posting ledger HRM dengan debit penuh saldo tersisa tepat akhir periode. Setelah expiry, posting dan koreksi baru ditolak. Pemeriksaan urutan expiry memakai `entry_sequence`, karena halaman repository terurut UUID. Tes gRPC–PostgreSQL `_test` dijalankan 20 kali dengan race detector dan mencakup penolakan dini, parsial, ulang, serta posting setelah expiry. Fmt, vet, seluruh tes race, build, dan Docker HRM diverifikasi. IAM/Gateway E2E belum tersedia.
