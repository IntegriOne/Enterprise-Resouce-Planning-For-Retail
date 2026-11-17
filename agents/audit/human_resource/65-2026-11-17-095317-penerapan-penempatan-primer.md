# Audit penerapan penempatan primer — 17 November 2026

## Hasil

`ApplyPrimaryAssignment` menjadi command HRM dan RPC gRPC privat dengan permission sensitif tersendiri. Actor manusia meminta penerapan pada worker aktif. Worker dikunci; validator membaca assignment approved bertipe `primary`, periode yang mencakup tanggal server, posisi aktif dalam badan hukum sama, dan bukti Foundation untuk worker. Jika worker sudah memiliki assignment primer, periode lama harus berakhir sebelum tanggal penerapan. Hanya `primary_assignment_id` dan `updated_at` worker berubah; outbox ditulis dalam transaksi yang sama. Guard CRUD menolak perubahan assignment primer dan pengulangan command ditolak.

## Bukti

- `go generate ./contract`, `go fmt ./...`, `go vet ./...`, `go test -race ./...`: lulus.
- Tes gRPC `bufconn` → PostgreSQL Docker `_test` sekali pakai: lulus untuk referensi salah, penerapan approved assignment, dua permintaan paralel (satu sukses, satu ditolak), baca worker setelah commit, penolakan CRUD, dan outbox tunggal untuk penerapan.
- `docker compose build human-resource` dan `docker compose up -d --wait human-resource`: lulus; container healthy.

## Batas

Tes positif memakai worker aktif dan posisi aktif sebagai fixture serta verifier Foundation tiruan. Aktivasi worker dari preboarding, kontrak bertanda tangan, adapter Foundation produksi, penutupan assignment primer, dan E2E IAM/Gateway belum tersedia. `updated_at` menjadi tanggal keputusan penerapan; event outbox menyimpan actor dan waktu audit.
