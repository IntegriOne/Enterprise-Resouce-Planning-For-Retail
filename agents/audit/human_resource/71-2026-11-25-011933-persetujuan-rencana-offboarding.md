# Persetujuan rencana offboarding

RPC privat ApproveOffboardingPlan memakai izin human_resource.offboarding_plans.approve dengan step-up. Actor manusia harus berbeda dari pembuat draf. Validator memeriksa status pekerja, batas tanggal efektif, dan bahwa hanya metadata persetujuan serta status yang berubah. Persetujuan mencatat waktu server dan outbox dalam transaksi yang sama. Keputusan persetujuan tidak mengakhiri hubungan kerja.

Integrasi gRPC bufconn ke PostgreSQL sementara lulus untuk penolakan self approval, persetujuan independen, penolakan pengulangan, dan jumlah event outbox. Policy IAM pada tes memakai stub. Jalur mulai, penyelesaian tugas dan waiver, clearance, serta penyelesaian plan masih terbuka.
