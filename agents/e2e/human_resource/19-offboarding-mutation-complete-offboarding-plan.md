# Finalisasi rencana offboarding

Tes gRPC menggunakan PostgreSQL sementara, final payroll dari fixture, verifier clearance tiruan, dan race detector.

| Kasus | Hasil |
| --- | --- |
| Referensi clearance ditolak verifier | FailedPrecondition |
| Tanggal terminasi pekerja berbeda dari tanggal efektif rencana | FailedPrecondition |
| Hasil pekerja berstatus cancelled pada run terminasi yang sudah disetujui | FailedPrecondition |
| Tugas wajib berstatus completed dengan waktu sebelum persetujuan rencana | FailedPrecondition dan rencana tetap in_progress |
| Bukti tugas sesuai setelah waktu fixture dipulihkan | Rencana completed, final payroll dan referensi clearance tersimpan |
| Finalisasi diulang | FailedPrecondition dan outbox hanya bertambah satu kali |

Run terminasi berstatus `approved` dengan hasil pekerja berstatus `paid` ditolak sebagai referensi yang tidak konsisten. Tes PostgreSQL dan race detector lulus.
