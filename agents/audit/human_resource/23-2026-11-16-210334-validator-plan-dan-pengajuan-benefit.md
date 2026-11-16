# Audit validator plan dan pengajuan benefit HRM

## Temuan dan perubahan

RPC Benefits sudah terdaftar, tetapi mutasi `hr_benefit_plans` dan `hr_benefit_enrollments` sebelumnya ditolak oleh Integrity karena tidak ada validator produksi. `BenefitsValidators` kini mengikat repository plan, enrollment, dan worker di composition root. Plan lokal hanya menerima aturan eligibilitas kosong `{}` serta tanpa referensi provider atau aturan kontribusi eksternal. Kode, versi, dan yurisdiksi unik; plan yang sudah mempunyai enrollment dibekukan dari perubahan dan penghapusan.

Enrollment hanya dapat diajukan sebagai `pending` untuk worker aktif dan periode plan aktif. Validator mengunci worker lalu plan, menolak permintaan ganda pada worker/plan yang sama, memaku identitas enrollment, dan menolak field provider, nominal kontribusi, consent, verifikasi, serta status final melalui CRUD. Penghapusan permintaan pending tetap dimungkinkan meski plan kemudian nonaktif. Approval dan perhitungan kontribusi masih memerlukan command dan verifikasi owner tersendiri.

## Bukti dan batas

- Tes unit mock validator mencakup duplikasi versi, referensi eksternal, eligibility khusus, pembekuan plan, status dan nominal enrollment, plan nonaktif, serta penghapusan pending.
- Tes `bufconn` → PostgreSQL Docker `_test` diulang lima kali dengan race detector: create/update/delete/read/finder/pagination, duplikasi, penolakan approval langsung, periode di luar plan, pembekuan plan, dan outbox transaksional.
- `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...` lulus; Docker `human-resource` dibangun ulang dan healthy.
- Tes memakai policy tiruan. Belum ada grant IAM, scope parsial, Gateway GraphQL, bukti Foundation untuk provider, keputusan eligibilitas khusus, atau integrasi Finance untuk kontribusi. Validator aktif 20/78 resource mutabel; 58 lainnya fail-closed.
