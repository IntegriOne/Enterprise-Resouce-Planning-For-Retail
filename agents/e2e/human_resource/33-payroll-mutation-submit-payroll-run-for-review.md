# Pengajuan payroll untuk review

Tes memakai gRPC privat, PostgreSQL sementara dan policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Hasil pekerja kedua belum lengkap | InvalidArgument |
| Principal selain penginisiasi mengajukan | PermissionDenied |
| Penginisiasi mengajukan hasil lengkap | Status review tanpa metadata approval |
| Pengajuan diulang | FailedPrecondition |
| Principal independen menyetujui | Status approved dan event terpisah |
| Mengubah digest run, digest hasil, atau baris komponen saat review | PostgreSQL menolak |
| Mengembalikan status review ke calculated dengan SQL | PostgreSQL menolak |

Percobaan approval dari status `calculated` ditolak oleh RPC dan trigger PostgreSQL. Setelah submit review, penginisiasi tetap ditolak dan penyetuju independen berhasil.
