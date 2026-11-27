# Mutasi enrollment terminal

Tes gRPC mereproduksi perubahan tenggat pada enrollment failed. Guard kini membekukan completed, failed, withdrawn, dan expired untuk CRUD. Tes PostgreSQL sementara membuktikan penolakan perubahan tenggat withdrawn, failed, dan expired. Riwayat expired tetap memungkinkan penugasan baru. Race detector, vet paket integrasi, dan diff check lulus.
