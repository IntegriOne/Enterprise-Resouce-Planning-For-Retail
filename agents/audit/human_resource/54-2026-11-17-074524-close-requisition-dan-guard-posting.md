# Audit penutupan requisition dan guard posting HRM

## Temuan

Requisition `open` belum mempunyai command penutupan. `hr_job_postings` belum mempunyai validator, sehingga CRUD draf tidak mengunci requisition parent dan dapat berlomba dengan penutupan. Posting published dan aplikasi aktif harus mencegah penutupan.

## Perbaikan

- Validator `hr_job_postings` mengunci requisition untuk create/update/delete draf. Create/update hanya pada requisition open yang disetujui, status dan atribut publikasi tidak dapat diisi lewat CRUD, serta posting yang direferensikan aplikasi tidak dapat diubah/dihapus. Draf tanpa aplikasi masih dapat dibersihkan setelah requisition closed.
- RPC `CloseRequisition` memakai permission sensitif `.close`, actor manusia, lock requisition dan pemeriksaan posting `published` serta aplikasi belum terminal. Perubahan status, waktu tutup server dan outbox berada dalam satu transaksi. Guard validator membekukan semua field lain.
- Katalog runtime menjadi 332 kode; validator mutasi terdaftar menjadi 40/78. Tidak ada grant IAM baru.

## Bukti dan batas

Unit/mock memeriksa parent lock, draft/publish/closed, child reference, status close dan retry. Tes gRPC `bufconn`→PostgreSQL `_test` dengan race memeriksa posting published dan aplikasi received menahan close, terminal state mengizinkan close, outbox tunggal, serta create posting setelah close ditolak. `buf lint/generate`, fmt/vet/race/build dan Docker healthcheck lulus. Fixture SQL hanya untuk menguji status child dan bukan bukti workflow publikasi/aplikasi terminal. Versi dokumen dan digest draf posting belum diverifikasi DRM; publish/close posting belum tersedia. Approval/open produksi masih gagal tertutup tanpa verifier Foundation/Finance dan grant IAM.
