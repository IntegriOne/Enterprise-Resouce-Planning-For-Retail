# Pengiriman perubahan SKU dan keputusan lifecycle berbasis fakta owner

- **id** — 2
- **owner** — product_lifecycle. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-03, KM-04, KM-06, KM-09.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-03, KM-04, KM-06, KM-09 yang dimiliki product_lifecycle melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Pengiriman perubahan SKU dan keputusan lifecycle berbasis fakta owner

- file: [cmd/wiring.go](../../../backend/services/product_lifecycle/cmd/wiring.go) · [infrastructure/outbox/router.go](../../../backend/services/product_lifecycle/infrastructure/outbox/router.go) · [services/integration/service.go](../../../backend/services/product_lifecycle/services/integration/service.go) · [infrastructure/integration/procurement/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/procurement/client.go) · [infrastructure/integration/drm/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/drm/client.go) · [services/revision/service.go](../../../backend/services/product_lifecycle/services/revision/service.go) · [services/changecycle/service.go](../../../backend/services/product_lifecycle/services/changecycle/service.go). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `PublishPendingEvents`, `Router.Publish`, `ArchiveProduct`, `CheckItemOpenCommitments`, `VerifyVendorWorkOrderScope`, `LinkControlledDocument`.
- logic: Pasang publisher nyata per consumer yang relevan untuk release/lifecycle/handling/UOM/identifier. Simpan ack per tujuan; rollback/ulang hanya target tertunda. Archive harus memverifikasi WMS residual, komitmen Procurement dan BOM aktif. Untuk tech pack, resolve vendor profile ke party, validasi WO/status dan SKU/revisi serta bukti DRM versi tepat. Jaga pembacaan versi historis setelah revisi baru.
- error_handling: Owner belum merespons atau port kosong → unavailable untuk keputusan yang membutuhkannya. Fakta nol memerlukan bukti otoritatif. Kesalahan dari satu consumer mempertahankan pekerjaan yang belum diakui.
- security: Autentikasi workload, validasi scope vendor dan klasifikasi versi DRM. Middleware dan client harus mendukung profil TLS deployment yang disetujui.
- expected_result: Rilis SKU menghasilkan receipt di consumer. SKU dengan komitmen terbuka tidak terarsip, dan vendor yang berhak atas satu WO hanya melihat item/revisi yang terkait.
- testing: Uji dua WO satu vendor, dua SKU satu WO bila berlaku, PO masih terbuka, Procurement down, archive sesudah closure; retry fan-out dengan satu ACK gagal dan pin tech pack di DRM lewat service aktual.

## Action

Urutan pengujian untuk product_lifecycle adalah revisi → ACK tiap tujuan, lalu cek komitmen SKU → arsip dan mandat vendor → versi tech pack. Perluas [pengujian yang sudah ada](../../../backend/services/product_lifecycle/services/masterdata/test) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik product_lifecycle dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
