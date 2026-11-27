# Overlap penempatan primer

`ApplyPrimaryAssignment` sebelumnya membandingkan akhir penempatan lama dengan waktu penerapan. Pada data historis, penempatan pengganti yang mulai pada hari terakhir penempatan lama masih dapat dipilih setelah hari itu berlalu.

Validator kini membandingkan tanggal akhir penempatan lama dengan tanggal mulai penempatan pengganti. Kedua rentang inklusif harus terpisah. Tes unit mereproduksi overlap dan memastikan pergantian pada hari berikutnya tetap diterima. Integrasi gRPC Personnel dengan PostgreSQL sementara dan race detector serta go vet lulus.
