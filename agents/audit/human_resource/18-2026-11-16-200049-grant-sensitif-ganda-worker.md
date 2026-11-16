# Audit grant sensitif ganda pada policy HRM

## Temuan

Policy sebelumnya membaca grant pertama untuk suatu kode permission. Jika snapshot memuat kode yang sama pada beberapa scope dengan flag `IsSensitive` tidak konsisten, grant yang tidak sensitif dapat melewati syarat MFA/reauth. Pada pembacaan worker, pemilihan grant juga berhenti pada kandidat pertama. Ini risiko fail-open untuk snapshot yang tidak konsisten.

## Perbaikan dan bukti

Policy kini menggabungkan sensitivitas pada tingkat kode permission sebelum mengevaluasi grant global atau `legal_entity`. Grant ganda yang tidak sensitif tidak menurunkan syarat MFA; filter SQL baca worker tetap berasal dari legal entity yang grant-nya sah. Unit test mencakup dua legal entity, grant ganda, penolakan tanpa MFA, penerimaan setelah MFA, serta status snapshot nonaktif. `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, dan `go build ./...` lulus pada modul HRM. `docker compose build human-resource` dan `docker compose up -d --wait --no-deps human-resource` lulus; container berstatus healthy.

## Batas bukti

Tes memakai proyeksi sesi tiruan; belum ada E2E IAM/Gateway untuk snapshot grant ganda. Perbaikan ini tidak mengaktifkan mutasi personel yang masih memerlukan verifikasi Foundation/DRM.
