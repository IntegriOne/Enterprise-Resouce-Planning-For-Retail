# Perubahan tanggal pembayaran payroll

Tes memakai gRPC privat dan PostgreSQL sementara. `TestPayrollRPCAndApprovalWithPostgreSQL` lulus dengan race detector.

| Kasus | Hasil |
| --- | --- |
| Ubah tanggal dengan field mask `pay_date` sebelum periode dirujuk | Tanggal baru tersimpan dan dikembalikan oleh RPC |
| Ubah tanggal setelah penyesuaian merujuk periode | `FailedPrecondition` |
| Jalankan kalkulasi setelah perubahan ditolak | Kalkulator menerima tanggal yang terakhir berhasil disimpan |
