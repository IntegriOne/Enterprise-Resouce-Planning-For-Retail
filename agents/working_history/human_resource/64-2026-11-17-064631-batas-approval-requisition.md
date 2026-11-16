# Batas approval requisition

Memperketat guard CRUD requisition: `opened_at` tidak dapat diisi langsung dan requisition non-draf tidak dapat diubah/dihapus. Tes gRPC–PostgreSQL disposable membuktikan penolakan draf dan status `open`; race/vet/build serta Docker healthy lulus. Audit menemukan schema belum menyimpan pembuat/pengusul untuk pemisahan tugas dan runner migrasi hanya menerima baseline checksum `000001`. Approval/open tetap tertutup sampai migrasi aditif, verifikasi owner, dan command berizin tersedia. Rincian pada audit HRM nomor 48.
