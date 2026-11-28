# Katalog usulan operasi CRM dan Sales

Rancangan pada 29 November 2026. Pada checkout ini CRM memiliki DBML, dengan **0 operasi CRM terimplementasi di service atau Gateway**. Target di bawah memuat **35 Query + 54 Mutation = 89 field GraphQL** serta **10 RPC privat usulan**. Dua dari 89 field berada pada kelompok handoff Customer Service yang memerlukan kontrak owner. Operasi pengiriman provider, quotation berbantuan, angka Finance dan lampiran juga dibuka setelah acceptance dependensinya lulus.

## List operasi

| Kelompok / paket | Service proto usulan | Query | Mutation |
| --- | --- | ---: | ---: |
| `relationship` | `RelationshipService` | 8 | 13 |
| `permission` | `ContactPermissionService` | 2 | 3 |
| `lead` | `LeadService` | 4 | 7 |
| `pipeline` | `OpportunityService` | 5 | 8 |
| `coordination` | `SalesCoordinationService` | 2 | 3 |
| `performance` | `SalesPerformanceService` | 3 | 4 |
| `communication` | `ConversationService` | 7 | 10 |
| `activity` | `SalesActivityService` | 3 | 5 |
| `service_handoff` | `RelationshipSupportService` | 1 | 1 |

## Penjelasan operasi, kebutuhan GraphQL dan RBAC

Setiap baris menetapkan satu root dan satu RPC usulan. Nama izin di tabel adalah suffix dari `customer_relationship_sales.`. Aktifkan izin yang telah disetujui di manifest IAM sejalan dengan implementasi. Tulis satu entri kebijakan gRPC untuk setiap method. Scope tabel berlaku juga pada node, relasi bersarang, pencarian, agregat dan total_count. Query detail memakai typed ID, daftar memakai Connection, mutasi membawa versi bila mengubah objek dan kunci idempoten untuk efek bisnis. Append timestamp, actor dan provenance dari identitas server.

### RelationshipService

