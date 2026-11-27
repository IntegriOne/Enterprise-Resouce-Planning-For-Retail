# Batas kedaluwarsa saldo cuti

Validator sebelumnya menerima expiry pada hari terakhir periode. Pengujian gRPC ke PostgreSQL mereproduksi posting berhasil pada pukul 23.59.59 UTC di hari terakhir. Validator sekarang mengizinkan posting setelah hari terakhir UTC, dengan tanggal efektif tetap pada hari terakhir periode. Pengujian membuktikan penolakan sebelum tengah malam, keberhasilan setelah tengah malam, serta penolakan debit parsial.

Tes modul dengan race detector, go vet dan integrasi gRPC ke PostgreSQL sementara lulus. Penolakan koreksi dan penutupan entitlement tanpa debit masih memerlukan representasi keputusan tersendiri karena skema aktif hanya menyimpan persetujuan koreksi dan saldo ledger.
