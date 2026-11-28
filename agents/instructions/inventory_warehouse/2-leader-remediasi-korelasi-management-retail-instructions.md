# Pergerakan WMS sampai penerimaan Procurement dan valuasi Finance

- **id** — 2
- **owner** — inventory_warehouse. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-01, KM-02, KM-09.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-01, KM-02, KM-09 yang dimiliki inventory_warehouse melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Pergerakan WMS sampai penerimaan Procurement dan valuasi Finance

- file: [internal/service/movement/service.go](../../../backend/services/inventory_warehouse/internal/service/movement/service.go) · [internal/service/receiving/service.go](../../../backend/services/inventory_warehouse/internal/service/receiving/service.go) · [internal/service/internal/servicekit/helpers.go](../../../backend/services/inventory_warehouse/internal/service/internal/servicekit/helpers.go) · [internal/integration/finance/client.go](../../../backend/services/inventory_warehouse/internal/integration/finance/client.go) · [internal/integration/procurement/client.go](../../../backend/services/inventory_warehouse/internal/integration/procurement/client.go) · [internal/integration/outbox/router.go](../../../backend/services/inventory_warehouse/internal/integration/outbox/router.go) · [internal/service/integration/service.go](../../../backend/services/inventory_warehouse/internal/service/integration/service.go) · [cmd/main.go](../../../backend/services/inventory_warehouse/cmd/main.go). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `EmitOutbox`, `DeliverOutboxEvent`, `mapProcurementEventType`, `IsValuationServiceReachable` dan relay. Wire adapter produksi ke consumer tujuan.
- logic: Saat posting, simpan payload immutable atau reference yang benar-benar bisa di-resolve per business effect dan baris. Pisahkan physical receipt, accepted, ownership, custody, return dan reversal. Kirim harga/basis biaya sesuai kontrak Finance; kirim qty UOM, PO/revisi/baris dan efek asal ke Procurement. Idempotensi memakai ID efek yang tetap meskipun event transport diterbitkan kembali. Persistekan ACK per tujuan; tetapkan skip eksplisit hanya untuk jenis yang secara bisnis memang tidak berlaku.
- error_handling: Efek wajib tanpa payload, rute atau ACK tetap pending/retry. Timeout sesudah owner commit → lookup effect sebelum retry. Status health, NOT_FOUND atau penolakan izin bukan bukti valuasi tercatat. Rekonsiliasi event lama per owner sebelum replay selektif.
- security: Batasi RPC per identitas workload dan entity/facility, verifikasi digest dan producer. Pastikan event custody murni tidak menghasilkan jurnal atau valuasi.
- expected_result: Satu receipt diterima sebagian lalu dibalik menghasilkan qty bersih yang cocok di Procurement dan efek biaya yang cocok di Finance. Caller dapat menunjukkan ID receipt/efek dan ACK kedua pemilik.
- testing: Uji service + PostgreSQL uji dan owner nyata: dua line berbeda, partial accept, custody-only, reversal, crash setelah satu target ACK, redelivery dengan event ID berbeda dan owner mati lalu pulih.

## Action

Urutan pengujian untuk inventory_warehouse adalah posting → outbox → valuasi Finance, lalu receipt/reversal → outstanding Procurement. Perluas [pengujian yang sudah ada](../../../backend/services/inventory_warehouse/internal/service/acceptance_test.go) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik inventory_warehouse dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