Scope dan izin tambahan mengikuti aturan ini. Unit EF dan assignee. PII tambahan hanya bagi petugas berizin. Proto target `contract/proto/customer_relationship_management/v1/relationship.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_relationship.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `relationshipSettings` | `GetRelationshipSettings` | `settings.read` | Entitas aktif → zona waktu, mata uang, skala dan retensi. |
| Mutation | `configureRelationshipSettings` | `ConfigureRelationshipSettings` | `settings.manage` | Entitas, policy dan versi → konfigurasi tunggal tervalidasi EF/Finance. |
| Query | `relationshipOrganization` | `GetRelationshipOrganization` | `relationships.read` | ID → nomor, nama dagang, unit, status dan relasi PIC. |
| Query | `relationshipOrganizations` | `ListRelationshipOrganizations` | `relationships.read` | Unit/status/kata kunci → Connection organisasi terscope. |
| Mutation | `createRelationshipOrganization` | `CreateRelationshipOrganization` | `relationships.write` | Nama, unit dan assignee → profil prospek perusahaan. |
| Mutation | `updateRelationshipOrganization` | `UpdateRelationshipOrganization` | `relationships.write` | Profil dan versi → perubahan data hubungan, simpan histori penugasan. |
| Mutation | `linkRelationshipParty` | `LinkRelationshipParty` | `relationships.link` | Organisasi, party, bukti → tautan EF customer B2B aktif dan unik. |
| Query | `relationshipContact` | `GetRelationshipContact` | `relationships.read` | ID → nama, organisasi, hint kontak, status verifikasi. |
| Query | `relationshipContacts` | `ListRelationshipContacts` | `relationships.read` | Unit/assignee/status → Connection dengan hint tertutup. |
| Mutation | `createRelationshipContact` | `CreateRelationshipContact` | `relationships.write` | Subjek, organisasi opsional dan unit → kontak/provisional. |
| Mutation | `updateRelationshipContact` | `UpdateRelationshipContact` | `relationships.write` | Profil, unit tujuan dan versi → validasi sumber dan tujuan perubahan. |
| Mutation | `linkRelationshipIdentity` | `LinkRelationshipIdentity` | `relationships.link` | Jenis identitas, ID owner dan bukti kepemilikan → tautan B2C/PIC terverifikasi. |
| Mutation | `addContactEndpoint` | `AddContactEndpoint` | `relationships.write` | Kontak, jenis, namespace dan endpoint → data terenkripsi serta hint. |
| Mutation | `deactivateContactEndpoint` | `DeactivateContactEndpoint` | `relationships.write` | Endpoint dan alasan → hentikan pemakaian baru, pertahankan suppression. |
| Query | `contactMatchCandidates` | `ListContactMatchCandidates` | `relationships.deduplicate` | Endpoint terotorisasi → kandidat tersamarkan, tanpa membuka identitas lintas unit. |
| Mutation | `mergeRelationshipContacts` | `MergeRelationshipContacts` | `relationships.deduplicate` | Sumber, tujuan, versi keduanya dan bukti → satu profil kanonik, histori tetap utuh. |
| Mutation | `anonymizeRelationshipContact` | `AnonymizeRelationshipContact` | `privacy.redact` | Kontak, alasan dan keputusan retensi/hold → hapus pengenal yang diizinkan, simpan jejak minimum. |
| Query | `relationshipTags` | `ListRelationshipTags` | `relationships.read` | Tag aktif → Connection label manual. |
| Mutation | `configureRelationshipTag` | `ConfigureRelationshipTag` | `settings.manage` | Kode dan nama → buat/perbarui/arsipkan label. |
| Mutation | `setRelationshipContactTags` | `SetRelationshipContactTags` | `relationships.write` | Kontak, versi dan daftar tag → ganti himpunan atomik tanpa duplikasi. |
| Query | `customerRelationshipOverview` | `GetCustomerRelationshipOverview` | `relationships.read` | Subjek → ringkasan CRM dan bagian dari owner sesuai hak per bagian, waktu pembaruan dan status ketersediaan. |

### ContactPermissionService

Scope dan izin tambahan mengikuti aturan ini. Subject dan channel dalam scope. Keputusan serta tindakan membuka blokir punya audit tersendiri. Proto target `contract/proto/customer_relationship_management/v1/permission.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_permission.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `contactPermissionHistory` | `ListContactPermissionHistory` | `permissions.read` | Endpoint, purpose, periode → Connection keputusan tersamarkan. |
| Query | `effectiveContactPermission` | `GetEffectiveContactPermission` | `permissions.read` | Endpoint, purpose, waktu sekarang → effective decision dan alasan minimum. |
| Mutation | `recordContactPreference` | `RecordContactPreference` | `permissions.record` | Purpose, dasar, bukti, versi notice, waktu berlaku → keputusan append-only. |
| Mutation | `suppressContactEndpoint` | `SuppressContactEndpoint` | `permissions.record` | Endpoint, scope larangan dan bukti → suppression aktif, batalkan antrean terdampak. |
| Mutation | `liftContactSuppression` | `LiftContactSuppression` | `permissions.lift` | ID larangan dan bukti baru → lift beraudit, evaluasi ulang izin dasar. |

### LeadService

