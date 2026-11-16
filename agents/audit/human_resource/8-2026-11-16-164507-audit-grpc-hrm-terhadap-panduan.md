# Audit gRPC HRM terhadap panduan management — 16 November 2026

Status: **audit transport selesai, kesiapan bisnis terbatas**. Audit source dibandingkan dengan `agents/how_to_code_grpc_managements.md`, `agents/arsitektur.md`, `agents/boundary.md`, dan kontrak HRM aktif. Bukti runtime Docker dan batasnya dicatat di bawah.

## Temuan dan perbaikan

| Area | Temuan | Perbaikan / status |
| --- | --- | --- |
| Registrasi 22 domain | `server.go` merujuk `handler.RegisterAll` yang belum ada sehingga `go test ./cmd/... ./transport/grpc/...` gagal setup. | `handler/register.go` dibuat dengan daftar domain eksplisit dan penolakan dependency nil/typed nil. Seluruh 22 handler dan 84 resource sudah dapat dikompilasi. |
| Kontrak JSON | `google.protobuf.Value` mengonversi angka JSON ke floating point dan berisiko mengubah angka >2^53. | `common.JsonDocument` berupa byte UTF-8 valid menggantikan kolom JSON; null SQL dan literal JSON `null` berbeda; uji angka besar pada helper lulus. Domain lain sedang disinkronkan. |
| Batas transport | Keepalive dan health bernama service belum ada. | Ditambahkan `MaxConnectionIdle=5m`, ping `2m`, timeout `20s`, enforcement `20s`, serta health global dan service. |
| Error gRPC | Konflik dipetakan ke `Aborted`; belum ada `ErrorInfo`/`BadRequest`; status yang dibungkus berpotensi membocorkan error asli. | Konflik → `FailedPrecondition`; sentinel utama → ErrorInfo, field validasi aman → BadRequest; status dibersihkan saat keluar. Uji redaksi/pembatalan/proof lulus. |
| Autentikasi | Bukti HMAC harus diverifikasi; metadata identitas ganda dapat ambigu. | Interceptor menolak proof invalid/kedaluwarsa dan metadata identitas ganda; tes unit lulus. |

Inventaris kontrak: 915 deklarasi RPC pada 22 proto domain dan 915 method handler konkret (`*Handler`/`*server`), termasuk 84 resource bisnis. Jumlah setara bukan bukti tiap permutasi data dan hak telah E2E; cek unit/wire dan runtime dipisahkan.
| Permission | Panduan menyebut penegakan hanya di Gateway, sedangkan keputusan D01–D12 dan `agents/boundary.md` mewajibkan service pemilik memeriksa permission, scope dan objek. | Pemeriksaan service tidak dicabut. Katalog HRM masih draf; runtime menolak operasi bisnis tanpa mapping. Perbedaan dokumen dicatat untuk koreksi owner, tanpa melonggarkan D01–D12. |

## Kesenjangan yang masih terbuka

- Katalog permission belum disetujui, tidak ada grant IAM baru. Validator integrity runtime masih fail-closed tanpa implementasi parent/owner lengkap. Tidak boleh menyebut CRUD bisnis siap pakai.
- Seluruh handler telah dikompilasi, tetapi E2E positif CRUD semua resource, scope lintas principal, dan Gateway GraphQL belum dilakukan. Bukti gRPC nyata saat ini hanya health serta penolakan autentikasi.
- Pola `internal/service` dan entity bertag GORM merupakan baseline HRM yang sudah disepakati sebelumnya, berbeda dari contoh layout Enterprise Foundation di panduan. Pemindahan massal membutuhkan audit dampak kontrak; jangan mengubahnya hanya untuk menyamakan nama folder.
- Outbox relay lintas service serta kontrak scope owner belum lengkap. Listener sehat saja tidak membuktikan alur bisnis siap.

## Bukti verifikasi pada checkout ini

- `GOWORK=off go generate ./...` dari `contract/`: lulus (`buf lint` dan `buf generate`).
- `GOWORK=off go vet ./...`, `GOWORK=off go test -race -count=1 ./...`, dan `gofmt -l` pada modul HRM: lulus sebelum perapian terakhir; jalankan ulang cek terdampak sebelum commit.
- `docker compose build human-resource`: lulus pada Ubuntu 26.10/Go 1.27.4. `docker compose up -d --no-deps --wait human-resource`: healthy. Redis, PostgreSQL, IAM, Gateway dan service lain tetap healthy.
- Database HRM mencatat tepat satu migrasi `000001` dengan checksum 64 karakter. Setelah restart container HRM, tracker tetap satu baris dan container kembali healthy.
- RPC personnel dengan sertifikat klien tetapi tanpa konteks pemanggil ditolak `Unauthenticated`; tanpa sertifikat klien handshake TLS ditolak. Sertifikat dan secret hanya berada di file lokal yang diabaikan Git.
