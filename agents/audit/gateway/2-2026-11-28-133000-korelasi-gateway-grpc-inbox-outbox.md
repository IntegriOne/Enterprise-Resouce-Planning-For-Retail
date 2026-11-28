# Audit koneksi Gateway — Gateway, gRPC dan korelasi bisnis

**Kesimpulan.** Kelima domain memiliki jalur inisialisasi client dan adapter di bootstrap Gateway. Dua kegagalan lintas owner perlu dibedakan dari kegagalan satu RPC: commit EF yang hasilnya belum diketahui dan pengajuan HR yang belum mendapat tanda terima IAM.

**Lingkup pemeriksaan.** Pembacaan statis checkout bersama pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`, termasuk source lokal yang belum di-commit. Bukti runtime terdahulu disebut terpisah bila tersedia. Tidak dilakukan transaksi HTTPS, transfer, deployment atau perubahan aplikasi pada audit ini.

## Jalur yang terkonfirmasi dari source

`NewRegistry` menyimpan status init per client. Bootstrap memasang adapter EF, Finance, Procurement, HRM dan IAM pada resolver gqlgen serta memasang `WithWorkerPrincipalProvisioner` bila koneksi HRM siap. `CreateEmployeeUser` melewati reservasi, pembuatan user, penautan worker dan penyelesaian IAM. Registry dan circuit breaker mengisolasi kegagalan upstream. Keberadaan field SDL maupun hasil dial non-blocking belum membuktikan otorisasi, commit dan pemrosesan asinkron. [cmd/main.go L71](../../../backend/gateway/cmd/main.go#L71) · [source_connections/registry.go L92](../../../backend/gateway/internal/source_connections/registry.go#L92) · [resolvers/iam_user.resolvers.go](../../../backend/gateway/graph/resolvers/iam_user.resolvers.go) · [iam/user_admin_adapter.go L568](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L568) · [resolvers/schema.resolvers.go](../../../backend/gateway/graph/resolvers/schema.resolvers.go) · [graphql/schema.graphql](../../../backend/gateway/graphql/schema.graphql) · [resolvers/procurement_sourcing_orders.resolvers.go](../../../backend/gateway/graph/resolvers/procurement_sourcing_orders.resolvers.go) · [resolvers/hrm_access_request.resolvers.go](../../../backend/gateway/graph/resolvers/hrm_access_request.resolvers.go) · [resolvers/finance_payroll.resolvers.go](../../../backend/gateway/graph/resolvers/finance_payroll.resolvers.go)

Pada kelima domain, parser SDL gqlparser menemukan **322 Query dan 509 Mutation**. Descriptor Buf menemukan **1.419 RPC** di bawah paket management terkait, termasuk workload privat. Kedua hitungan tersebut memakai unit berbeda. Angka SDL adalah root fields, bukan jumlah variasi finder/filter atau jumlah tes.

| Folder sumber | Query | Mutation | Service proto | RPC |
|---|---:|---:|---:|---:|
| iam | 25 | 48 | 9 | 91 |
| enterprise_foundation | 64 | 48 | 5 | 112 |
| human_resource | 176 | 326 | 24 | 1022 |
| finance_accounting | 11 | 32 | 11 | 73 |
| procurement_sourcing | 46 | 55 | 7 | 121 |

## Peta korelasi lintas management

| Produsen → konsumen | Bukti di checkout | Kekurangan atau batas verifikasi |
|---|---|---|
| Gateway → IAM, EF, HR, Finance, PS | Registry, adapter dan resolver per domain. | GW-01–04. Koneksi diinisialisasi terpisah; profil default Compose HRM dapat dinonaktifkan. |
| IAM → EF dan IAM → Redis | Pull kelayakan target scoped saat assignment. Outbox IAM memiliki tanda effective/retry untuk proyeksi keamanan. | IAM-01. Kontrak memuat target_id, pembaca belum mengikat hasil ke UUID yang diminta. |
| EF ↔ IAM/Gateway untuk vendor/supplier | IAM membuat user pending; EF membuat kontak atau pending_link. EF mencatat perubahan role/portal dalam outbox. | GW-01, EF-02. ACK tautan yang hilang membuat kompensasi ambigu; evaluasi akses portal tetap milik EF. |
| HR → IAM → HR | Pengajuan worker lewat Gateway, receipt IAM, createEmployeeUser dan konfirmasi worker–account. Event terminasi mempunyai rute privat ke IAM. | GW-02, HR-01, IAM-02. Pemulihan otomatis pengajuan awal belum terbukti. |
| HR ↔ IAM, EF, Finance untuk mandat/offboarding | IAM memverifikasi mandat, HR memverifikasi unit EF; owner Finance dan IAM memberi bukti clearance. HR memakai verifier reservasi/release anggaran. | Konfigurasi listener khusus dan dependency owner menentukan kesiapan. Laporan terdahulu membuktikan sebagian jalur pada stack disposable. |
| HR → Finance → HR | Request posting → inbox Finance, posting/accrual → callback. Hasil pembayaran dan settlement kembali ke HR. Event payroll.published menuju pemroses PDF. | FA-04 dan HR-02. Pemicu publikasi Finance dan penyimpanan PDF masih terbuka. Pengujian terdahulu memakai bukti pembayaran fixture. |
| PS ↔ EF, IAM, PLM, Finance | EF memverifikasi partner/role dan Finance mata uang; approval PS dapat memeriksa role/scope IAM. Proto PLM mendefinisikan ValidateSubstitution. | PS-03/04. SKU dan technical approval belum selalu diverifikasi; unit/lokasi tidak tercakup port EF PS saat ini. |
| PS → WMS/Finance | PO issued/revised/cancelled, return dan WO memiliki producer untuk sebagian event. Finance juga menarik PO line dan service acceptance. | PS-01. Rute outbox PS yang berjalan di kode hanya mengambil reversal jasa untuk Finance. Nama/versi dan konsumen tiap event lain harus disepakati. |
| WMS → PS/Finance | Lookup receiving evidence oleh Finance untuk matching. Router WMS mencantumkan receipt/effect. | PS-02, FA-01. Mapping receipt/reversal menuju PS tidak lengkap. Probe valuasi Finance tidak menulis entry. |
| Finance → PS | Invoice posted serta allocation/payment/credit/reversal per PO line punya publisher, inbox dan proyeksi. | FA-02. Relay gabungan dapat menyisihkan event tujuan umum; E2E historis dengan fixture tidak mencakup seluruh konfigurasi. |

Bukti per jalur dan pemilik tindak lanjut. [Audit IAM](../iam/17-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Audit Enterprise Foundation](../enterprise_foundation/1-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Audit HRM](../human_resource/198-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Audit Finance & Accounting](../finance_accounting/8-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Audit Procurement & Sourcing](../procurement_sourcing/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md)

## Kecocokan dengan rancangan bisnis

Pada batas Gateway, procurement §23 memerlukan pemindahan referensi legal entity, item/revisi dan tujuan tanpa kehilangan makna. Inventory §19/22 menuntut bukti receipt dan valuasi. PLM §19–21 memisahkan identitas SKU dari stock dan harga. GW-04 menunjukkan langsung tempat sebagian fakta hilang. Saldo dan pembayaran harus dibaca dari pemilik; resolver tidak menghitung atau mendistribusikan event. [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md)

## Temuan dan dampak

### GW-01 — P1 · Kompensasi provisioning mitra setelah hasil EF ambigu

`createPartyPortalUser` memanggil `rollbackUnlinkedUser` pada setiap error dari `LinkPortalUser`. Fungsi itu mengirim `SoftDeleteUser` dengan konteks yang terpisah dari pembatalan request. EF bisa sudah mengikat kontak ketika responsnya hilang. Pada kondisi tersebut penghapusan lunak membuat akun IAM dan tautan EF bertentangan. Temuan berasal dari urutan source, tanpa klaim telah terjadi pada pengguna nyata. [iam/user_admin_adapter.go L739](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L739) · [enterprise_foundation/portal_provisioning.go L109](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L109) · [party/contact.go L57](../../../backend/services/enterprise_foundation/services/party/contact.go#L57)

### GW-02 — P1 · Pengiriman pengajuan akun HR masih mengikuti umur request HTTP

`RequestEmployeeAccount` menyimpan access request kemudian memanggil IAM. Retry memakai key per worker dan mengecek aktor asli. Pembacaan `readIAMEmployeeAccountStatus` mengembalikan model HR pada NOT_FOUND, dan status verification_unavailable ketika IAM terganggu. Jalur pembacaan tidak mengirim ulang. Jika proses berhenti setelah commit HR, pengguna harus mengulang pengajuan agar IAM menerima melalui jalur ini. [human_resource/access_request.go L20](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L20) · [human_resource/access_request.go L179](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L179) · [accessrequest/service.go L42](../../../backend/services/human_resource/internal/service/accessrequest/service.go#L42) · [provisioning/employee.go](../../../backend/services/iam/services/provisioning/employee.go)

### GW-03 — P2 · Profil transport empat upstream

IAM, EF, Finance dan Procurement menggunakan `newResilientGRPCConn` dengan `insecure.NewCredentials()`; HRM memiliki opsi CA, client certificate dan key tersendiri. HMAC melindungi konteks pemanggil. Kode dial empat jalur tersebut tidak menyajikan client certificate ke listener yang mewajibkan TLS. Audit ini tidak menilai lapisan enkripsi jaringan eksternal. [source_connections/resilience.go L154](../../../backend/gateway/internal/source_connections/resilience.go#L154) · [source_connections/iam_client.go L101](../../../backend/gateway/internal/source_connections/iam_client.go#L101) · [source_connections/ef_client.go L110](../../../backend/gateway/internal/source_connections/ef_client.go#L110) · [source_connections/finance_client.go L124](../../../backend/gateway/internal/source_connections/finance_client.go#L124) · [source_connections/procurement_sourcing_client.go L116](../../../backend/gateway/internal/source_connections/procurement_sourcing_client.go#L116) · [cmd/main.go L88](../../../backend/gateway/cmd/main.go#L88)

### GW-04 — P1 · Field baris PO hilang antara GraphQL dan gRPC

`CreatePOLineInput` di proto memiliki item_revision_reference_id, tax_code_reference_id, delivery_location_reference_id, delivery_party_reference_id, delivery_address_snapshot, cost_center_reference_id dan specification_reference. Ketujuh field absen pada SDL `CreatePurchaseOrderLineInput` dan mapping `CreatePurchaseOrder`. Handler gRPC membaca referensi tersebut. Operator yang membuat PO lewat UI tidak bisa menyatakan revisi, tujuan dan referensi biaya/pajak melalui jalur itu. [procurement_sourcing/procurement_sourcing_orders.graphql](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql) · [procurement_sourcing/orders_adapter.go L12](../../../backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go#L12) · [v1/purchase_order.proto](../../../backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/purchase_order.proto) · [handler/purchase_order_handler.go L25](../../../backend/services/procurement_sourcing/transport/grpc/handler/purchase_order_handler.go#L25) · [purchaseorder/service.go L149](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L149)

## Tindak lanjut

[Instruksi per temuan, pemilik, pengujian dan kriteria penutupan](../../instructions/gateway/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md). Prioritaskan bukti komitmen PO dan pemulihan tautan identitas bersama pemilik domain.

**Jejak baseline lokal.** SHA-256 gabungan 27 file backend/logic yang ditautkan pada audit ini adalah `5e7f242e3ccad636c8d2cc9084428fb9a2010306377f7a8fcb6adbe47292bf9d`. Urutkan path relatif unik, lalu untuk tiap path gabungkan UTF-8 path, byte NUL dan digest SHA-256 biner dari isi file. Hash gabungan membantu mencocokkan checkout ketika source lokal belum dipublikasikan.
