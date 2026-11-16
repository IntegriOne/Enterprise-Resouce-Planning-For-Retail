# Validasi worker preboarding

Menambah validator `hr_workers` untuk create dan koreksi tanggal `preboarding`, dengan identitas/nomor tetap dan larangan link IAM, aktivasi, assignment, terminasi, serta delete melalui CRUD. Port Foundation wajib membuktikan party orang dan legal entity; wiring produksi belum memasang adapter sehingga create/update gagal tertutup. Tes unit/mock, gRPC–PostgreSQL disposable dengan verifier tiruan, race/vet/build, dan Docker healthy lulus. Total validator terdaftar 38/78. Detail dan batas pada audit HRM nomor 46.
