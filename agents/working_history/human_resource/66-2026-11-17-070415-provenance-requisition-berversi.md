# Provenance requisition berversi

Migrasi `000002` menambah pembuat/pengusul requisition dan waktu submit tanpa mengubah baseline. Create draf kini mengisi pembuat dari actor server; CRUD tidak dapat mengubah provenance atau menjalankan workflow. Generator DBML divalidasi agar tidak menimpa migrasi historis, sementara model generated mengikuti DBML terkini. Tes PostgreSQL mencakup upgrade, idempotensi, constraint SoD dan gRPC; pemeriksaan Go serta Docker HRM lulus. Approval dan pembukaan requisition masih menunggu command serta bukti owner Foundation/Finance/IAM.
