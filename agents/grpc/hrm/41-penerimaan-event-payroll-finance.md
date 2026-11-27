# List operasi

| RPC privat HRM | Bukti |
| --- | --- |
| `DeliverPayrollEvent` | Jurnal posting, penerbitan slip, pembayaran individual, atau settlement akhir. |

## Penjelasan DeliverPayrollEvent

Permintaan membawa UUID run, versi skema, dan satu jenis bukti. HRM menurunkan sumber Finance dari identitas sertifikat sebelum meneruskan permintaan ke handler bisnis. Bukti hanya membawa referensi. Verifier pemilik tetap mencocokkan referensi dengan snapshot yang tersimpan. Balasan sukses diberikan setelah handler selesai, sehingga kegagalan sementara dapat diproses ulang.

## Butuh GraphQL

RPC privat ini tidak menambah Query atau Mutation GraphQL. Operator menggunakan `requestPayrollPosting` untuk mengajukan posting.

## RBAC

Receiver menerima workload dengan rantai sertifikat terverifikasi dan SAN URI tunggal `spiffe://integrione/finance-accounting`. Caller operator Gateway tidak dapat mengirim event melalui RPC ini. Kontrak dan handler tersedia. Aktivasi listener, PKI, dan verifier runtime masih perlu sinkronisasi dengan pemilik Finance. [Proto](../../../backend/services/human_resource/contract/proto/human_resource/workload/payroll/v1/payroll_event.proto) · [Catatan Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md)

`NewPayrollEventServer` membuat instance gRPC terpisah memakai TLS 1.3 dan CA klien yang dikonfigurasi. Probe loopback dengan sertifikat uji membuktikan koneksi tanpa sertifikat ditolak, sertifikat IAM dari CA tepercaya mendapat `PERMISSION_DENIED`, dan sertifikat Finance yang cocok meneruskan bukti posting ke receiver fixture. Daemon mendukung listener opsional. Wiring deployment menunggu konfigurasi dan verifier Finance. [Server workload](../../../backend/services/human_resource/transport/grpc/payroll_events.go)

Regresi `TestPayrollEventServerAuthenticatesFinanceCertificate` membuat CA dan sertifikat sementara lalu melakukan handshake TLS 1.3 asli lewat gRPC in-memory. Tes memeriksa caller tanpa sertifikat, sertifikat IAM dari CA tepercaya, serta sertifikat Finance dengan SAN URI yang sesuai. [Tes mTLS](../../../backend/services/human_resource/transport/grpc/payroll_events_test.go)

## Konfigurasi startup

`HRM_PAYROLL_EVENTS_ADDR` mengaktifkan listener privat pada port yang berbeda dari gRPC bisnis dan verifier terminasi IAM. Set juga `HRM_PAYROLL_EVENTS_CERT_FILE`, `HRM_PAYROLL_EVENTS_KEY_FILE`, dan `HRM_PAYROLL_EVENTS_CLIENT_CA_FILE`. Jika alamat diminta dan semua provider payroll masih kosong, startup berhenti sebelum listener payroll dibuka. Tahap yang providernya belum dipasang tetap mengembalikan `UNAVAILABLE` saat dipanggil. Startup standar tidak membuka port payroll.

Daemon menghentikan listener bisnis dan payroll dalam tenggat shutdown yang sama. Event payroll aktif diberi kesempatan selesai sebelum koneksi diputus. Uji mTLS memblokir receiver sampai tenggat dan memastikan event yang belum selesai tidak mendapat acknowledgement sukses.

Setelah tenggat, daemon memutus koneksi gRPC dan melanjutkan shutdown meski verifier mengabaikan pembatalan context. Regresi menahan receiver secara sengaja sampai koneksi diputus dan memeriksa bahwa pemanggil tidak menerima sukses.
