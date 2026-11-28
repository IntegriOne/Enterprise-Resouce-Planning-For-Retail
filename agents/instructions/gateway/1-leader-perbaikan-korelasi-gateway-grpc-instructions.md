# Leader backend — Gateway — Perbaikan korelasi lintas management

- **ID** — 1
- **Owner** — Agent Gateway sebagai pelaksana, leader backend sebagai peninjau. Perubahan di service lain dikoordinasikan dengan pemilik pada matriks.
- **Priority** — Kerjakan temuan P1 lebih dahulu. P2 menjaga kebenaran jalur pada konfigurasi produksi.
- **Status** — Siap dikerjakan.
- **Baseline** — Working tree lokal ERP Retail pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`. Baca ulang source yang dirujuk karena beberapa service masih merupakan perubahan lokal.

## Task

Kelima domain memiliki jalur inisialisasi client dan adapter di bootstrap Gateway. Dua kegagalan lintas owner perlu dibedakan dari kegagalan satu RPC: commit EF yang hasilnya belum diketahui dan pengajuan HR yang belum mendapat tanda terima IAM. Selesaikan tiap temuan melalui entrypoint aplikasi atau workload asli hingga hasilnya dapat dibaca kembali dari pemilik fakta. [Audit Gateway](../../audit/gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Peta korelasi seluruh management](../../audit/gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md)

## Request

### GW-01 — P1 · Kompensasi provisioning mitra setelah hasil EF ambigu

- file: [iam/user_admin_adapter.go L739](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L739) · [enterprise_foundation/portal_provisioning.go L109](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L109) · [party/contact.go L57](../../../backend/services/enterprise_foundation/services/party/contact.go#L57). Path menunjukkan titik baca/perubahan; ubah kode milik Gateway dan sepakati kontrak pasangan dengan owner terkait.
- function: Perbaiki `createPartyPortalUser` dan `rollbackUnlinkedUser` untuk createVendorUser dan createSupplierUser. Gunakan lookup kontak EF menurut IAM user/party dan kontrak replay di EF-02.
- logic: Bedakan penolakan definitif sebelum tulis dari deadline, cancel atau respons hilang setelah kirim. Pada status ambigu, cocokkan party, contact, kind, user dan status owner. Pertahankan identitas permintaan agar pengulangan di replika lain dapat melanjutkan akun pending yang sama. Letakkan catatan pemulihan durable di owner; simpan referensi saja di Gateway.
- error_handling: Jika pembacaan owner juga gagal, balas pending/unavailable sesuai payload dan jadwalkan rekonsiliasi. Kompensasi hanya akun yang terbukti dibuat untuk operasi itu dan terbukti aman dibatalkan.
- security: Recheck izin IAM dan scope party EF saat retry; jangan mengadopsi user hanya karena email sama. Jangan log password atau sesi.
- expected_result: Bila EF commit dan respons hilang, hasil akhir satu user dan satu ikatan yang cocok. Penolakan definitif tetap dapat dikompensasi dan diaudit.
- testing: Uji jalur dengan dan tanpa contact_id, retry ke gateway kedua setelah respons EF hilang, kegagalan lookup, same-key/different-party dan kompensasi gagal; baca kembali kedua owner.

### GW-02 — P1 · Pengiriman pengajuan akun HR masih mengikuti umur request HTTP

- file: [human_resource/access_request.go L20](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L20) · [human_resource/access_request.go L179](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L179) · [accessrequest/service.go L42](../../../backend/services/human_resource/internal/service/accessrequest/service.go#L42) · [provisioning/employee.go](../../../backend/services/iam/services/provisioning/employee.go). Path menunjukkan titik baca/perubahan; ubah kode milik Gateway dan sepakati kontrak pasangan dengan owner terkait.
- function: Selaraskan `RequestEmployeeAccount` dan `readIAMEmployeeAccountStatus` dengan producer HR-01 dan receipt IAM-02.
- logic: Setelah commit HR, tampilkan request ID yang stabil dan status pengiriman yang dapat dibaca. Jika jalur sinkron dipertahankan untuk latensi, pakai identitas bisnis yang sama dengan jalur pemulihan. Tampilkan status linked hanya setelah verifikasi owner.
- error_handling: Bedakan pending delivery, gagal verifikasi dan konflik binding. Jangan membuat akun kedua saat request diulang.
- security: Operator lain hanya boleh membaca/merekonsiliasi sesuai grant. Atribusi aktor asli dipertahankan oleh owner.
- expected_result: Ketika IAM pulih permintaan lama masuk ke antreannya tanpa submit browser kedua. Pengguna dapat melihat bahwa pengajuan sedang menunggu.
- testing: Hentikan alur setelah commit HR sebelum receipt, pulihkan IAM, baca status lewat replika lain. Replay bersamaan harus menghasilkan satu receipt.

### GW-03 — P2 · Profil transport empat upstream

- file: [source_connections/resilience.go L154](../../../backend/gateway/internal/source_connections/resilience.go#L154) · [source_connections/iam_client.go L101](../../../backend/gateway/internal/source_connections/iam_client.go#L101) · [source_connections/ef_client.go L110](../../../backend/gateway/internal/source_connections/ef_client.go#L110) · [source_connections/finance_client.go L124](../../../backend/gateway/internal/source_connections/finance_client.go#L124) · [source_connections/procurement_sourcing_client.go L116](../../../backend/gateway/internal/source_connections/procurement_sourcing_client.go#L116) · [cmd/main.go L88](../../../backend/gateway/cmd/main.go#L88). Path menunjukkan titik baca/perubahan; ubah kode milik Gateway dan sepakati kontrak pasangan dengan owner terkait.
- function: Tambahkan transport credentials per upstream pada `newResilientGRPCConn`, registry dan config. Koordinasikan listener Gateway di tiap owner.
- logic: Lakukan rollout listener dan trust lebih dahulu. Pilih profil TLS terverifikasi per target aktif. Pertahankan deadline, breaker, signing metadata dan isolasi domain opsional pada kedua replika.
- error_handling: Kesalahan CA/SAN atau credential tidak lengkap menolak kanal itu; jangan diam-diam kembali ke plaintext. Batasi plaintext yang masih dibutuhkan pada profil development eksplisit.
- security: Pisahkan sertifikat identitas Gateway dari identitas workload PS/Finance/HR. Listener internal khusus callback tetap privat.
- expected_result: Semua target yang dikonfigurasi TLS menerima identitas Gateway yang tepat. Domain lain tetap bisa digunakan saat satu upstream rusak.
- testing: Uji handshake dan satu query/mutation berizin per domain; uji CA salah, sertifikat workload berbeda, timeout, failover ke replika kedua dan isolasi upstream.

### GW-04 — P1 · Field baris PO hilang antara GraphQL dan gRPC

- file: [procurement_sourcing/procurement_sourcing_orders.graphql](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql) · [procurement_sourcing/orders_adapter.go L12](../../../backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go#L12) · [v1/purchase_order.proto](../../../backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/purchase_order.proto) · [handler/purchase_order_handler.go L25](../../../backend/services/procurement_sourcing/transport/grpc/handler/purchase_order_handler.go#L25) · [purchaseorder/service.go L149](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L149). Path menunjukkan titik baca/perubahan; ubah kode milik Gateway dan sepakati kontrak pasangan dengan owner terkait.
- function: Agent PS mengubah SDL/adapter domain PS di Gateway; agent Gateway mereview model/registrasi bersama. Cocokkan create dan revision line per field.
- logic: Teruskan input operator yang diperlukan; untuk fakta otoritatif yang dihitung owner, derive via service dan expose hasilnya sebagai snapshot read-only. Gunakan pola field opsional existing agar caller lama tetap berfungsi. Tambahkan label SKU/revisi, tujuan, pajak dan cost center untuk UI sesuai grant, gunakan ID hanya untuk referensi.
- error_handling: Tolak target ganda, referensi salah tipe atau beda entity. Jangan menurunkan versi atau mengosongkan field diam-diam.
- security: Kirim hanya fakta/ID terotorisasi; backend memverifikasi ulang PLM, EF dan Finance sesuai PS-04.
- expected_result: PO lewat GraphQL menghasilkan fakta referensi/tujuan yang sama dengan jalur gRPC setara dan bisa dibaca lagi setelah revisi.
- testing: Uji HTTPS create_purchase_order dan create_purchase_order_revision untuk barang/jasa, input minimal lama, tujuan lokasi atau party, cross-entity yang ditolak serta histori revisi lama.

## Action

1. Cocokkan baseline, kode, proto/SDL, status working tree dan kontrak lawan. Sepakati ID operasi/event, owner, scope, waktu, versi, ACK dan bentuk respons gagal sebelum menyalakan jalur baru. Pertahankan kontrak pemanggil lama selama rollout.
2. Kerjakan blok Request sesuai dependency. Tambahkan fixture kecil untuk kasus yang belum terliput. Unit test server gRPC hanya di layer service. Buktikan constraint/efek persisten melalui entrypoint normal dengan database Docker disposable. Untuk Gateway uji resolver/adapter serta HTTPS. Dilarang menulis grant IAM langsung ke DB.
3. Dari `backend/gateway` jalankan `go vet ./...` dan `go test -race -count=1 ./...` pada modul yang berubah. Bila kontrak berubah jalankan generator resmi module; untuk Gateway gunakan `GOWORK=off go generate ./...` di `backend/gateway` dan periksa diff. Uji Caddy di `https://erp.localhost:8443/api/graphql` melalui dua replika. Login superadmin, pilih permission katalog, buat role, pasang permission dan assignment scoped via GraphQL IAM. Gunakan aktor berbeda bila SoD disyaratkan. Cabut grant sementara setelah tes.
4. Untuk jalur workload, cek listener, identitas peer, source event/request, aggregate/effect, versi, correlation ID dan status owner sebelum/sesudah retry. Bedakan receipt pending dari efek diterapkan; lampirkan bukti hasil aktual, termasuk kondisi gagal/tertunda. Hindari PII, voucher, isi slip atau secret pada log.
5. Laporkan per ID file yang diubah, kontrak pasangan, kasus lulus/gagal, rollback/replay dan SHA. Tandai DONE hanya setelah seluruh kriteria aktif terbukti, sesuai aturan repo. Tim yang menunggu dependency dapat meneruskan temuan independen.

## Context

- [agents/arsitektur.md](../../arsitektur.md) · [agents/boundary.md](../../boundary.md) · [agents/users.md](../../users.md) · [agents/how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [agents/how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md) · [agents/instruction_template.md](../../instruction_template.md) · [AGENTS.md](../../../AGENTS.md) · [docs/README.md](../../../docs/README.md)
- [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md) · [Konteks Gateway](../../context/gateway/1-context.md)
- [Instruksi IAM](../iam/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Enterprise Foundation](../enterprise_foundation/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi HRM](../hrm/6-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Finance & Accounting](../finance_accounting/5-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Procurement & Sourcing](../procurement_sourcing/2-leader-perbaikan-korelasi-gateway-grpc-instructions.md)
