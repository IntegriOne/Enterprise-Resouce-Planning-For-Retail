# Regresi gRPC HRM dan handoff Gateway

`GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...` dari modul HRM selesai dengan exit code 0 pada PostgreSQL Docker dengan basis data disposable `*_test`. Ada 68 paket tes lulus. Keluaran memuat paket pengujian handler untuk 22 domain bisnis dan helper respons. Cakupan paket mencakup regresi modul. Bukti ini tidak mengukur keberhasilan setiap RPC secara individual.

Rekonsiliasi 1.006 operasi membandingkan nama, tipe request dan response, klasifikasi Query atau Mutation serta permission terhadap proto, handler, service dan katalog HRM. Hasilnya 669 Query dan 337 Mutation cocok di 22 domain. [Inventaris](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md)

Koneksi produksi Finance dan DRM masih menjadi dependensi untuk penerbitan, pembayaran, dan akses PDF slip secara mandiri. Pengujian integrasi modul berjalan melalui fixture dan koneksi PostgreSQL dalam lingkungan tes. [Kebutuhan Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [Kebutuhan DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md)
