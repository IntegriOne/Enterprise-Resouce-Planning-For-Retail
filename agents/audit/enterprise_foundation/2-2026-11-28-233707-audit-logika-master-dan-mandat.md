# Audit logika Enterprise Foundation

**Keputusan.** Prioritaskan aktivasi dan pencabutan akses PIC, integritas status dan hierarki, serta validitas kepala unit. Delapan temuan terbuka di bawah dapat dikerjakan memakai titik masuk yang sudah tersedia. Perubahan yang memerlukan kontrak privat melibatkan IAM, Procurement atau HRM sesuai kepemilikan data.

**Baseline dan metode.** Pemeriksaan statis pada working tree 28 November 2026, HEAD `c49b9e1119409aa38d58bf9727e37d8360c7944e`. Sumber usulan lokal adalah `agents/logic/enterprise_foundation/logic.md`. Cakupan mencakup lima kelompok master EF, kontrak gRPC, resolver dan adapter Gateway, jalur HRM dan penerbitan event terkait. Seluruh 112 root EF dicocokkan dari AST SDL ke deklarasi RPC serta pemanggilan dalam resolver dan adapter produksi. Fungsi dan DDL yang menentukan temuan dibaca langsung, termasuk rentang SQL yang hanya terindeks parsial. Temuan konkurensi menjelaskan urutan commit yang diizinkan kode. Tidak ada pengujian runtime atau audit terhadap berkas tes pada pekerjaan ini. [Matriks per operasi](../../../agents/grpc/enterprise_foundation/1-operasi-master-organisasi-dan-mandat.md).

| Prioritas | ID | Dampak utama |
| --- | --- | --- |
| Tinggi | EFL-01 | Pembuatan kontak dapat mengaktifkan portal melalui pemeriksaan yang lebih ringan |
| Tinggi | EFL-02 | Izin mewakili mitra belum dibedakan per role bisnis |
| Tinggi | EFL-03 | Event pelepasan kontak kehilangan identitas principal sebelumnya |
| Tinggi | EFL-04 | Penyimpanan profil dapat menimpa status dan pilihan PIC utama |
| Tinggi | EFL-05 | Pemindahan hierarki belum menjamin graf bebas siklus dan batas kedalaman seluruh cabang |
| Tinggi | EFL-06 | Referensi kepala unit lolos validasi format tanpa verifikasi penugasan HRM |
| Sedang | EFL-07 | Find-or-create alamat tidak atomik dan melakukan pencarian paginasi |
| Sedang | EFL-08 | Pendaftaran party menerima status aktif langsung dari pemanggil |

## EFL-01 — Pembuatan kontak dapat mengaktifkan portal melalui pemeriksaan yang lebih ringan

Gateway meneruskan status portal dari addPartyContact dengan izin tulis berscope tanpa pemeriksaan sensitif. activatePartyPortalAccess memakai pemeriksaan sensitif dengan grant global. Handler dan layanan pembuatan menerima ACTIVE untuk party layak dan principal terisi. Dua pintu masuk menghasilkan status bisnis yang sama dengan persyaratan berbeda.

