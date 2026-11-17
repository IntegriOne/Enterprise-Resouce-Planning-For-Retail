# Penutupan requisition dan guard posting

`CloseRequisition` kini RPC privat sensitif yang mengunci requisition, menolak posting published/aplikasi aktif, lalu menyimpan waktu tutup dan outbox atomik. Validator `hr_job_postings` mengunci parent untuk CRUD draf dan menolak pembuatan setelah close maupun publikasi lewat CRUD. Katalog 332 kode, validator 40/78. Tes unit/mock, gRPC–PostgreSQL `_test` dengan race, fmt/vet/build, dan Docker healthy lulus. Publikasi posting masih menunggu bukti DRM, sedangkan approval/open produksi menunggu verifier Foundation/Finance serta grant IAM.
