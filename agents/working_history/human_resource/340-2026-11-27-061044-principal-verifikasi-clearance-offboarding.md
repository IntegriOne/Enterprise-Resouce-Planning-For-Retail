# Principal untuk verifikasi clearance offboarding

Kontrak internal verifier membedakan operator dari principal IAM yang tertaut ke pekerja. HRM mengambil principal pekerja dari record dalam transaksi penyelesaian offboarding. Tes gRPC dengan PostgreSQL dan race detector memastikan fixture verifier menerima dua principal yang berbeda dan rencana berhasil ditutup dengan bukti yang cocok. Adapter produksi masih memerlukan kontrak dari IAM dan Finance.
