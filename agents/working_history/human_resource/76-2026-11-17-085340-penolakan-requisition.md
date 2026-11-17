# Penolakan requisition

Menambahkan `RejectRequisition` dari DBML/migrasi `000006` hingga gRPC, dengan permission sensitif, pemisahan actor, alasan/waktu keputusan, status terminal dan outbox atomik. Finder alasan/actor ditambahkan. Tes unit/race, gRPC–PostgreSQL, migrasi down/up, DBML, fmt/vet/build dan Docker healthy lulus. Katalog HRM kini 338 kode tanpa grant IAM. Approval/open produksi tetap memerlukan bukti owner Foundation/Finance; E2E Gateway/IAM belum tersedia.
