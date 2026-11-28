# Koordinasi perbaikan lintas management, transport dan bukti HTTPS

- **id** — 2
- **owner** — gateway. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-05, KM-06, KM-09, KM-11.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-05, KM-06, KM-09, KM-11 yang dimiliki gateway melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

### Instruksi untuk tiap pemilik

- [inventory_warehouse — Pergerakan WMS sampai penerimaan Procurement dan valuasi Finance](../inventory_warehouse/2-leader-remediasi-korelasi-management-retail-instructions.md)
- [procurement_sourcing — Bukti PO, receipt dan mandat vendor untuk WMS dan PLM](../procurement_sourcing/3-leader-remediasi-korelasi-management-retail-instructions.md)
- [finance_accounting — Penerimaan efek biaya, kredensial bukti DRM dan publikasi payroll](../finance_accounting/6-leader-remediasi-korelasi-management-retail-instructions.md)
- [product_lifecycle — Pengiriman perubahan SKU dan keputusan lifecycle berbasis fakta owner](../product_lifecycle/2-leader-remediasi-korelasi-management-retail-instructions.md)
- [web_commerce — Mandat buyer, consumer katalog dan saga checkout](../web_commerce/2-leader-remediasi-korelasi-management-retail-instructions.md)
- [enterprise_foundation — Fakta representasi party dan kontrak verifier lintas owner](../enterprise_foundation/2-leader-remediasi-korelasi-management-retail-instructions.md)
- [document_record — Otorisasi versi, bukti payroll, retensi dan penghapusan tahan crash](../document_record/2-leader-remediasi-korelasi-management-retail-instructions.md)
- [hrm — Penyambungan payroll Finance dan akses DRM per worker](../hrm/7-leader-remediasi-korelasi-management-retail-instructions.md)

## Request

### Koordinasi perbaikan lintas management, transport dan bukti HTTPS

- file: [internal/source_connections/registry.go](../../../backend/gateway/internal/source_connections/registry.go) · [internal/source_connections/document_record_client.go](../../../backend/gateway/internal/source_connections/document_record_client.go) · [internal/source_connections/product_lifecycle_client.go](../../../backend/gateway/internal/source_connections/product_lifecycle_client.go) · [internal/source_connections/inventory_warehouse_client.go](../../../backend/gateway/internal/source_connections/inventory_warehouse_client.go) · [internal/adapter/document_record/ingestion_pub.go](../../../backend/gateway/internal/adapter/document_record/ingestion_pub.go) · [internal/adapter/enterprise_foundation/legal_entity_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/legal_entity_adapter.go) · [internal/e2e](../../../backend/gateway/internal/e2e/). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `DocumentContentChunk`, authorizer scoped, registry upstream/transport per service, fixture akun uji dan verifikasi kedua replika Gateway.
- logic: Gunakan policy owner terikat subject/resource di setiap entrypoint termasuk Node dan nested; teruskan konteks terpercaya dan hak akses terbatas. Dukung trust root, peer identity dan client cert pada hop sensitif. Ganti pembuatan aktor uji yang mengubah user_kind via SQL dengan provisioning resmi lewat IAM–HRM/EF; setup dan cleanup grant melalui GraphQL. Uji flow lintas service dengan correlation dan ack per target, sertakan jumlah skip. Selidiki pembacaan EF yang sekali timeout lalu pulih dengan trace dua replika.
- error_handling: Gangguan upstream harus terisolasi dan tidak mencabut sesi IAM. Hasil health/SDL dan hasil baca dengan denied/unavailable diberi status berbeda dari acceptance antar-owner. Tes yang tidak memiliki DSN test wajib jelas skip atau gagal di job wajib.
- security: Opaque session, CSRF, origin, MFA dan scope tetap berlaku. Superadmin menyiapkan role melalui GraphQL; login sebagai akun peer, inactive/denied dan workload sah dalam uji berbeda. Redaksi cookie, token, nomor rekening dan byte dokumen.
- expected_result: Pada dua replika, metadata dan konten yang sama mengikuti keputusan owner yang sama. Satu downstream gagal tidak mengganggu login. Laporan final memiliki bukti aktor resmi, outcome bisnis dan ack di owner penerima.
- testing: HTTPS Caddy dengan trusted CA; happy path, role salah, objek lintas batas, logout/rekonsiliasi, failure isolation dan retry. Pakai data uji disposable dan cek definisi fixture sebelum mengulang E2E.

## Action

Urutan pengujian untuk gateway adalah provisioning aktor → grant via GraphQL, lalu izin dokumen pada dua replika → pengujian ulang alur tiap owner. Perluas [pengujian yang sudah ada](../../../backend/gateway/internal/e2e/document_record_e2e_test.go) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik gateway dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
