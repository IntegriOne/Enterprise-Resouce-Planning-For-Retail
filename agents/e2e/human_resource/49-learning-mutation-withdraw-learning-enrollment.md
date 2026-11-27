# WithdrawLearningEnrollment

Tes integrasi gRPC melalui bufconn dengan PostgreSQL disposable dan race detector lulus. Policy menggunakan fixture.

| Kasus | Hasil |
| --- | --- |
| Enrollment sudah completed | `FailedPrecondition`. |
| Enrollment sedang berjalan dan pekerja disuspensi | `OK`, status `withdrawn`. Waktu mulai yang dibaca dari PostgreSQL tetap sama dan waktu selesai tidak diisi. |
| Mulai enrollment yang sudah withdrawn | `FailedPrecondition`. |
| Ulangi withdrawal setelah sukses | `FailedPrecondition`. |
| Jejak outbox untuk enrollment yang ditarik | Satu peristiwa `hr_learning_enrollments.withdraw`. |
| Nonaktifkan kursus setelah enrollment mencapai status akhir | `OK`; penugasan baru setelahnya ditolak. |

[Handoff Learning](../../grpc/hrm/11-operasi-learning.md)