Scope dan izin tambahan mengikuti aturan ini. Sales melihat penugasan sendiri. Supervisor memakai grant unit. Assignment lintas unit memerlukan mandat keduanya. Proto target `contract/proto/customer_relationship_management/v1/lead.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_lead.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `salesLeadSources` | `ListLeadSources` | `leads.read` | Sumber aktif → daftar dengan kode/display name. |
| Mutation | `configureLeadSource` | `ConfigureLeadSource` | `settings.manage` | Kode, jenis dan producer → simpan konfigurasi asal yang dapat ditelusuri. |
| Query | `salesLead` | `GetSalesLead` | `leads.read` | ID → nomor, subjek, status, assignee, qualification dan kepentingan produk. |
| Query | `salesLeads` | `ListSalesLeads` | `leads.read` | Assignee/unit/status/sumber → Connection terscope. |
| Mutation | `createSalesLead` | `CreateSalesLead` | `leads.write` | Subjek dan sumber → lead new, deduplikasi per source record. |
| Mutation | `assignSalesLead` | `AssignSalesLead` | `leads.assign` | Lead dan salesperson aktif → assignment teraudit, status dipertahankan. |
| Mutation | `transitionSalesLead` | `TransitionSalesLead` | `leads.write` | State tujuan, versi dan alasan → transisi operasional/contacted/nurturing/disqualified/archive. Konversi melalui perintah terpisah. |
| Mutation | `qualifySalesLead` | `QualifySalesLead` | `leads.write` | Bukti kebutuhan dan versi → qualified dengan waktu dan histori. |
| Mutation | `recordLeadInterest` | `RecordLeadInterest` | `leads.write` | Lead, kebutuhan, item/UOM opsional, qty dan target waktu → interest. |
| Mutation | `convertSalesLead` | `ConvertSalesLead` | `leads.convert` | Lead qualified + opportunity baru/terverifikasi atau order sah untuk subjek sama → bukti konversi tunggal. |
| Query | `salesLeadHistory` | `ListSalesLeadHistory` | `leads.read` | Lead → Connection transisi, alasan dan waktu. |

### OpportunityService

Scope dan izin tambahan mengikuti aturan ini. Unit dan pemilik opportunity. Konfigurasi pipeline memerlukan mandat unit. Harga binding berasal dari owner. Proto target `contract/proto/customer_relationship_management/v1/pipeline.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_pipeline.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `salesPipelines` | `ListSalesPipelines` | `pipeline.read` | Unit → pipeline dan tahapan terurut, termasuk status aktif. |
| Mutation | `configureSalesPipeline` | `ConfigureSalesPipeline` | `pipeline.configure` | Unit, identitas, status → pipeline dengan minimal tahap open/won/lost saat diaktifkan. |
| Mutation | `configureSalesPipelineStage` | `ConfigureSalesPipelineStage` | `pipeline.configure` | Pipeline, kode, urutan, kind dan probabilitas → versi tahap yang dapat dipakai. |
| Query | `salesOpportunity` | `GetSalesOpportunity` | `pipeline.read` | ID → nomor, lead, stage, item, estimasi dan bukti penutupan. |
| Query | `salesOpportunities` | `ListSalesOpportunities` | `pipeline.read` | Stage/assignee/expected close → Connection kesempatan. |
| Mutation | `createSalesOpportunity` | `CreateSalesOpportunity` | `pipeline.write` | Subjek, optional lead, pipeline dan tahap open → opportunity aktif. |
| Mutation | `updateOpportunityEstimate` | `UpdateOpportunityEstimate` | `pipeline.write` | Versi, baris requirements, nilai/tanggal/probabilitas → estimasi terverifikasi decimal. |
| Mutation | `assignSalesOpportunity` | `AssignSalesOpportunity` | `pipeline.assign` | Opportunity dan assignee aktif → ubah tanggung jawab, credit beku dipertahankan. |
| Mutation | `moveSalesOpportunityStage` | `MoveSalesOpportunityStage` | `pipeline.write` | Tahap open dalam pipeline yang sama → stage history dan event. |
| Mutation | `markSalesOpportunityLost` | `MarkSalesOpportunityLost` | `pipeline.write` | Alasan dan versi → closed lost, probabilitas nol. |
| Mutation | `reopenSalesOpportunity` | `ReopenSalesOpportunity` | `pipeline.review` | Alasan, bukti dan versi → kembali ke open melalui review. |
| Query | `salesOpportunityHistory` | `ListSalesOpportunityHistory` | `pipeline.read` | Opportunity → Connection riwayat stage dan perubahan probabilitas. |
| Query | `salesPipelineForecast` | `GetSalesPipelineForecast` | `pipeline.report` | Unit, periode, mata uang → sisa estimate × probability untuk peluang terbuka. |

### SalesCoordinationService

