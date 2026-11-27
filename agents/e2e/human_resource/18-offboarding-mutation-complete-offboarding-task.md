# Penyelesaian tugas offboarding

Tes memakai PostgreSQL sementara, policy dan verifier owner tiruan, serta race detector.

| Kasus | Hasil |
| --- | --- |
| Referensi konfirmasi milik tugas lain | FailedPrecondition |
| Konfirmasi cocok tetapi persetujuan rencana tercatat setelah waktu penyelesaian | FailedPrecondition |
| Konfirmasi cocok setelah waktu persetujuan fixture dipulihkan | Status completed, waktu dan referensi konfirmasi tersimpan |

Tes gRPC dengan PostgreSQL sementara dan race detector mengonfirmasi bahwa tugas ditolak dengan `FailedPrecondition` saat tanggal terminasi pekerja berbeda dari rencana. Setelah tanggal pada fixture diselaraskan, RPC mencatat bukti penyelesaian. Status terminasi pada tes berasal dari fixture database.
