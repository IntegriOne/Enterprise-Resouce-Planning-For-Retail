# Penerimaan efek biaya, kredensial bukti DRM dan publikasi payroll

- **id** — 6
- **owner** — finance_accounting. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-01, KM-06, KM-07, KM-10.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-01, KM-06, KM-07, KM-10 yang dimiliki finance_accounting melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Penerimaan efek biaya, kredensial bukti DRM dan publikasi payroll

- file: [cmd/wiring.go](../../../backend/services/finance_accounting/cmd/wiring.go) · [internal/service/inventory_cost/service.go](../../../backend/services/finance_accounting/internal/service/inventory_cost/service.go) · [internal/integration/warehouse/client.go](../../../backend/services/finance_accounting/internal/integration/warehouse/client.go) · [internal/integration/drm/client.go](../../../backend/services/finance_accounting/internal/integration/drm/client.go) · [domain/service/ports.go](../../../backend/services/finance_accounting/domain/service/ports.go) · [internal/service/payroll/service.go](../../../backend/services/finance_accounting/internal/service/payroll/service.go) · [internal/service/accounts_payable/service.go](../../../backend/services/finance_accounting/internal/service/accounts_payable/service.go) · [contract/proto/finance_accounting/v1/payroll.proto](../../../backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `ApplyInventoryCostEvent`, `VerifyDocumentVersion`, `PublishPayrollRun`, `PublishPayroll`, `VerifyPayrollPublication`; koordinasikan lookup/ack dan collection dengan owner.
- logic: Terima efek per line dengan source effect, ledger entity, original effect untuk reversal dan dasar harga yang dapat direkonsiliasi. Sediakan lookup hasil idempoten untuk WMS. Perkuat verifier DRM agar mengikat type, source, owner, entity, subject/result dan digest expected. Run payroll disahkan Finance sebelum HRM render; setelah file sealed, catat penerbitan item secara terpisah. AP/AR/treasury punya aturan bukti sendiri. Untuk Commerce jadikan kontrak collection/tax/refund terpisah dari AP payout.
- error_handling: Outcome unknown → lookup owner, tahan posting kedua. Tolak mismatch sumber/versi. Penerbitan PDF, health check atau upload receipt tidak mengubah status pembayaran. Beri status gated untuk kontrak customer collection yang belum tersedia.
- security: Identitas workload, SHA-256 berversi, SoD dan scope entity. Jangan meminta byte slip untuk membuktikan identity metadata bila attestasi owner cukup. Jangan melibatkan bank/provider riil pada E2E.
- expected_result: Efek stok yang wajib memiliki tepat satu catatan valuasi; replay tidak menambah GL. Bukti PDF sah hanya diterima pada hasil payroll yang cocok; receipt pembayaran diterbitkan dari keputusan Finance.
- testing: Uji cost per baris dan original reversal, doc sah namun salah result/entity/jenis, Finance mati saat render, dua callback identik; cocokkan GL/valuasi dan PDF melalui caller sebenarnya.

## Action

Urutan pengujian untuk finance_accounting adalah efek valuasi → ACK WMS, lalu validasi bukti DRM → penerbitan payroll per hasil. Perluas [pengujian yang sudah ada](../../../backend/services/finance_accounting/internal/service/inventory_cost/test) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik finance_accounting dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
