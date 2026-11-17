# Audit provenance requisition berversi

## Temuan

Requisition belum mencatat principal pembuat atau pengusul. Approval tidak dapat menegakkan pemisahan tugas bila hiring manager dipakai sebagai pengganti pengusul. Generator DBML lama menimpa baseline `000001` yang sudah tercatat dengan checksum.

## Perbaikan

- Migrasi `000002` menambah identitas pembuat/pengusul dan waktu submit; constraint mencegah waktu tanpa actor serta approver yang sama dengan pembuat/pengusul saat provenance tersedia.
- Kolom baru nullable untuk baris lama. Workflow approval kelak wajib menolak provenance kosong; tidak ada backfill identitas hasil tebakan.
- Create draf menyimpan principal actor dari policy server. CRUD tetap menolak metadata submit/approval dan perubahan pembuat.
- Generator DBML tidak lagi menulis baseline. Generator entity resmi membaca DBML terbaru agar model dan validasi mengikuti skema.

## Bukti dan batas

Tes PostgreSQL `_test` lulus untuk pemasangan baru, upgrade dari versi 1, idempotensi, versi asing, dan checksum rusak. Tes gRPC→PostgreSQL membuktikan pembuat tersimpan serta constraint penolakan. `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`, verifikasi SQL baseline up/down/up, build Docker dan healthcheck HRM lulus; tracker development berisi versi 1 dan 2. Command submit/approve/open, bukti anggaran/headcount owner, grant IAM, dan E2E Gateway belum tersedia.
