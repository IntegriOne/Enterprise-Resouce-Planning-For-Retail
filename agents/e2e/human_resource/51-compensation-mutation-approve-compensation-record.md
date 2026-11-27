# ApproveCompensationRecord

Tes `TestCompensationProposalRPCWithPostgreSQL` memakai gRPC bufconn, PostgreSQL disposable, policy fixture, dan race detector.

| Kasus | Respons atau hasil |
| --- | --- |
| Approval disisipkan melalui create CRUD | `FailedPrecondition`. |
| Pembuat mengesahkan proposal sendiri | `PermissionDenied`. |
| Editor mengesahkan proposal yang diubahnya | `PermissionDenied`. |
| Pekerja menjadi penyetuju kompensasi miliknya | `PermissionDenied`. |
| Penyetuju independen mengesahkan proposal | `OK`, status `approved`, principal dan waktu persetujuan terisi. |
| Ulangi persetujuan atau hapus proposal yang disetujui | `FailedPrecondition`. |
| Setujui proposal yang periodenya bertumpang tindih untuk komponen dan penempatan yang sama | `FailedPrecondition`; proposal tetap `proposed`. |
| Outbox setelah percobaan gagal dan persetujuan berhasil | Tepat satu event approval. |

Aktivasi efektif kompensasi dan grant IAM produksi berada di tahap berikutnya. [Handoff Compensation](../../grpc/hrm/6-operasi-compensation.md)
