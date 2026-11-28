# Fakta representasi party dan kontrak verifier lintas owner

- **id** — 2
- **owner** — enterprise_foundation. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-05, KM-09.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-05, KM-09 yang dimiliki enterprise_foundation melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Fakta representasi party dan kontrak verifier lintas owner

- file: [contract/proto/enterprise_foundation/v1/party.proto](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/party.proto) · [services/party](../../../backend/services/enterprise_foundation/services/party) · [cmd/main.go](../../../backend/services/enterprise_foundation/cmd/main.go) · [transport/grpc](../../../backend/services/enterprise_foundation/transport/grpc). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: Periksa `ValidatePartyRoleScope`, hubungan PIC/portal contact dan keputusan legal entity. Tambahkan verifier representasi privat hanya bila kontrak yang ada tidak mengembalikan seluruh fakta.
- logic: Pisahkan kelayakan party sebagai target grant IAM dari kewenangan seorang user mewakilinya. Keputusan perlu mengikat principal IAM, party, peran, entity, resource/purpose yang relevan, status PIC/portal dan expiry/versi. Wiring caller DRM serta Commerce harus memeriksa identitas dalam response. Pertahankan verifikasi exact target yang sudah ada di IAM.
- error_handling: Party aktif dengan PIC tidak cocok → denied. Data tidak tersedia atau stale → unavailable; bedakan dari keputusan negatif. Perubahan PIC/revoke mengakhiri akses di request berikutnya.
- security: Batasi lookup fakta kepada workload berwenang. Jangan mengirim daftar seluruh kontak sebagai pengganti keputusan; jangan memberi grant IAM otomatis.
- expected_result: User dengan role yang sama memperoleh hasil berbeda sesuai mandat sebenarnya. Revoke portal link menutup akses B2B/DRM yang bergantung padanya.
- testing: Uji melalui owner aktif, PIC valid/expired, role vendor vs supplier, legal entity berbeda, IAM target cocok/mismatch serta timeout. Ulangi legalEntities pada kedua replika saat menelusuri anomali latency di laporan.

## Action

Urutan pengujian untuk enterprise_foundation adalah relasi principal/PIC → keputusan EF, lalu propagasi pencabutan ke Commerce dan DRM. Perluas [pengujian yang sudah ada](../../../backend/services/enterprise_foundation/services/party/test) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik enterprise_foundation dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
