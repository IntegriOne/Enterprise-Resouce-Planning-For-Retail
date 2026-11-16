# Approval koreksi cuti lewat gRPC

Menambah `ApproveLeaveAdjustment` dengan permission sensitif tersendiri, SoD pengusul/penyetuju, audit actor/waktu server, dan outbox. Validator koreksi membatasi draft serta approval; validator ledger menerima koreksi HRM yang sah dan merekonstruksi semua entry sambil menjaga `max_balance`. Kontrak protobuf dihasilkan ulang dengan generator resmi. Tes gRPC–PostgreSQL `_test`, fmt/vet/race/build lulus; Docker HRM healthy. Katalog kini 324 kode dan validator 18/78. IAM/Gateway E2E dan sumber Workforce belum tersedia.
