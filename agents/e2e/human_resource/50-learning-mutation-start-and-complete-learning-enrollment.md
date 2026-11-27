# StartLearningEnrollment dan CompleteLearningEnrollment

Tes `TestLearningCourseCompetencyAndEnrollmentWithPostgreSQL` memakai RPC privat melalui bufconn, PostgreSQL disposable, dan policy fixture. Suite berjalan dengan race detector.

| Kasus | Hasil |
| --- | --- |
| Penyelesaian diajukan sebelum enrollment dimulai | `FailedPrecondition`. |
| Waktu mulai melewati masa berlaku kursus meski tanggal penugasan masih sah | `FailedPrecondition`. |
| Mulai enrollment yang direncanakan setelah masa berlaku kursus dipulihkan | `OK`, status `in_progress` dan waktu mulai terisi. |
| Ulangi permintaan mulai | `FailedPrecondition`. |
| Nonaktifkan kursus saat enrollment sedang berjalan | `FailedPrecondition`. |
| Selesaikan enrollment yang sedang berjalan | `OK`, status `completed` dan waktu selesai pada atau setelah waktu mulai. |
| Ulangi penyelesaian | `FailedPrecondition`. |
| Nonaktifkan kursus setelah enrollment mencapai status akhir | `OK`; penugasan baru ditolak. |

Operasi mencatat progres administratif. Verifikasi kelulusan, sertifikat, serta grant IAM dan resolver Gateway produksi berada di tahap integrasi. [Handoff Learning](../../grpc/hrm/11-operasi-learning.md)
