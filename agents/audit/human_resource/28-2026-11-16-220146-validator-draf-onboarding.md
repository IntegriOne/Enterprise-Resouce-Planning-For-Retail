# Audit draf onboarding HRM — 16 November 2026

## Cakupan

Audit ini membuktikan CRUD draf `hr_onboarding_plans` dan `hr_onboarding_tasks` melalui service gRPC privat. Ini belum membuktikan lifecycle onboarding sampai selesai atau grant IAM.

## Aturan yang diaktifkan

- Plan: worker HRM harus `preboarding`/`active`; `(worker_id, plan_code, plan_version)` unik; status tetap `draft`; template eksternal ditolak; worker, ID dan principal pembuat tidak dapat diganti; plan yang memiliki tugas tidak dapat diubah/dihapus.
- Task: parent plan dikunci sebelum worker; status plan harus `draft`; kode unik per plan; tipe PIC dibatasi enam nilai skema; principal penugasan hanya boleh sama dengan link IAM worker untuk tugas bertipe `worker`; jatuh tempo berada dalam tanggal plan (UTC); penyelesaian/bukti eksternal tidak dapat ditulis lewat CRUD.
- Semua mutasi tetap melewati policy, transaksi repository dan outbox; kesalahan parent, duplikasi, dan workflow dipetakan menjadi status gRPC yang aman.

## Bukti pengujian

- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`: lulus pada modul HRM.
- `go test -tags integration -race ./transport/grpc/handler/onboarding/test -run TestOnboardingDraftRPCWithPostgreSQL -count=5`: lulus dengan PostgreSQL Docker database khusus `human_resource_onboarding_test`; mencakup create/read/update/delete plan dan task, finder/list berhalaman, duplikasi, parent hilang, perubahan status/penyelesaian ilegal, referensi child, dan jumlah outbox.
- `docker compose build human-resource` dan `docker compose up -d --wait human-resource`: lulus; container `healthy`.

## Batas dan pekerjaan berikutnya

Tes integrasi memakai policy tiruan dan `bufconn`; belum menguji grant/scope IAM, HMAC/mTLS runtime, maupun Gateway. Katalog `start`, `complete` plan dan `complete` task masih rancangan. Verifikasi versi bukti DRM, relasi PIC selain worker, audit aktor/waktu workflow, dan E2E lintas management diperlukan sebelum aktivasi. 27 dari 78 resource mutasi kini memiliki validator; 51 tetap fail-closed.
