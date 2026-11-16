# Integrasi rencana suksesi gRPC dan PostgreSQL

Tanggal: 17 November 2026 WIB. Audit nomor 36 memasang validator produksi untuk draf `hr_succession_plans`; audit ini membuktikan jalur positifnya melalui gRPC `bufconn` → service → repository GORM → PostgreSQL pada database disposable `hrm_disposable_*_test`.

## Kasus yang diuji

| Kasus | Hasil |
| --- | --- |
| Create draf untuk posisi aktif pada tanggal review sah | Berhasil; `confidentiality_class=highly_restricted`. |
| Create kedua dengan posisi, kode dan versi sama | `AlreadyExists`; tidak ada draf kedua. |
| Find berdasarkan posisi dan GetAll dengan limit/offset | Total dan jumlah item sesuai halaman. |
| Update `succession_risk` melalui CRUD | `FailedPrecondition`; field workflow terlindungi. |
| Update tanggal review draf melalui FieldMask | Berhasil dan tanggal baru terbaca. |
| Create kandidat tanpa validator produksi | `Unavailable`; tidak ada baris kandidat. |
| Delete draf tanpa kandidat, lalu Get by ID | Delete berhasil dan lookup `NotFound`. |
| Outbox create/update/delete | Tiga event tersimpan untuk agregat plan. |

Fixture job family, job dan posisi aktif ditulis langsung ke database uji karena CRUD posisi belum memiliki validator produksi. Policy tiruan mengizinkan operasi agar alur database teruji; tes ini tidak membuktikan IAM grant, HMAC/mTLS runtime, atau otorisasi Gateway. Kandidat tetap sengaja gagal tertutup sampai aturan readiness dan panel tersedia.

## Verifikasi

`GOWORK=off go test -race -tags integration ./transport/grpc/handler/succession/test -run TestSuccessionPlanRPCWithPostgreSQL -count=1 -v` lulus pada PostgreSQL Docker dengan database disposable dan opt-in eksplisit. `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./...` lulus pada modul HRM. Database uji dihapus; hitungan database `hrm_disposable_*_test` setelah tes adalah nol. Tes tidak mengubah database HRM development.