Scope dan izin tambahan mengikuti aturan ini. Sales berwenang untuk pelanggan dan opportunity tersebut. Koreksi atribusi sensitif memakai approval terpisah. Proto target `contract/proto/customer_relationship_management/v1/coordination.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_coordination.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Mutation | `requestAssistedQuotation` | `RequestAssistedQuotation` | `coordination.request` | Opportunity B2B, PIC/buyer dan requirements → request durable ke Commerce, return request ID/status. |
| Query | `salesCoordinationRequest` | `GetSalesCoordinationRequest` | `coordination.read` | Request ID → pending/selesai/gagal dan nomor quotation bila sudah terverifikasi. |
| Query | `salesDocumentLinks` | `ListSalesDocumentLinks` | `coordination.read` | Subjek/opportunity → Connection dokumen dengan nomor owner dan versi. |
| Mutation | `requestSalesCreditCorrection` | `RequestSalesCreditCorrection` | `coordination.correct` | Link, assignee/unit usulan, tanggal efektif dan bukti → permintaan koreksi. |
| Mutation | `approveSalesCreditCorrection` | `ApproveSalesCreditCorrection` | `coordination.approve` | Koreksi dan versi → perbarui histori berlaku per tanggal sesudah persetujuan independen. |

### SalesPerformanceService

Scope dan izin tambahan mengikuti aturan ini. Sales sendiri atau supervisor unit. Finance menentukan revenue, HRM menentukan insentif. Proto target `contract/proto/customer_relationship_management/v1/performance.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_performance.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `salesTargets` | `ListSalesTargets` | `targets.read` | Salesperson/unit, periode dan metric → Connection target berversi. |
| Query | `salesTarget` | `GetSalesTarget` | `targets.read` | ID → term, status, approver dan supersession. |
| Mutation | `createSalesTarget` | `CreateSalesTarget` | `targets.write` | Salesperson/unit, periode, metric, currency, amount → draft. |
| Mutation | `approveSalesTarget` | `ApproveSalesTarget` | `targets.approve` | Draft dan versi → approved sesudah cek overlap dan SoD. |
| Mutation | `supersedeSalesTarget` | `SupersedeSalesTarget` | `targets.approve` | Target lama + draft pengganti → efektif atomik untuk lingkup sama, jejak lama tersimpan. |
| Mutation | `cancelSalesTarget` | `CancelSalesTarget` | `targets.approve` | Target, versi, alasan → cancelled tanpa menghapus angka historis. |
| Query | `salesAchievement` | `GetSalesAchievement` | `targets.read` | Salesperson/unit, metric dan as_of → aktual, target, rasio opsional, watermark dan sumber owner. |

### ConversationService

