# Audit batas approval requisition HRM

## Temuan

Guard CRUD `hr_requisitions` menolak perubahan status dan metadata approval, tetapi belum secara eksplisit menolak `opened_at`. Daftar status beku pada guard juga tidak mencakup `open`; validator saat ini masih menolak mutasi tersebut, namun guard harus tetap aman saat workflow resmi ditambahkan.

Katalog HRM mensyaratkan pengusul tidak menyetujui usulannya sendiri. Tabel requisition belum memiliki identitas pembuat/pengusul atau waktu submit. `RunMigrationsUp` hanya menerima baseline `000001` dengan checksum tetap; menulis ulang baseline akan menggagalkan database yang sudah terpasang. Referensi Foundation, mandat actor, serta anggaran/headcount untuk approval juga belum dapat dibuktikan di runtime produksi.

## Perbaikan

- Guard menolak `opened_at` lewat create/update CRUD dan seluruh update/delete ketika status lama bukan `draft`.
- Tes gRPC→PostgreSQL disposable menolak update timestamp pembukaan pada draf dan update/delete requisition `open` tanpa aplikasi.

## Bukti dan langkah berikutnya

Tes integrasi tersebut, seluruh `go test -race`, `go vet`, `go build`, dan Docker healthy lulus. Approval/open tetap belum diaktifkan. Langkah berikutnya adalah migrasi aditif berversi untuk identitas pengusul serta invariant status, command submit/approve/open dengan SoD dan validasi owner, kemudian E2E IAM/Gateway/HRM/Foundation/Finance sesuai jenis requisition. Mengganti identitas pengusul dengan hiring manager akan melemahkan aturan SoD.
