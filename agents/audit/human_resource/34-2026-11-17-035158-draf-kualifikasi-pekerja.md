# Audit draf kualifikasi pekerja HRM

Tanggal: 17 November 2026 WIB. Cakupan: `hr_worker_qualifications` dari DBML, entity, repository, service, validator, katalog izin, handler gRPC dan wiring `cmd`. Tabel menyimpan kredensial pekerja, status verifikasi, referensi issuer Foundation, dan bukti DRM. CRUD generik sebelumnya tetap fail-closed karena validator produksi belum dipasang.

## Keputusan dan perubahan

- `CompetencyValidators` sekarang memasang validator kualifikasi dan mewajibkan repository worker/kualifikasi. Create dan update hanya menerima status `unverified` tanpa aktor/waktu verifikasi; update tidak dapat memindah worker atau mengubah tipe kualifikasi. Delete hanya menerima draf. Mutasi mengunci worker; create/update memerlukan worker aktif.
- Referensi issuer dan credential eksternal ditolak `Unavailable` sampai bukti dari owner Foundation/DRM tersedia. Verifikasi resmi tetap memerlukan command, izin `human_resource.worker_qualifications.verify`, pemisahan pencatat/pemeriksa, dan bukti owner; CRUD ini tidak mengesahkan kualifikasi.
- Baca dan mutasi kualifikasi masuk operasi sensitif yang memerlukan step-up independen dari flag IAM.

## Bukti

- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`: lulus pada modul HRM.
- Tes unit memeriksa create/update/delete draf, penolakan verifikasi, identitas immutable, referensi eksternal, worker nonaktif/hilang, dan pembatalan context.
- Tes gRPC `bufconn` ke PostgreSQL sementara `_test` membuktikan create, read, finder berhalaman, update metadata, delete, serta penolakan issuer dan update status verifikasi.

## Batas

Tes ini memakai policy tiruan, bukan grant IAM nyata. Verifikasi kredensial, bukti DRM, issuer Foundation, dan E2E HTTPS Gateway belum tersedia. Dari 78 resource dengan izin mutasi, 32 kini memiliki validator produksi; 46 masih fail-closed.
