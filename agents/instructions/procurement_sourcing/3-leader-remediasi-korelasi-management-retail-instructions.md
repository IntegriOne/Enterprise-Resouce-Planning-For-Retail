# Bukti PO, receipt dan mandat vendor untuk WMS dan PLM

- **id** — 3
- **owner** — procurement_sourcing. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-02, KM-04.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-02, KM-04 yang dimiliki procurement_sourcing melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Bukti PO, receipt dan mandat vendor untuk WMS dan PLM

- file: [cmd/wiring.go](../../../backend/services/procurement_sourcing/cmd/wiring.go) · [infrastructure/outbox/relay.go](../../../backend/services/procurement_sourcing/infrastructure/outbox/relay.go) · [internal/service/fulfillment/service.go](../../../backend/services/procurement_sourcing/internal/service/fulfillment/service.go) · [internal/service/integration/service.go](../../../backend/services/procurement_sourcing/internal/service/integration/service.go) · [internal/service/purchaseorder](../../../backend/services/procurement_sourcing/internal/service/purchaseorder) · [internal/service/vendorcontrol](../../../backend/services/procurement_sourcing/internal/service/vendorcontrol) · [contract/proto/procurement_sourcing/v1/fulfillment.proto](../../../backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/fulfillment.proto) · [contract/proto/procurement_sourcing/v1/purchase_order.proto](../../../backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/purchase_order.proto). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `ApplyAcceptedReceipt`, `ApplyReceiptReversal`, `ApplyReturnDispatched`, publisher PO/WO dan verifier untuk komitmen SKU serta scope WO.
- logic: Putuskan dengan WMS satu command/consumer kanonik per efek. Gunakan use case fulfillment yang ada. Tautkan PO/revisi/line, original effect, UOM, party dan qty; proyeksikan outstanding sesudah accepted. Wire event PO issued/revised/cancelled, WO dan return sesuai kebutuhan owner; dokumentasikan jika ada jalur lookup yang mengganti push. Sediakan fakta PLM tentang jumlah PO/WO terbuka per SKU dan kecocokan party EF ↔ vendor profile ↔ WO released ↔ material/revisi yang diminta.
- error_handling: Kontrak kosong/ambigu → unavailable atau conflict. Tahan event dengan digest berbeda dan nomor versi loncat. Cegah penerapan qty melebihi batas, duplicate dan penghitungan reversal ganda.
- security: Workload mTLS terikat purpose; PIC dan role diperiksa pada fakta EF. User dari party lain tidak boleh melihat harga atau tech pack.
- expected_result: WMS dan Procurement sepakat atas outstanding per line. PLM memperoleh keputusan otoritatif untuk SKU dan WO yang tepat; tidak perlu menganggap vendor_profile_id sama dengan party_id.
- testing: Uji issued → accepted partial → reversal → return; dua SKU dan dua WO vendor yang sama; penolakan archive saat masih ada PO terbuka. Verifikasi replay setelah ack hilang dan saldo akhir melalui API masing-masing owner.

## Action

Urutan pengujian untuk procurement_sourcing adalah penerimaan per baris → outstanding → reversal/return, lalu verifikasi PO/WO untuk PLM. Perluas [pengujian yang sudah ada](../../../backend/services/procurement_sourcing/internal/service/test/acceptance_test.go) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik procurement_sourcing dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
