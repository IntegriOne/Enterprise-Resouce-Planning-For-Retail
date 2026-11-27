# Pembacaan komponen untuk persetujuan payroll

Persetujuan memuat komponen per kelompok maksimal 100 hasil pekerja. Pemeriksaan jumlah sumber terhadap hasil terfilter scope, batas 10.000 komponen, rekonsiliasi, dan transaksi outbox dipertahankan. Tes aplikasi melintasi batas kelompok dengan 101 hasil pekerja dan 404 komponen. Integrasi gRPC Payroll dengan PostgreSQL sementara memakai fixture satu hasil pekerja. Race detector, vet bertag integration dan diff check lulus.
