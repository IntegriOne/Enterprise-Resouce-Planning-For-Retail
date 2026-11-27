# Bekukan snapshot payroll saat review

PostgreSQL kini mencegah mutasi run, hasil, dan komponen setelah pengajuan review. Transisi ke review harus berasal dari calculated tanpa perubahan snapshot. Approval tetap berlaku pada snapshot yang sama. Tes gRPC–PostgreSQL mereproduksi perubahan digest sebelum perbaikan dan memverifikasi seluruh penolakan serta approval sesudahnya. Migrasi naik–turun–naik lulus.
