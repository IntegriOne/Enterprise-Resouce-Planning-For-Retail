# Grant SELF eksplisit untuk slip gaji

Snapshot IAM dan snapshot bersama kini mengabaikan grant global ketika mengevaluasi izin `human_resource.payslips.read_self`. Tes membuktikan grant global saja ditolak untuk principal sendiri dan penambahan grant SELF sah berhasil walaupun grant global muncul lebih dulu. Pengujian domain dan sesi IAM, shared authz, adapter Gateway, dan policy HRM lulus. Kontrak GraphQL tidak berubah.