Scope dan izin tambahan mengikuti aturan ini. Agent inbox yang ditugaskan atau supervisor unit. Endpoint, isi, lampiran dan template mempunyai kontrol akses terpisah. Proto target `contract/proto/customer_relationship_management/v1/communication.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_communication.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `relationshipChannels` | `ListRelationshipChannels` | `messages.read` | Unit → daftar inbox aktif dan kemampuan manual/provider. |
| Mutation | `configureRelationshipChannel` | `ConfigureRelationshipChannel` | `channels.manage` | Akun provider, secret reference dan unit → konfigurasi tanpa secret dalam response. |
| Query | `relationshipConversations` | `ListRelationshipConversations` | `messages.read` | Akun, status, assignee → Connection percakapan terotorisasi. |
| Query | `relationshipConversation` | `GetRelationshipConversation` | `messages.read` | ID → peserta yang dapat dilihat, subjek dan status. |
| Mutation | `openRelationshipConversation` | `OpenRelationshipConversation` | `messages.write` | Channel, subjek dan endpoint → thread stabil, terikat satu akun. |
| Mutation | `assignRelationshipConversation` | `AssignRelationshipConversation` | `messages.assign` | Conversation dan assignee → assignment valid untuk inbox dan unit. |
| Mutation | `resolveConversationParticipant` | `ResolveConversationParticipant` | `relationships.link` | Peserta unresolved dan bukti → cocokkan kontak resmi setelah review. |
| Mutation | `transitionRelationshipConversation` | `TransitionRelationshipConversation` | `messages.write` | Status dan versi → pending/resolved/reopen sesuai transisi, simpan waktu. |
| Query | `relationshipMessages` | `ListRelationshipMessages` | `messages.read` | Conversation + cursor → pesan yang diizinkan dan hint status. |
| Mutation | `queueRelationshipMessage` | `QueueRelationshipMessage` | `messages.send` | Satu recipient, purpose, key dan template/version → queued setelah keputusan purpose yang valid. |
| Mutation | `cancelQueuedRelationshipMessage` | `CancelQueuedRelationshipMessage` | `messages.send` | Pesan queued dan versi → cancelled sebelum claim dispatch. |
| Query | `relationshipMessageDelivery` | `GetRelationshipMessageDelivery` | `messages.read` | Pesan → status, waktu bukti provider dan ketidakpastian. |
| Query | `relationshipMessageTemplates` | `ListRelationshipMessageTemplates` | `templates.read` | Channel, purpose, bahasa → Connection versi template. |
| Mutation | `submitRelationshipMessageTemplate` | `SubmitRelationshipMessageTemplate` | `templates.write` | Konten, schema variabel dan bahasa → versi baru pending review. |
| Mutation | `reviewRelationshipMessageTemplate` | `ReviewRelationshipMessageTemplate` | `templates.approve` | Versi, keputusan dan alasan → approve/reject/retire sesuai state dan SoD. |
| Mutation | `attachRelationshipDocument` | `AttachRelationshipDocument` | `attachments.link` | Pesan, document/version, digest → lampiran setelah verifikasi scope serta DRM. |
| Query | `relationshipMessageAttachment` | `GetRelationshipMessageAttachment` | `attachments.read` | ID → metadata teredaksi dan akses mediated DRM untuk versi yang tetap diizinkan. |

### SalesActivityService

Scope dan izin tambahan mengikuti aturan ini. Aktivitas dan task dibatasi assignee/unit serta subjek yang dirujuk. Proto target `contract/proto/customer_relationship_management/v1/activity.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_activity.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Query | `salesFollowUps` | `ListSalesFollowUps` | `activities.read` | Assignee, status, due window → Connection tugas termasuk overdue turunan. |
| Query | `salesFollowUp` | `GetSalesFollowUp` | `activities.read` | ID → konteks, due, reminder dan outcome. |
| Mutation | `scheduleSalesFollowUp` | `ScheduleSalesFollowUp` | `activities.write` | Subjek, assignee, due dan reminder → tugas pending. |
| Mutation | `rescheduleSalesFollowUp` | `RescheduleSalesFollowUp` | `activities.write` | Tugas, versi dan jadwal baru → batalkan reminder lama, simpan jejak. |
| Mutation | `completeSalesFollowUp` | `CompleteSalesFollowUp` | `activities.write` | Tugas + outcome → completed dan satu activity atomik. |
| Mutation | `cancelSalesFollowUp` | `CancelSalesFollowUp` | `activities.write` | Tugas, versi dan alasan → cancelled, reminder usang diabaikan. |
| Mutation | `recordSalesActivity` | `RecordSalesActivity` | `activities.write` | Subjek, waktu dan outcome manual → activity, misalnya panggilan telepon. |
| Query | `salesActivityTimeline` | `ListSalesActivityTimeline` | `activities.read` | Kontak/lead/opportunity → Connection kronologis terscope. |

### RelationshipSupportService

Scope dan izin tambahan mengikuti aturan ini. Bersyarat pada kontrak Customer Service dan kepemilikan subjek. Data tiket ditampilkan sesuai grant owner. Proto target `contract/proto/customer_relationship_management/v1/service_handoff.proto`. SDL target `backend/gateway/graphql/customer_relationship_sales/crs_service_handoff.graphql`.

| GraphQL | Operation name | RPC | Izin | Input → hasil / aturan |
| --- | --- | --- | --- | --- |
| Mutation | `requestRelationshipSupport` | `RequestRelationshipSupport` | `support.request` | Subjek, alasan, correlation id dan bukti → request durable, nomor tiket sesudah ACK dari CS. |
| Query | `relationshipSupportRequests` | `ListRelationshipSupportRequests` | `support.read` | Subjek → Connection status permintaan dan referensi tiket owner yang dapat dilihat. |

## Operasi khusus workload

Kontrak usulan di `contract/proto/customer_relationship_management/v1/integration.proto`, service `RelationshipIntegrationService`. Gateway publik tidak mendaftarkan kesepuluh RPC ini. Apabila producer menggunakan DeliverRelationshipEvent, jalur dispatch memakai implementasi service yang sama dengan RPC bertipe dan identitas efek yang sama sehingga dua jalur penerimaan tetap idempoten.

