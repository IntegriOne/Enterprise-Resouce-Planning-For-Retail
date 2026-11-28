# Mandat buyer, consumer katalog dan saga checkout

- **id** — 2
- **owner** — web_commerce. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-03, KM-05, KM-09, KM-10.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-03, KM-05, KM-09, KM-10 yang dimiliki web_commerce melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Mandat buyer, consumer katalog dan saga checkout

- file: [cmd/wiring.go](../../../backend/services/web_commerce/cmd/wiring.go) · [infrastructure/integration/foundation/client.go](../../../backend/services/web_commerce/infrastructure/integration/foundation/client.go) · [infrastructure/integration/finance/client.go](../../../backend/services/web_commerce/infrastructure/integration/finance/client.go) · [infrastructure/outbox/router.go](../../../backend/services/web_commerce/infrastructure/outbox/router.go) · [services/buyer/service.go](../../../backend/services/web_commerce/services/buyer/service.go) · [services/checkout/service.go](../../../backend/services/web_commerce/services/checkout/service.go) · [services/integration/service.go](../../../backend/services/web_commerce/services/integration/service.go) · [services/order/service.go](../../../backend/services/web_commerce/services/order/service.go). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `VerifyB2BPartyMandate`, `OpenBuyerProfile`, `Publish`, `RelayPendingOutbox`, invalidasi quote/listing, `PrepareCheckoutQuote` dan `SubmitCheckout`.
- logic: Hubungkan EF verifier principal+party+role+entity dan cek kembali mandat saat menerima quote/membaca dokumen B2B. Ganti publisher log dengan delivery per owner; efek invalidasi dan inbox processed harus satu transaksi yang benar-benar sukses. Wiring Pricing untuk quote/kuota/expiry, OMS untuk submit/lookup dan Finance untuk tax, sesi bayar, refund; WMS untuk hold/fence/expiry. Pertahankan fase yang belum mempunyai owner sebagai gated. Simpan ref dokumen Finance/DRM per transaksi dan partial shipment.
- error_handling: Tolak mandat yang kosong/expired; bila invalidasi gagal jangan ACK. Perubahan harga/cart membatalkan quote; outcome OMS/Finance tidak pasti direkonsiliasi dengan key stabil. Jangan membuka prepaid dengan return true statis.
- security: Dapatkan identitas dari sesi opaque, batasi batch/Node dan nested field, pisahkan B2C SELF/B2B. Jangan menambah endpoint browser selain /api/graphql.
- expected_result: Perubahan PLM mencapai proyeksi Commerce sekali dan quote stale ditolak; buyer B2B hanya dapat mewakili pihak yang memberi mandat. Submit ganda menghasilkan satu order dan satu sesi bayar ketika gate siap.
- testing: Uji EF unavailable/revoke, buyer beda party, redelivery release SKU dan diskon kedaluwarsa, dua submit paralel, hold expiry dan late payment dengan owner nyata. Kasus belum siap diuji sebagai penolakan aman.

## Action

Urutan pengujian untuk web_commerce adalah mandat buyer → penawaran B2B, lalu revisi katalog → invalidasi quote, terakhir prasyarat checkout. Perluas [pengujian yang sudah ada](../../../backend/services/web_commerce/services/storefront/test) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik web_commerce dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
