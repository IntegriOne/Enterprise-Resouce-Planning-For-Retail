# Audit permintaan draf penempatan pekerja

Tanggal: 17 November 2026 WIB. Scope: `human_resource` personnel dan dampak kompensasi; tanpa perubahan IAM/Gateway/Foundation.

## Temuan dan perbaikan

- Sebelum perubahan, CRUD penempatan sudah memiliki kontrak tetapi mutasinya gagal tertutup karena validator belum terdaftar. Skema tidak mencatat siapa yang meminta penempatan, sehingga pemisahan tugas approval tidak dapat dibuktikan.
- Migrasi `000007` menambah `requested_by_principal_id` dan `requested_at`, indeks pengaju, serta CHECK pasangan pengajuan, pasangan approval dan SoD/waktu approval. DBML dan model mengikuti skema; draf yang dibuat melalui service mengambil pengaju/waktu dari actor manusia terverifikasi.
- Validator draf mengunci worker dan posisi, memeriksa status, legal entity, periode, tipe, serta manager HRM. Status pekerjaan dan approval dari klien ditolak; update/delete generik tetap meminta workflow. Pencarian pengaju menggunakan paginasi dan scope server.
- Generator model resmi kini melewati file model/validasi rekrutmen yang dipelihara manual; output generator berikutnya tidak menimpa aturan migrasi aditif tersebut.

## Bukti verifikasi

- `go generate ./contract` dan generator model DBML berhasil; hanya proto personnel yang berubah pada keluaran kontrak.
- `go fmt ./...`, `go vet ./...`, dan `go test -race ./...` pada modul HRM lulus.
- Tes `-race -tags=integration` untuk paket `db`, `transport/grpc/handler/personnel/test`, dan `transport/grpc/handler/compensation` lulus pada PostgreSQL `_test` terpisah; database uji sudah dihapus. Matriks mencakup pengaju server, finder terpaginasikan, approval palsu, update/delete generik, constraint SQL SoD, outbox tunggal, serta migrasi down/up.
- `docker compose build human-resource` dan `docker compose up -d --wait human-resource` berhasil; container healthy dan development DB memiliki tujuh migrasi.

## Batas klaim

Tes memakai policy/verifier tiruan dan posisi fixture. Approval, FTE temporal, aktivasi posisi, penerapan penempatan primer, grant IAM dan E2E Gateway belum tersedia. Migrasi pada basis data lama yang memiliki approval tanpa provenance akan gagal tertutup sampai data tersebut diremediasi; development DB sebelum migrasi memiliki nol baris penempatan.
