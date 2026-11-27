# Menyelesaikan wawancara

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy tiruan dan fixture feedback.

| Kasus | Hasil |
| --- | --- |
| Belum ada feedback panel | FailedPrecondition |
| Feedback terakhir mendahului jadwal | FailedPrecondition |
| Versi baru dalam rentang waktu yang sah | Status completed |
| Mengulang penyelesaian | FailedPrecondition |
| Outbox penyelesaian | Satu event |
