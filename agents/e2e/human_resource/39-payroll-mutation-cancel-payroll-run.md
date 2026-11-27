# Pembatalan run payroll setelah review

Tes menggunakan gRPC privat, PostgreSQL sementara, policy dan kalkulator tiruan, serta race detector.

| Kasus | Hasil |
| --- | --- |
| Batalkan run yang belum masuk review | `FailedPrecondition` |
| Penginisiasi membatalkan run sendiri | `PermissionDenied` |
| Penyetuju mengirim alasan kosong | `InvalidArgument` |
| Reviewer independen membatalkan run | Status `cancelled`, alasan, principal, dan waktu tersimpan, digest tetap sama |
| Setujui run yang sudah dibatalkan | `FailedPrecondition` |
| Ubah baris hasil run yang dibatalkan | PostgreSQL menolak |
| Periksa outbox | Satu event `hr_payroll_runs.cancel` |

Verifikasi lanjutan dengan PostgreSQL dan race detector lulus. Pembatalan ulang ditolak dan jumlah event tetap satu. PostgreSQL menolak pembukaan kembali run. Penginisiasi berhasil membuat run draf pengganti pada periode yang sama dengan nomor run, idempotency key, dan snapshot baru. Status serta digest run lama tetap utuh.

Pada run pengganti yang sudah masuk review, approval dan pembatalan dijalankan serentak. Tepat satu RPC berhasil, RPC lainnya menerima `FailedPrecondition`, dan tepat satu event keputusan tercatat. Skenario ini lulus tiga kali dengan PostgreSQL sementara dan race detector.
