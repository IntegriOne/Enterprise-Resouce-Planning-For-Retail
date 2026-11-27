# Memulai rencana onboarding

Pengujian pada transport gRPC memakai PostgreSQL sementara serta policy dan dependency tiruan. Klasifikasi Mutation digunakan untuk handoff GraphQL.

| Kasus | Hasil |
| --- | --- |
| Rencana tanpa tugas wajib | FailedPrecondition |
| Tugas wajib tanpa principal penanggung jawab | FailedPrecondition |
| Principal pekerja ditetapkan pada tugas wajib | Rencana berhasil dimulai |
| Rencana aktif dimulai ulang | FailedPrecondition |
| Audit tugas setelah siklus lengkap | Tiga event outbox untuk pembuatan, penugasan, dan penyelesaian |