**Bukti produksi.** [adapter/enterprise_foundation/party_sub_adapter.go L611](../../../backend/gateway/internal/adapter/enterprise_foundation/party_sub_adapter.go#L611) · [adapter/enterprise_foundation/party_sub_adapter.go L749](../../../backend/gateway/internal/adapter/enterprise_foundation/party_sub_adapter.go#L749) · [adapter/enterprise_foundation/permissions.go L46](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions.go#L46) · [grpc/handler/party_contact_handler.go L139](../../../backend/services/enterprise_foundation/transport/grpc/handler/party_contact_handler.go#L139) · [services/party/contact.go L135](../../../backend/services/enterprise_foundation/services/party/contact.go#L135) · [adapter/enterprise_foundation/portal_provisioning.go L161](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L161).

**Kriteria penutupan.** Kontak baru tidak dapat efektif melalui operasi profil. Provisioning supplier/vendor tetap dapat menyelesaikan alur pending ke aktif sekali.

## EFL-02 — Izin mewakili mitra belum dibedakan per role bisnis

EvaluateRepresentativeAccess mengecek peran aktif milik party dan status portal kontak. Penyimpanan mengikat party dengan principal, tanpa role pada mandat PIC. Pada mitra yang memegang dua peran, bukti kelayakan EF dapat mencakup peran yang belum diberikan kepada PIC. Akses akhir tetap melibatkan IAM dan pemeriksaan pemilik transaksi.

**Bukti produksi.** [domain/service/portal_policy.go L29](../../../backend/services/enterprise_foundation/domain/service/portal_policy.go#L29) · [services/party/contact.go L96](../../../backend/services/enterprise_foundation/services/party/contact.go#L96) · [db/migrations/000001_create_enterprise_foundation_schema.up.sql L293](../../../backend/services/enterprise_foundation/db/migrations/000001_create_enterprise_foundation_schema.up.sql#L293) · [adapter/enterprise_foundation/portal_provisioning.go L76](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L76) · [enterprise_foundation/v1/party.proto L651](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/party.proto#L651).

**Kriteria penutupan.** PIC khusus supplier hanya lolos sebagai supplier pada party dua peran. Revokasi satu mandat tidak mengubah mandat lain yang sah.

## EFL-03 — Event pelepasan kontak kehilangan identitas principal sebelumnya

UnlinkPortalUser mengosongkan IAMUserID. Outbox dibangun sesudah mutasi dari kontak terbaru, sehingga event pelepasan hanya menyebut party/contact dan status akhir tanpa principal terdahulu. Konsumen yang harus membatalkan mandat terdahulu memerlukan pemetaan tambahan. Relay menandai effective sesudah Publish berhasil. Itu membuktikan keberhasilan menurut publisher yang dikonfigurasi.

**Bukti produksi.** [domain/entity/partycontact.go L243](../../../backend/services/enterprise_foundation/domain/entity/partycontact.go#L243) · [services/party/contact.go L383](../../../backend/services/enterprise_foundation/services/party/contact.go#L383) · [services/party/contact.go L439](../../../backend/services/enterprise_foundation/services/party/contact.go#L439) · [services/party/contact.go L517](../../../backend/services/enterprise_foundation/services/party/contact.go#L517) · [domain/service/outbox_builder.go L26](../../../backend/services/enterprise_foundation/domain/service/outbox_builder.go#L26) · [infrastructure/outbox/relay.go L376](../../../backend/services/enterprise_foundation/infrastructure/outbox/relay.go#L376).

**Kriteria penutupan.** Setelah pelepasan, IAM dapat mengidentifikasi hak lama secara deterministik, termasuk setelah kontak ditautkan ke orang lain. Event duplikat tidak merusak tautan baru.

## EFL-04 — Penyimpanan profil dapat menimpa status dan pilihan PIC utama

Layanan profil membaca versi awal party atau badan hukum. Update repository menulis kembali status bersama profil, termasuk penanda primary pada badan hukum. Transaksi yang tumpang tindih dengan suspend/dissolve berisiko menulis kembali nilai lama. Pada kontak, pengosongan primary lalu insert/update belum dijaga indeks unik per party. Hal ini melengkapi IEP-01 untuk persaingan profil kontak dengan pencabutan portal.

**Bukti produksi.** [services/party/service.go L215](../../../backend/services/enterprise_foundation/services/party/service.go#L215) · [repository/party/party_repository.go L246](../../../backend/services/enterprise_foundation/repository/party/party_repository.go#L246) · [services/legalentity/service.go L244](../../../backend/services/enterprise_foundation/services/legalentity/service.go#L244) · [repository/legalentity/legalentity_repository.go L331](../../../backend/services/enterprise_foundation/repository/legalentity/legalentity_repository.go#L331) · [services/party/contact.go L195](../../../backend/services/enterprise_foundation/services/party/contact.go#L195) · [repository/party/party_contact_repository.go L273](../../../backend/services/enterprise_foundation/repository/party/party_contact_repository.go#L273) · [db/migrations/000001_create_enterprise_foundation_schema.up.sql L321](../../../backend/services/enterprise_foundation/db/migrations/000001_create_enterprise_foundation_schema.up.sql#L321).

**Kriteria penutupan.** Penyimpanan profil berbarengan dengan suspensi menghasilkan profil baru dan status suspended, atau satu konflik eksplisit. Setiap party mempunyai paling banyak satu PIC utama aktif.

## EFL-05 — Pemindahan hierarki belum menjamin graf bebas siklus dan batas kedalaman seluruh cabang

ValidateParentChain berjalan ke leluhur yang diusulkan. Penempatan ulang tidak menghitung tinggi keturunan yang ikut pindah. Unit mengunci baris unit sendiri ketika diubah, tetapi parent/leluhur dibaca tersendiri. Dua pemindahan bersamaan pada baris berbeda dapat lulus berdasarkan struktur lama. Validasi parent tunggal yang juga dipakai lokasi dan badan hukum perlu invariant yang sama.

**Bukti produksi.** [domain/service/hierarchy_policy.go L44](../../../backend/services/enterprise_foundation/domain/service/hierarchy_policy.go#L44) · [services/organizationunit/service.go L253](../../../backend/services/enterprise_foundation/services/organizationunit/service.go#L253) · [services/organizationunit/service.go L460](../../../backend/services/enterprise_foundation/services/organizationunit/service.go#L460) · [services/location/service.go L430](../../../backend/services/enterprise_foundation/services/location/service.go#L430) · [services/legalentity/service.go L504](../../../backend/services/enterprise_foundation/services/legalentity/service.go#L504) · [repository/pgstore/tx.go L78](../../../backend/services/enterprise_foundation/repository/pgstore/tx.go#L78) · [repository/organizationunit/organizationunit_repository.go L208](../../../backend/services/enterprise_foundation/repository/organizationunit/organizationunit_repository.go#L208).

**Kriteria penutupan.** Seluruh cabang tetap di dalam batas kedalaman. Dua perpindahan yang saling bertentangan tidak dapat sama-sama commit membentuk siklus.

## EFL-06 — Referensi kepala unit lolos validasi format tanpa verifikasi penugasan HRM

Gateway menerima typed ID Worker, dan EF menyimpan head_person_id pada create, update atau assign. DBML mendefinisikannya sebagai worker_id HRM. Jalur-jalur penulisan tersebut tidak memverifikasi keberadaan, masa aktif atau kesesuaian legal entity penugasan di HRM. Kepala unit dapat menjadi referensi keputusan organisasi di management lain.

**Bukti produksi.** [adapter/enterprise_foundation/organization_unit_adapter.go L338](../../../backend/gateway/internal/adapter/enterprise_foundation/organization_unit_adapter.go#L338) · [services/organizationunit/service.go L192](../../../backend/services/enterprise_foundation/services/organizationunit/service.go#L192) · [services/organizationunit/service.go L253](../../../backend/services/enterprise_foundation/services/organizationunit/service.go#L253) · [services/organizationunit/service.go L311](../../../backend/services/enterprise_foundation/services/organizationunit/service.go#L311) · [enterprise_foundation/services/registry.go L1](../../../backend/services/enterprise_foundation/services/registry.go#L1) · [db/schema/enterprise_foundation.dbml L195](../../../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml#L195) · [integration/enterprise_foundation/client.go L198](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L198).

**Kriteria penutupan.** Tiga jalur menghasilkan keputusan sama. Kepala yang sah dapat dipasang, assignment usang atau salah perusahaan ditolak, edit nama dengan preserve tetap bekerja.

## EFL-07 — Find-or-create alamat tidak atomik dan melakukan pencarian paginasi

Create dan FindOrCreate mencari duplikat sebelum Insert terpisah. Pencarian membaca halaman kandidat dan membandingkan alamat di Go. Tabel tidak memiliki constraint keunikan untuk kesetaraan itu. Pada permintaan paralel, pemanggil dapat menerima dua ID untuk alamat yang setara.

**Bukti produksi.** [services/address/service.go L99](../../../backend/services/enterprise_foundation/services/address/service.go#L99) · [services/address/service.go L129](../../../backend/services/enterprise_foundation/services/address/service.go#L129) · [services/address/service.go L232](../../../backend/services/enterprise_foundation/services/address/service.go#L232) · [services/address/service.go L281](../../../backend/services/enterprise_foundation/services/address/service.go#L281) · [repository/address/address_repository.go L145](../../../backend/services/enterprise_foundation/repository/address/address_repository.go#L145) · [db/migrations/000001_create_enterprise_foundation_schema.up.sql L109](../../../backend/services/enterprise_foundation/db/migrations/000001_create_enterprise_foundation_schema.up.sql#L109).

**Kriteria penutupan.** Permintaan identik serentak mengembalikan satu ID dengan satu hasil created. Apartemen/baris kedua atau label berbeda tetap mengikuti semantik kontrak lama.

## EFL-08 — Pendaftaran party menerima status aktif langsung dari pemanggil

createParty memilih PENDING_VERIFICATION sebagai default, tetapi meneruskan nilai status masukan ke NewParty. Status ACTIVE memenuhi pemeriksaan kelayakan yang melihat status record, walau perintah verifyAndActivateParty belum dijalankan. Ini celah alur verifikasi bisnis. Kedua operasi saat ini menggunakan permission tulis party.

**Bukti produksi.** [adapter/enterprise_foundation/party_adapter.go L211](../../../backend/gateway/internal/adapter/enterprise_foundation/party_adapter.go#L211) · [grpc/handler/party_handler.go L144](../../../backend/services/enterprise_foundation/transport/grpc/handler/party_handler.go#L144) · [services/party/service.go L160](../../../backend/services/enterprise_foundation/services/party/service.go#L160) · [domain/entity/party.go L76](../../../backend/services/enterprise_foundation/domain/entity/party.go#L76) · [domain/entity/party.go L208](../../../backend/services/enterprise_foundation/domain/entity/party.go#L208) · [graphql/enterprise_foundation/ef_party.graphql L173](../../../backend/gateway/graphql/enterprise_foundation/ef_party.graphql#L173) · [integration/enterprise_foundation/client.go L323](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L323).

**Kriteria penutupan.** Party yang hanya didaftarkan belum lolos eligibility. Verifikasi tercatat dapat membuatnya aktif. Individu HRM dapat dipakai tanpa diberi role vendor/supplier.

## Keputusan atas saran domain

- **Master orang dan pegawai.** HRM sudah memvalidasi party individual yang aktif sebelum membuat worker. Pertahankan referensi itu dan ID lama. EF memegang identitas pihak, HRM memegang PII ketenagakerjaan, status dan assignment. Pembuatan worker tidak mensyaratkan role komersial. Saran master orang baru memerlukan peta migrasi dan konsumen sebelum dapat diputuskan. [integration/enterprise_foundation/client.go L323](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L323).
- **Mitra, stok dan pembayaran.** Party yang telah diverifikasi serta role aktif adalah bukti EF. Procurement menguasai kualifikasi dan persetujuan vendor/supplier. Finance menguasai payable dan pembayaran. EF memvalidasi identitas fasilitas dan subtipe warehouse, WMS memegang stok dan mutasi. Jangan menambah mutation pembayaran, penilaian vendor atau pengelolaan stok ke EF. [Kebijakan kelayakan target scope](../../../backend/services/enterprise_foundation/domain/service/scope_policy.go) · [Mandat produk EF](../../../docs/03-managements/enterprise-foundation.md).
- **Suspensi dan penutupan definitif.** Pertahankan perintah suspend/deactivate untuk memblokir pemakaian baru segera. Bila diperlukan penutupan operasional lintas domain, rancang bukti penyelesaian terpisah dengan owner stok, assignment, dan komitmen finansial. Pembacaan histori menggunakan ID dan snapshot asal. Audit ini belum membuktikan kesiapan alur clearance lintas owner. [services/legalentity/service.go L372](../../../backend/services/enterprise_foundation/services/legalentity/service.go#L372).
- **Identitas partner dan alamat.** Pertahankan kode dan kunci pajak dengan aturan keunikan yang ada. Nama serupa, email atau alamat bersama hanya kandidat peninjauan petugas. Penggabungan identitas otomatis berisiko mencampur saldo, dokumen dan mandat dua party berbeda. [db/migrations/000001_create_enterprise_foundation_schema.up.sql L267](../../../backend/services/enterprise_foundation/db/migrations/000001_create_enterprise_foundation_schema.up.sql#L267).
- **CRM, Commerce dan DRM.** EF menjadi rujukan party, relationship dan kontak. CRM menyimpan lead, Commerce memegang akun belanja melalui IAM, DRM menyimpan dokumen dan hak baca. Referensi bukti pada verifikasi dapat menyimpan ID opaque DRM. Hindari menyalin dokumen, data penskoran atau seluruh profil ke EF. Perubahan ini tidak memerlukan endpoint HTTP baru. [Kepemilikan identitas dan atribut](../../../agents/users.md).

## Handoff dan batas verifikasi

Agent EF memimpin perbaikan lintas management, menjaga alur provisioning yang sudah ada dan melaporkan bukti runtime tiap ID. Nomor EFL khusus audit ini. IEP-01, IEP-03 dan IEP-05 tetap dilacak oleh instruksi terdahulu sesuai bukti masing-masing. Angka 112 adalah cakupan kontrak statis. Status delapan perbaikan adalah **terbuka**. [Instruksi implementasi baru](../../../agents/instructions/enterprise_foundation/4-leader-remediasi-logika-master-dan-mandat-instructions.md) · [Handoff EF sebelumnya](../../../agents/instructions/enterprise_foundation/3-leader-remediasi-iam-ef-procurement-instructions.md) · [Handoff IAM terkait](../../../agents/instructions/iam/4-leader-remediasi-logic-dan-siklus-akses-iam-instructions.md).
