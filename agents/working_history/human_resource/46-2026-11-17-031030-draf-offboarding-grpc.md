# Aktivasi draf offboarding HRM

Menambahkan validator produksi untuk rencana dan tugas offboarding, terpasang pada composition root. Draf mengunci worker/parent, menjaga satu rencana aktif, tanggal serta kode tugas, dan menolak status final maupun referensi payroll, Finance, DRM atau IAM yang belum dibuktikan. Tes unit mock dan gRPC `bufconn`→PostgreSQL `_test` disposable memeriksa CRUD, pagination, penolakan, transaksi/outbox. Total validator menjadi 29/78 resource; 49 tetap fail-closed. Command clearance/approval/finalisasi dan E2E grant IAM belum tersedia.
