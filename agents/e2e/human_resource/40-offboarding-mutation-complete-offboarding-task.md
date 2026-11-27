# CompleteOffboardingTask melalui gRPC dan PostgreSQL

Tes `TestOffboardingDraftRPCWithPostgreSQL` lulus dengan race detector dan database PostgreSQL sementara. Verifier konfirmasi memakai fixture dengan penghitung panggilan.

| Kasus | Hasil |
| --- | --- |
| Tugas masih berada pada rencana draf | `FailedPrecondition` dan verifier tidak dipanggil |
| Rencana berjalan dengan referensi konfirmasi tidak cocok | `FailedPrecondition` |
| Rencana berjalan dengan waktu persetujuan di masa depan | `FailedPrecondition` |
| Rencana berjalan dengan bukti dan waktu yang valid | Tugas selesai dan referensi konfirmasi tercatat |

Pemeriksaan prasyarat dilakukan dalam transaksi sebelum permintaan bukti. Jalur produksi tetap membutuhkan verifier terhubung.
