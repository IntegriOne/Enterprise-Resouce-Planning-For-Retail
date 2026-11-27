# Penugasan pelatihan melalui gRPC

Tes menggunakan PostgreSQL sementara, race detector, dan policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Penugasan baru saat pekerja memiliki riwayat expired pada kursus yang sama | Riwayat terminal tidak menghalangi penugasan |
| Dua penugasan serentak dengan key berbeda untuk peserta dan kursus yang sama | Satu berhasil, satu AlreadyExists |
| Penugasan tambahan dengan key lain ketika peserta masih memiliki enrollment aktif | AlreadyExists |
| Ubah tenggat enrollment withdrawn, failed, atau expired | FailedPrecondition untuk setiap status |
| Isi skor melalui update enrollment planned | FailedPrecondition |
| Hapus enrollment planned yang belum diproses | Berhasil |

Status terminal disiapkan langsung pada database test. Kontrak Learning belum menyediakan RPC pencatatan hasil pelatihan.
