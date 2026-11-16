# Validator opening cuti HRM

Menambahkan validator produksi policy `annual_grant`, entitlement dan ledger `opening`; wiring ke `BuildDependencies`. Policy yang dipakai entitlement dibekukan. Entitlement dimulai dengan saldo nol; posting awal dibatasi hak tahunan/maksimum saldo dan direkonstruksi dari ledger dalam transaksi. Tes gRPC–PostgreSQL `_test` mencakup jalur berhasil dan penolakan mutasi langsung, duplikat, grant berlebih, dan usage. Fmt/vet/race/build lulus; container HRM healthy. Kode lain tetap fail-closed. Belum E2E IAM/Gateway atau sumber Workforce.
