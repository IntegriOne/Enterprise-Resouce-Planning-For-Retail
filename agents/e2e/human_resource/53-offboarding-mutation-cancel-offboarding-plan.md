# Pengujian RPC pembatalan rencana offboarding

`TestOffboardingDraftRPCWithPostgreSQL` lulus dengan `-race -tags=integration` pada database sementara PostgreSQL Docker. Tes menjalankan handler gRPC dan migrasi HRM dengan fixture policy dan verifier lintas management.

| Kasus | Hasil |
| --- | --- |
| Batalkan rencana draf | `FailedPrecondition` |
| Batalkan rencana disetujui | Status `cancelled`, alasan dan waktu pembatalan tersimpan, serta approver dipertahankan |
| Mulai rencana yang dibatalkan | `FailedPrecondition` |
| Buat rencana pengganti untuk pekerja yang sama | Draf berhasil dibuat |
| Mulai dan batalkan rencana yang sama secara bersamaan | Satu berhasil, satu mendapat `FailedPrecondition`, dan satu event outbox tercatat |

Suite yang sama mencakup regresi koreksi dan penghapusan rencana, selesainya tugas, pengecualian tugas; cabang penutupan dengan fixture clearance memerlukan database disposable dengan opt-in. [Wiring Gateway](../../grpc/hrm/26-status-wiring-gateway-hrm.md).

Cabang penutupan dengan fixture clearance juga kemudian lulus pada [suite database disposable](54-query-mutation-regresi-grpc-postgresql.md).