| RPC | Arah / identitas | Input, output dan pemeriksaan |
| --- | --- | --- |
| `VerifySalesOpportunityContext` | Commerce → CRM | opportunity, buyer/party, unit, item requirements, version → subject, validity, version/digest dan expiry. Tidak mengembalikan endpoint. |
| `ApplyVerifiedSalesDocument` | Commerce → CRM | Quotation atau referensi order, ID Commerce serta bukti status OMS, subjek, net/currency/line, versi dan digest → link stabil. Hanya order confirmed yang memenuhi syarat won. |
| `ApplySalesDocumentReversal` | Commerce dengan fakta OMS → CRM | Owner key, baris atau jumlah yang dibatalkan, versi dan sebab → kurangi booking belum dipenuhi dan ajukan review bila dasar won hilang. |
| `ApplyRecognizedSalesEvidence` | Finance → CRM | Posting atau credit, original sale, line, jumlah neto, mata uang, tanggal akuntansi, versi, digest → proyeksi berwatermark, dapat dibangun ulang. |
| `GetVerifiedSalesAchievement` | HRM atau reporting workload → CRM | Worker/principal, mandat unit, metric/period/as_of → jumlah, basis credit, kualitas data dan digest. HRM memutuskan hak insentif. |
| `EvaluateContactPermission` | Adapter berizin → CRM | Referensi endpoint internal, purpose, account dan konteks sah → allow/deny, alasan minimum, policy revision. Evaluasi ulang saat kirim. |
| `ReceiveRelationshipMessage` | Adapter inbound → CRM | Provider account, key, thread, peserta dan waktu → pesan sekali, thread unresolved jika subjek belum diketahui. |
| `ApplyMessageDeliveryEvidence` | Adapter delivery → CRM | Account, provider message/event ID, status dan event time → bukti append-only, status monoton, hasil unknown direkonsiliasi. |
| `ApplySupportHandoffResult` | Customer Service → CRM | Correlation/request, opaque ticket ref, status yang boleh dibagi dan versi → proyeksi respons handoff. Bersyarat pada kontrak CS. |
| `DeliverRelationshipEvent` | Producer allowlist → CRM | Envelope event ID/type/version, aggregate, correlation dan digest → ACK inbox durable. Dispatcher internal memilih bounded handler di atas. ACK transport terpisah dari hasil bisnis. |

## Kontrak yang dilihat pengguna

- Detail memuat nomor kontak/lead/opportunity, nama tampilan, unit, assignee, stage/status, tanggal, nominal beserta mata uang sesuai izin. Masking endpoint dan field keuangan dilakukan server. Global ID hanya untuk navigasi klien.
- Untuk `salesLeads` usulkan `salesLeads(first: Int, after: String, last: Int, before: String, status: SalesLeadStatus, assigned_user_id: ID): SalesLeadConnection!`. Urutan stabil menggunakan pasangan tanggal dan ID. Otorisasi dan filter yang sama dipakai edges, nodes, total_count, cursor, dan Node. Terapkan pola serupa untuk daftar lain.
- Untuk `qualifySalesLead`, input berisi lead_id, expected_version, evidence dan idempotency_key. Respons `QualifySalesLeadPayload implements MutationResult` memuat sales_lead, errors, was_successful. Untuk request lintas owner, respons memuat request ID, nomor bisnis, status dan waktu bukti, dengan progress yang dapat ditanya kembali.
- Grafik forecast berbasis qty/value residual yang masih terbuka. Dashboard per periode menampilkan dua metrik terpisah, target, aktual, as_of dan watermark source. Tampilkan bagian tidak tersedia ketika kontrak owner terganggu. Nilai nol hanya diberikan apabila benar-benar terukur.
- Tidak ada field publik untuk mengubah won, menandai pesan delivered, memasukkan angka pencapaian manual, memublikasikan kampanye, atau mengotorisasi payout/refund. Bukti producer privat menggerakkan perubahan tersebut sesuai kepemilikan masing-masing.

[Urutan implementasi backend dan Gateway](../../../agents/instructions/customer_relationship_sales/1-leader-implementasi-server-grpc-dan-gateway-instructions.md) · [Kontrak integrasi dan penerima event](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md)
