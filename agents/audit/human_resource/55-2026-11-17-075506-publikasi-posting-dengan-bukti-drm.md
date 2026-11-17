# Audit publikasi posting rekrutmen HRM

## Cakupan dan temuan

Command `PublishJobPosting` dan RPC privat mengubah draf menjadi `published` hanya setelah pemeriksaan actor manusia, requisition terbuka, digest 32 byte, dan bukti DRM yang verified, immutable, serta cocok pada versi, digest, owner `human_resource`, dan purpose `recruitment_job_posting`. Validator menjaga perubahan status/waktu saja dalam transaksi yang mengunci parent. Outbox ditulis atomik. Izin `human_resource.job_postings.publish` dipisahkan dari CRUD dan diklasifikasikan sensitif.

## Bukti pengujian

- `buf lint` dan `buf generate`: lulus.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- `go test -race -tags=integration ./transport/grpc/handler/recruitment/test -run TestRecruitmentDraftRPCWithPostgreSQL -count=1 -v`: lulus pada database PostgreSQL disposable. Menguji publish, retry, larangan CRUD status, dan jumlah outbox.
- `docker compose build human-resource` dan `docker compose up -d --wait human-resource`: lulus; container healthy.

## Batasan

Runtime produksi belum memiliki adapter verifier DRM; operasi sengaja mengembalikan `Unavailable`. Tes memakai mock bukti DRM, belum koneksi antarmanagement atau grant IAM nyata. `published` menyatakan rilis konten untuk distribusi, belum keberhasilan tayang pada kanal eksternal. Close/withdraw posting belum tersedia.
