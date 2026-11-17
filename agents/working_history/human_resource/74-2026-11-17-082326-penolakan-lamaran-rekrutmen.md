# Penolakan lamaran rekrutmen

Menambahkan `RejectApplication` dengan izin sensitif, alasan wajib, keputusan actor/waktu server, stage history dan outbox atomik. Transisi received/screening diverifikasi pada kandidat serta requisition; screening hanya oleh recruiter yang ditugaskan. Migrasi `000004` mewajibkan metadata status rejected. Tes unit/mock, gRPC–PostgreSQL, DBML, migrasi up/down/up, fmt/vet/race/build dan Docker healthy lulus. Alur shortlist/interview/assessment/offer serta E2E IAM/owner nyata masih perlu dikerjakan.
