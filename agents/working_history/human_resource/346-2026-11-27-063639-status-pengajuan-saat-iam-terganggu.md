# Status pengajuan saat IAM terganggu

Query HRM mempertahankan pengajuan lokal dan menandainya verification_unavailable jika IAM tidak tersedia. Penolakan akses tetap dikembalikan sebagai error. Workspace menawarkan pemeriksaan ulang sambil menahan klaim bahwa akun sudah atau belum tertaut. Tes adapter mencakup Unavailable, NotFound, dan PermissionDenied.
