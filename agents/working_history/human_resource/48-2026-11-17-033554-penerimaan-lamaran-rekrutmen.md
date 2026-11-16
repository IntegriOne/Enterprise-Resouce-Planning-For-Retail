# Penerimaan lamaran HRM

Memperbaiki guard yang menolak `applied_at` server pada `CreateApplication`. Validator lamaran mengizinkan create `received` setelah mengunci kandidat dan requisition, memeriksa notice/retensi dan waktu pembukaan lowongan. Field seleksi, keputusan, hiring, posting serta update/delete tetap ditolak sampai workflow tersedia. Tes unit/race dan gRPC→PostgreSQL disposable lulus; fmt/vet/build modul HRM lulus. Katalog validator menjadi 31/78. Requisition `open` masih fixture, IAM/Gateway belum diuji nyata.
