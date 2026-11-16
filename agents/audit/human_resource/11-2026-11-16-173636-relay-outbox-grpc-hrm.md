# Audit relay outbox gRPC HRM

Tanggal: 16 November 2026. Acuan: `agents/how_to_code_grpc_managements.md`, kontrak `hr_outbox_events`, dan batas versi event HRM.

## Implementasi dan bukti

- Relay berjalan bersama gRPC server dan berhenti saat SIGINT/SIGTERM; kegagalan worker membuat proses gagal agar Docker dapat restart.
- Store memilih event jatuh tempo dengan `FOR UPDATE SKIP LOCKED`, `published_at IS NULL`, dan `attempt_count` di bawah batas. Sukses mengisi `published_at`; kegagalan menunda `available_at` secara eksponensial. Kegagalan permanen tersimpan sebagai batas percobaan, tanpa menghapus baris.
- Mode default standby tidak mengubah pending. Mode webhook mengirim envelope dengan `event_id` untuk deduplikasi dan HMAC; mode log ditolak di production. Payload digest diverifikasi sebelum publish.
- `go generate`, `go fmt`, `go vet`, seluruh `go test -race -count=1 ./...` lulus. Tes integrasi pada database PostgreSQL disposable `_test` membuktikan claimant kedua tidak mengambil row terkunci dan `published_at` tersimpan. Build Docker lulus; container HRM healthy dengan log standby.

## Batas dan keputusan retensi

- Belum ada endpoint consumer nyata, sehingga belum ada bukti delivery lintas management. Mode standby mempertahankan antrean saat deployment development.
- Cleanup published/dead letter sengaja tidak dipasang: `operation.emit` masih menghitung aggregate version dari riwayat outbox. Menghapus baris dapat mengulang nomor versi. Strategi counter versi persisten perlu dibuat sebelum retensi aktif.
- Scope IAM parsial, validator 76 resource mutabel, dan E2E positif lintas Gateway/IAM/DB tetap belum selesai; goal HRM penuh belum tercapai.
