# Penutupan posting rekrutmen

Menambahkan `CloseJobPosting` dari service ke RPC privat dengan izin sensitif `human_resource.job_postings.close`. Validator hanya menerima `published` → `closed`, menjaga field lain dan mengunci requisition terbuka. Waktu tutup dan outbox tersimpan atomik. Tes unit/mock, gRPC–PostgreSQL disposable, fmt/vet/race/build, dan Docker healthy lulus. Tes requisition kini menutup posting lewat RPC tanpa reset status SQL. Penutupan kanal eksternal serta E2E IAM/DRM belum tersedia.
