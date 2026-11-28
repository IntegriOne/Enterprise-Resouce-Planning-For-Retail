# Audit kode inti hubungan management ERP Retail

**Keputusan audit.** Kode saat ini belum sepenuhnya sesuai dengan logika bisnis. Pemeriksaan mengidentifikasi 16 temuan, termasuk salah ikat pemilik slip gaji, substitusi vendor yang disetujui tanpa verifikasi PLM, valuasi pergerakan stok multibaris, pemilihan event Finance dan notifikasi DRM. [Matriks 27 arah hubungan, sembilan sambungan Gateway dan empat owner yang masih direncanakan](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

**Dasar pemeriksaan.** Source produksi pada working tree bersama di atas `a316fd7`, 28 November 2026 WIB. Sebagian source lokal belum dipublikasikan pemiliknya. Pemeriksaan mengikuti startup, use case, adapter, request/reply RPC, query repository, batas transaksi serta efek setelah commit. Kesimpulan menilai cabang kode dan konsekuensi yang dapat diturunkan darinya. Laporan ini tidak menghitung transaksi pelanggan yang telah terdampak. [Boundary](../../../agents/boundary.md) · [Arsitektur](../../../agents/arsitektur.md) · [IAM](../../../agents/users.md).

**Aturan domain.** [inventory_warehouse](../../../agents/logic/inventory_warehouse/logic.md) · [procurement](../../../agents/logic/procurement/logic.md) · [product_life_cycle](../../../agents/logic/product_life_cycle/logic.md) · [web_commerce](../../../agents/logic/web_commerce/logic.md) · [document_record](../../../agents/logic/document_record/logic.md).

## Perubahan yang sudah terlihat

- Verifikasi party di Commerce/DRM sekarang meminta keputusan EF untuk principal dan party tertentu. HRM kini dapat memasang client DRM dan font pada startup. Kesalahan publikasi slip yang tersisa terjadi pada pemeriksaan Finance. [services/buyer/service.go](../../../backend/services/web_commerce/services/buyer/service.go#L56) · [infrastructure/integration/foundation/client.go](../../../backend/services/web_commerce/infrastructure/integration/foundation/client.go#L118) · [infrastructure/integration/foundation/client.go](../../../backend/services/document_record/infrastructure/integration/foundation/client.go#L130) · [cmd/main.go](../../../backend/services/human_resource/cmd/main.go#L122) · [internal/service/payroll/service.go](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L497).
- DRM kini menyimpan fence sebelum delete dan memeriksa kesesuaian source record pada retention anchor. Recovery setelah crash serta pengiriman event ke owner masih memerlukan perbaikan. [services/records/service.go](../../../backend/services/document_record/services/records/service.go#L695) · [cmd/main.go](../../../backend/services/document_record/cmd/main.go#L87).
- WMS tidak lagi memakai keberhasilan pemeriksaan koneksi Finance sebagai ACK valuasi. Namun format snapshot dan granularitas efeknya belum sesuai dengan skema dan penerima. [internal/service/movement/service.go](../../../backend/services/inventory_warehouse/internal/service/movement/service.go#L435) · [db/migrations/000001_create_inventory_warehouse_schema.up.sql](../../../backend/services/inventory_warehouse/db/migrations/000001_create_inventory_warehouse_schema.up.sql#L703) · [internal/integration/finance/client.go](../../../backend/services/inventory_warehouse/internal/integration/finance/client.go#L90).
- IAM memeriksa kecocokan worker/principal/entity pada provisioning dan target scope EF pada grant. Gateway mengorkestrasi penautan worker ketika membuat employee. [infrastructure/terminationaccess/client.go](../../../backend/services/iam/infrastructure/terminationaccess/client.go#L90) · [infrastructure/scopeverifier/ef_verifier.go](../../../backend/services/iam/infrastructure/scopeverifier/ef_verifier.go#L235) · [adapter/iam/user_admin_adapter.go](../../../backend/gateway/internal/adapter/iam/user_admin_adapter.go#L568).

## Temuan

P1 berarti cabang kode dapat menghasilkan keputusan transaksi atau akses yang salah, menggagalkan transaksi sah, atau kehilangan delivery wajib. P2 memerlukan pengetatan konfigurasi atau kebijakan sebelum aktivasi alur terkait.

### KI-01 · P1 · Posting stok dapat gagal sebelum commit

**Pemilik** — WMS.

`PostMovement` memasukkan seluruh JSON ke `payload_reference`. Kolom `warehouse_outbox_events.payload_reference` hanya menyediakan 300 karakter. Header, UUID baris, SKU dan referensi PO memakai kolom yang sama. Begitu representasi melampaui 300 karakter, insert outbox menggagalkan transaksi stok yang sah.

**Hasil perbaikan yang diperlukan** — Simpan snapshot efek di payload berkapasitas sesuai skema atau tabel immutable dengan referensi pendek dan digest. Pertahankan atomisitas ledger dan outbox, serta resolusi event format lama.

[internal/service/movement/service.go](../../../backend/services/inventory_warehouse/internal/service/movement/service.go#L435) · [internal/service/internal/servicekit/helpers.go](../../../backend/services/inventory_warehouse/internal/service/internal/servicekit/helpers.go#L95) · [repository/integration/repository.go](../../../backend/services/inventory_warehouse/repository/integration/repository.go#L25) · [db/migrations/000001_create_inventory_warehouse_schema.up.sql](../../../backend/services/inventory_warehouse/db/migrations/000001_create_inventory_warehouse_schema.up.sql#L703).

### KI-02 · P1 · Satu event mencampur data beberapa efek stok

**Pemilik** — WMS · Finance.

`postedLines[0]` menentukan item dan jumlah, sedangkan loop seluruh `postedEffects` menimpa referensi PO/receiving sebelumnya. Producer menjadikan nama event sebagai `movement_kind` dan tidak mengisi `unit_cost`. Client Finance meminta biaya untuk efek kepemilikan dan memakai jumlah baris alih-alih delta kepemilikan bertanda. Finance memakai jenis seperti `receipt` dan `issue`. Expense/receipt multibaris dapat tertahan atau terproyeksi pada baris yang keliru.

Constraint unik outbox saat ini memakai agregat, versi agregat dan jenis event. Saat menerbitkan beberapa efek bertipe sama, skema identitas event juga perlu menyertakan pembeda efek.

**Hasil perbaikan yang diperlukan** — Bentuk event per efek bisnis dengan relasi baris yang sama, key stabil pada replay, delta bertanda, kategori Finance, dan sumber basis valuasi yang disepakati. Catat penerimaan per tujuan. Pertahankan custody tanpa perubahan kepemilikan sebagai efek fisik.

[internal/service/movement/service.go](../../../backend/services/inventory_warehouse/internal/service/movement/service.go#L435) · [internal/integration/finance/client.go](../../../backend/services/inventory_warehouse/internal/integration/finance/client.go#L90) · [domain/entity/enums.go](../../../backend/services/finance_accounting/domain/entity/enums.go#L263) · [internal/service/inventory_cost/service.go](../../../backend/services/finance_accounting/internal/service/inventory_cost/service.go#L63).

### KI-03 · P1 · Reversal dan dispatch putus di batas WMS–Procurement

**Pemilik** — WMS · Procurement.

WMS dapat menyalin `reverses_effect_id` lokal ke `OriginalReceiptLinkId`. Procurement memakai nilai itu sebagai PK `PurchaseReceiptLink` miliknya. WMS mengabaikan isi respons `ApplyAcceptedReceipt`. `ReturnDispatched` dan `StockIssued` diproduksi outbound tetapi tidak tercantum di `DefaultRoutes`. Router mengembalikan sukses saat tujuan kosong, juga saat satu target sukses dan target lain belum terpasang.

**Hasil perbaikan yang diperlukan** — Simpan pemetaan durable dari effect WMS ke receipt link Procurement atau lakukan lookup melalui source effect. Daftarkan tipe dispatch yang relevan, bedakan target wajib dan efek lokal, pastikan retry tidak menggandakan receipt, reversal, outstanding maupun valuasi.

[internal/integration/procurement/client.go](../../../backend/services/inventory_warehouse/internal/integration/procurement/client.go#L189) · [internal/service/fulfillment/service.go](../../../backend/services/procurement_sourcing/internal/service/fulfillment/service.go#L145) · [internal/service/outbound/service.go](../../../backend/services/inventory_warehouse/internal/service/outbound/service.go#L969) · [internal/integration/outbox/router.go](../../../backend/services/inventory_warehouse/internal/integration/outbox/router.go#L53) · [internal/service/integration/service.go](../../../backend/services/inventory_warehouse/internal/service/integration/service.go#L211).

### KI-04 · P1 · Rencana inbound PO memakai otorisasi dari input

**Pemilik** — WMS · Procurement.

`CreateInboundOrder` menyalin sumber, versi, referensi dan kuantitas yang diberikan caller. `ApplyReceiptDisposition` membatasi acceptance terhadap `ExpectedBaseQuantity` lokal. Client menyediakan `ValidatePOAuthorization` tetapi startup menggunakan client Procurement sebagai target outbox saja. Batas kuantitas lokal bukan bukti versi PO yang sudah diterbitkan.

**Hasil perbaikan yang diperlukan** — Untuk inbound berbasis PO, dapatkan otorisasi baris dari Procurement sebelum release/accept. Cocokkan entity, SKU, UOM, revisi, status dan saldo yang diizinkan. Bedakan inbound non-PO; bila PO direvisi, rekonsiliasi kuantitas yang sudah diterima.

[internal/service/receiving/service.go](../../../backend/services/inventory_warehouse/internal/service/receiving/service.go#L94) · [cmd/main.go](../../../backend/services/inventory_warehouse/cmd/main.go#L77) · [internal/integration/procurement/client.go](../../../backend/services/inventory_warehouse/internal/integration/procurement/client.go#L189) · [internal/service/purchaseorder/service.go](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L672).

### KI-05 · P1 · Valuasi Finance menerima perhitungan yang kehilangan presisi

**Pemilik** — Finance.

Handler `ApplyInventoryCostEvent` mengabaikan dua error `ParseAmount`, lalu menghitung nilai dengan `StringFixed(2)`. Domain menyediakan `ParseQuantity` dan `ParseUnitPrice` dengan skala berbeda serta menyimpan nilai berskala enam. Harga satuan valid dengan tujuh atau delapan desimal dapat gagal diparse dan memengaruhi perhitungan. Service baru menerima `ValueDelta` yang telah dihitung.

**Hasil perbaikan yang diperlukan** — Validasi jenis mutasi, kuantitas dan harga satuan sebelum perkalian. Gunakan skala buku atau kebijakan mata uang yang disahkan saat pembulatan, simpan tanda delta dan alasan reversal, jangan komit jumlah tanpa nilai yang semestinya.

[transport/grpc/handler/inventory_cost_handler.go](../../../backend/services/finance_accounting/transport/grpc/handler/inventory_cost_handler.go#L34) · [domain/entity/decimal.go](../../../backend/services/finance_accounting/domain/entity/decimal.go#L39) · [internal/service/inventory_cost/service.go](../../../backend/services/finance_accounting/internal/service/inventory_cost/service.go#L63) · [domain/entity/enums.go](../../../backend/services/finance_accounting/domain/entity/enums.go#L263).

### KI-06 · P1 · Arsip SKU dan ketersediaan memakai hasil daftar yang tidak lengkap

**Pemilik** — PLM · Procurement · WMS · Commerce.

PLM mengecek 100 entri pertama per status pada PO/WO dan membaca nested lines. `SelectPurchaseOrders` menghasilkan header tanpa lines. PLM dan Commerce juga mengambil 100 baris pertama ringkasan stok seluruh gudang lalu menyaring SKU di memori. PLM memperlakukan WMS `NotFound` sebagai residual kosong, dan Commerce mengembalikan `Known=true` setelah satu halaman. SKU yang tidak terlihat dapat dianggap eligible untuk arsip atau kehabisan stok.

**Hasil perbaikan yang diperlukan** — Sediakan query fakta per SKU pada owner, termasuk scope channel/entity, komitmen aktif dan dimensi custody/transit. Bila memakai pagination, konsumsi seluruh halaman dan detail baris dengan snapshot konsisten. Keadaan tidak diketahui harus menghalangi arsip dan tidak boleh menjadi janji availability.

[services/changecycle/service.go](../../../backend/services/product_lifecycle/services/changecycle/service.go#L387) · [infrastructure/integration/procurement/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/procurement/client.go#L63) · [repository/purchase_order/purchase_order_repository.go](../../../backend/services/procurement_sourcing/repository/purchase_order/purchase_order_repository.go#L218) · [repository/external_production/external_production_repository.go](../../../backend/services/procurement_sourcing/repository/external_production/external_production_repository.go#L448) · [infrastructure/integration/inventory/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/inventory/client.go#L60) · [infrastructure/integration/wms/client.go](../../../backend/services/web_commerce/infrastructure/integration/wms/client.go#L55) · [transport/grpc/handler/transfer_count_prod.go](../../../backend/services/inventory_warehouse/transport/grpc/handler/transfer_count_prod.go#L799).

### KI-07 · P1 · Fan-out perubahan produk dapat selesai setelah satu penerima

**Pemilik** — PLM · Commerce.

`ProductRevisionReleased` menargetkan WMS, Procurement, Finance, Commerce. Startup PLM memasang publisher Commerce. Router melewati target tanpa publisher; satu sukses cukup. Di Commerce, `DeliverInboxEvent` mengenali subset tipe termasuk revisi dan UOM; event lain seperti `ProductIdentifierChanged` dapat dicatat processed tanpa invalidasi. Startup Commerce memasang router outbound tanpa target.

**Hasil perbaikan yang diperlukan** — Pisahkan status per pasangan event dan penerima. Simpan efek invalidasi SKU/barcode yang sesuai, tandai tipe tak dikenal untuk rekonsiliasi, serta sertakan versi/digest dan identitas item. Event Commerce yang diwajibkan downstream harus memiliki consumer atau tetap tertunda.

[infrastructure/outbox/router.go](../../../backend/services/product_lifecycle/infrastructure/outbox/router.go#L42) · [cmd/wiring.go](../../../backend/services/product_lifecycle/cmd/wiring.go#L85) · [infrastructure/integration/commerce/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/commerce/client.go#L91) · [services/integration/service.go](../../../backend/services/web_commerce/services/integration/service.go#L161) · [infrastructure/outbox/router.go](../../../backend/services/web_commerce/infrastructure/outbox/router.go#L36) · [cmd/wiring.go](../../../backend/services/web_commerce/cmd/wiring.go#L52).

### KI-08 · P1 · Listing dapat terbit tanpa konfirmasi PLM

**Pemilik** — Commerce · PLM.

Wiring Commerce hanya memasang `ports.PLM` saat `PLMEnforcePublish` aktif. `PublishStorefrontListing` memanggil `ResolveSellableItem` hanya jika port tersedia, lalu tetap bisa mengubah status menjadi published. Memiliki alamat gRPC di konfigurasi tidak menjamin SKU masih boleh dijual.

**Hasil perbaikan yang diperlukan** — Izinkan draft sesuai aturan authoring. Untuk publish dan republish, wajibkan fakta sellable per SKU/revisi dari PLM; jika sumber tidak tersedia, pertahankan status sebelumnya. Terapkan juga ketika perubahan lifecycle membatalkan listing lama.

[services/catalog/service.go](../../../backend/services/web_commerce/services/catalog/service.go#L253) · [cmd/wiring.go](../../../backend/services/web_commerce/cmd/wiring.go#L52) · [services/revision/service.go](../../../backend/services/product_lifecycle/services/revision/service.go#L672) · [](../../../backend/docker-compose.yml#L332).

### KI-09 · P1 · Finance meminta pemilik dokumen yang keliru untuk slip

**Pemilik** — Finance · HRM · DRM.

`StorePrivatePayslip` mendaftarkan `human_resource`, tipe `payroll_result` dan result ID. `PublishPayroll` memanggil verifier generik Finance yang membentuk `expected_owner_management=finance_accounting` dan tidak mengirim result ID. DRM memeriksa owner persis. Jadi slip sah dari jalur HRM gagal pada publikasi per item walaupun digest benar.

**Hasil perbaikan yang diperlukan** — Gunakan verifier dengan pemilik HRM, result ID, tipe, entity dan digest. Bedakan kegagalan kontrak dari gangguan jaringan. Finance mengesahkan run; HRM membuat PDF; DRM memastikan versinya; Finance mencatat publikasi item. Pertahankan akses pegawai aktif dan jalur HR untuk mantan pegawai.

[infrastructure/integration/document_records/drm_client.go](../../../backend/services/human_resource/infrastructure/integration/document_records/drm_client.go#L170) · [internal/service/payroll/service.go](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L497) · [internal/integration/drm/client.go](../../../backend/services/finance_accounting/internal/integration/drm/client.go#L73) · [services/document/service.go](../../../backend/services/document_record/services/document/service.go#L386) · [internal/service/payroll/payslip_reissue.go](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L27).

### KI-10 · P1 · Relay privat Finance dapat melewatkan seluruh event Procurement

**Pemilik** — Finance · Procurement.

Dalam mode standby, saat target HRM tersedia, `buildOutboxRelay` lebih dulu memilih `ForHRMPayrollDelivery` dan mengisi publisher. Ketika publisher Procurement juga ada, cabang untuk mengganti query menjadi gabungan hanya berjalan bila publisher masih nil. Akibatnya route AP terpasang tetapi `SelectPendingBatchForUpdate` hanya mengambil event payroll.

**Hasil perbaikan yang diperlukan** — Pilih himpunan event yang dapat diklaim sesudah seluruh tujuan diketahui. Pada konfigurasi privat HRM + Procurement, claim harus meliputi posted invoice, payment allocation/reversal, credit dan payroll yang didukung. Pastikan event payroll privat tetap terisolasi dari webhook umum.

[cmd/wiring.go](../../../backend/services/finance_accounting/cmd/wiring.go#L359) · [repository/integration/integration_repository.go](../../../backend/services/finance_accounting/repository/integration/integration_repository.go#L220) · [internal/integration/procurement/client.go](../../../backend/services/finance_accounting/internal/integration/procurement/client.go#L69) · [internal/service/integration/service.go](../../../backend/services/procurement_sourcing/internal/service/integration/service.go#L500).

### KI-11 · P1 · DRM menandai event terkirim ketika hanya menulis log

**Pemilik** — DRM.

`runOutboxWorker` mengambil baris pending, menulis log, mengubah status ke `EvtProcessed` dan mengisi `PublishedAt`. Tidak ada panggilan consumer pada jalur ini. Owner dokumen dapat kehilangan pemberitahuan revoke, legal hold atau disposition walaupun baris berstatus selesai.

**Hasil perbaikan yang diperlukan** — Daftarkan tujuan sesuai owner dan tipe event. Simpan tanda terima inbox durable per tujuan, jadwalkan ulang pending dan tangani replay. Rekonsiliasi baris historis dengan tanda terima penerima sebelum backfill.

[cmd/main.go](../../../backend/services/document_record/cmd/main.go#L87) · [services/publication/service.go](../../../backend/services/document_record/services/publication/service.go#L445) · [services/records/service.go](../../../backend/services/document_record/services/records/service.go#L695).

### KI-12 · P1 · Kontrol akses dokumen berbeda antara daftar dan isi

**Pemilik** — Gateway · DRM.

`ListMyPublications` memakai query recipient `kind='user'`. Pembacaan byte memiliki jalur tambahan untuk party dengan konfirmasi EF. Untuk staf, flag dari grant global Gateway dapat melewati pemeriksaan recipient; pengecualian khusus dalam service hanya berlaku pada dokumen HR sensitif untuk aktor non-system. Dokumen Finance/Procurement yang sensitif belum mendapat pemeriksaan owner yang setara pada jalur tersebut.

**Hasil perbaikan yang diperlukan** — Definisikan satu keputusan akses per version, owner, entity, tujuan, principal dan klasifikasi. Gunakan keputusan itu untuk metadata, daftar, byte dan reissue. Untuk party, evaluasi PIC aktif di EF; untuk workload, ikat identitas pemanggil pada tujuan dan resource.

[adapter/document_record/ingestion_pub.go](../../../backend/gateway/internal/adapter/document_record/ingestion_pub.go#L234) · [services/publication/service.go](../../../backend/services/document_record/services/publication/service.go#L445) · [repository/publication/repo.go](../../../backend/services/document_record/repository/publication/repo.go#L182) · [infrastructure/integration/foundation/client.go](../../../backend/services/document_record/infrastructure/integration/foundation/client.go#L130) · [services/party/contact.go](../../../backend/services/enterprise_foundation/services/party/contact.go#L96).

### KI-13 · P1 · Disposisi DRM masih dapat tersangkut setelah restart

**Pemilik** — DRM.

`ApproveDisposition` kini mengommit fence dan job sebelum delete. Penghapusan dan finalisasi masih berjalan dalam satu permintaan pemanggil. Persetujuan ulang menerima status requested saja, sedangkan worker startup yang terlihat membaca outbox, bukan meneruskan storage job processing. Setelah proses berhenti di antara dua fase, file dan metadata memerlukan rekonsiliasi eksplisit.

**Hasil perbaikan yang diperlukan** — Jalankan worker storage yang dapat mengambil kembali job menurut lease, recheck hold/clearance, periksa generasi objek persis dan simpan bukti delete sebelum tombstone. Pertahankan blok akses/restore selama hasil belum diketahui.

[services/records/service.go](../../../backend/services/document_record/services/records/service.go#L695) · [cmd/main.go](../../../backend/services/document_record/cmd/main.go#L87).

### KI-14 · P2 · Validasi referensi EF bergantung pada wiring opsional

**Pemilik** — WMS · PLM · Commerce.

Contoh pada WMS dan PLM, validasi entity/party/lokasi hanya berjalan bila port disuntikkan. Profil Compose menyediakan opsi menonaktifkan pemeriksaan tersebut. Pada WMS, `ValidateFacility` menerima ID fasilitas saja. Commerce juga mengaktifkan sales channel dengan entity dan lokasi dari input setelah hanya memeriksa lokasi aktif. Lokasi aktif milik entity lain dapat lolos pemeriksaan tersebut. [Validasi fasilitas WMS](../../../backend/services/inventory_warehouse/internal/integration/foundation/client.go#L112) · [Pembuatan sales channel](../../../backend/services/web_commerce/services/storefront/service.go#L179).

**Hasil perbaikan yang diperlukan** — Bedakan draft lokal dari transisi yang memerlukan fakta EF. Pada transisi itu wajibkan owner benar, entity serta status saat efektif. Kembalikan unavailable saat keputusan belum tersedia; role party global perlu dilengkapi scope dari pemilik transaksi.

[cmd/main.go](../../../backend/services/inventory_warehouse/cmd/main.go#L77) · [internal/service/location/service.go](../../../backend/services/inventory_warehouse/internal/service/location/service.go#L31) · [cmd/wiring.go](../../../backend/services/product_lifecycle/cmd/wiring.go#L85) · [services/masterdata/service.go](../../../backend/services/product_lifecycle/services/masterdata/service.go#L85) · [](../../../backend/docker-compose.yml#L332).

### KI-15 · P2 · Bukti clean DRM bergantung pada pemeriksaan berkas lokal

**Pemilik** — DRM.

Startup DRM memasang scanner lokal. Engine memeriksa MIME, format executable dan indikator statis. Format yang dikenali dan lolos pemeriksaan dapat diteruskan ke seal. Keputusan ini perlu dibatasi menurut kemampuan scanner dan kebijakan publikasi untuk dokumen dari pihak eksternal.

**Hasil perbaikan yang diperlukan** — Simpan identitas engine, versi kebijakan dan cakupan format bersama digest. Untuk format atau status scan yang belum dapat dipastikan, tahan seal/publikasi sesuai aturan kelas dokumen. Rekonsiliasi hasil saat engine kembali tersedia.

[infrastructure/scanner/engine.go](../../../backend/services/document_record/infrastructure/scanner/engine.go#L66) · [cmd/wiring.go](../../../backend/services/document_record/cmd/wiring.go#L40).

### KI-16 · P1 · Persetujuan substitusi vendor tidak meminta keputusan teknis PLM

**Pemilik** — Procurement · PLM.

`ApproveMaterialSubstitution` menegakkan pemisahan pengaju dan approver serta keberadaan UUID approval. Fungsi kemudian menyimpan `approved` tanpa memanggil `ValidateSubstitution`. Padahal client Procurement dan handler PLM memiliki kontrak yang mencocokkan approval, WO, material line, kuantitas dan digest.

**Hasil perbaikan yang diperlukan** — Bentuk snapshot dari data WO/material di Procurement. Panggil verifier PLM sebelum keputusan final dan pastikan snapshot tetap sama saat commit. Approval tidak sah atau kedaluwarsa menghentikan transisi. Simpan bukti keputusan PLM dan pertahankan SoD komersial.

[internal/service/vendorcontrol/service.go](../../../backend/services/procurement_sourcing/internal/service/vendorcontrol/service.go#L457) · [internal/integration/plm/client.go](../../../backend/services/procurement_sourcing/internal/integration/plm/client.go#L95) · [services/changecycle/service.go](../../../backend/services/product_lifecycle/services/changecycle/service.go#L387) · [transport/grpc/handler/revision_structure_change.go](../../../backend/services/product_lifecycle/transport/grpc/handler/revision_structure_change.go#L934).

## Instruksi per agent

- [WMS](../../../agents/instructions/inventory_warehouse/3-leader-remediasi-code-inti-inventory-warehouse-instructions.md) — KI-01, KI-02, KI-03, KI-04, KI-06, KI-14.
- [Procurement](../../../agents/instructions/procurement_sourcing/4-leader-remediasi-code-inti-procurement-sourcing-instructions.md) — KI-03, KI-04, KI-06, KI-10, KI-16.
- [Finance](../../../agents/instructions/finance_accounting/7-leader-remediasi-code-inti-finance-accounting-instructions.md) — KI-02, KI-05, KI-09, KI-10.
- [PLM](../../../agents/instructions/product_lifecycle/3-leader-remediasi-code-inti-product-lifecycle-instructions.md) — KI-06, KI-07, KI-08, KI-14, KI-16.
- [Commerce](../../../agents/instructions/web_commerce/3-leader-remediasi-code-inti-web-commerce-instructions.md) — KI-06, KI-07, KI-08, KI-14.
- [DRM](../../../agents/instructions/document_record/3-leader-remediasi-code-inti-document-record-instructions.md) — KI-09, KI-11, KI-12, KI-13, KI-15.
- [HRM](../../../agents/instructions/hrm/8-leader-remediasi-code-inti-hrm-instructions.md) — KI-09.
- [Gateway](../../../agents/instructions/gateway/3-leader-remediasi-code-inti-gateway-instructions.md) — KI-12.

Mulai dari posting WMS, persetujuan substitusi, publikasi payroll dan pengiriman notifikasi dokumen. Lanjutkan dengan agregasi per SKU, penerbitan listing, akses dokumen dan recovery. Untuk kontrak yang dipakai dua management, tutup temuan sesudah kedua sisi mengevaluasi identitas dan status bisnis yang sama.
