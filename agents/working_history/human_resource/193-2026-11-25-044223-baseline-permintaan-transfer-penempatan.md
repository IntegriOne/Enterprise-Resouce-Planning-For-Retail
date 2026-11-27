# Baseline permintaan transfer penempatan

DBML dan migrasi tunggal mencatat permintaan transfer secara terpisah dari penempatan primer. Constraint mengikat penempatan sumber ke pekerja dan penempatan pengganti ke pekerja serta posisi tujuan. Keputusan harus berasal dari principal berbeda dan penerapan wajib merujuk penempatan pengganti. Tes SQL pada PostgreSQL sementara lulus untuk baseline up, regresi integritas, down, dan up kembali. Repository, command transaksi, dan RPC menyusul.
