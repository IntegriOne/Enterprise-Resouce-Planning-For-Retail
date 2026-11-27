# Waiver tugas offboarding

Tes gRPC memakai PostgreSQL sementara, race detector, dan policy tiruan. Tautan pekerja ke principal disiapkan langsung di fixture.

| Kasus | Hasil |
| --- | --- |
| Alasan kosong | InvalidArgument |
| Pelaku adalah principal pekerja yang menjadi subjek rencana | PermissionDenied |
| Pelaku berbeda dengan alasan sah | Status waived dan alasan serta pelaku tersimpan |
| Waiver diulang | FailedPrecondition tanpa duplikasi outbox |
