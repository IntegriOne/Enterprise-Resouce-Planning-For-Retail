# List operasi Document & Records

**22 Query dan 32 Mutation, total 54 field root usulan.** Semua menggunakan satu Gateway dan satu service privat `document_record`. Operasi privat serta operasi HRM/Commerce yang digunakan ulang memiliki daftar tersendiri. Nama operasi dan permission di bawah merupakan usulan implementasi. [Instruksi DB sampai Gateway](../../instructions/document_record/1-leader-implementasi-grpc-gateway-dan-integrasi-dokumen-instructions.md) · [Kontrak dua arah per management](2-kontrak-dan-wiring-lintas-management.md)

## RBAC

Kode pada kolom izin memakai awalan usulan `document_record.`. Cocokkan katalog IAM aktual sebelum mendaftarkan kode. Semua operasi memeriksa sesi IAM terkini, izin tindakan, entity/scope, klasifikasi maksimum pemanggil, pemilik transaksi dan hubungan terhadap objek. Hak staf perpustakaan terbatas pada mandatnya. **Penerima** berarti user persis atau PIC aktif untuk party + peran + resource yang tepat, dengan publikasi aktif dan hak owner masih berlaku. Grant SELF milik seorang pembeli tidak berlaku atas dokumen orang lain. Untuk slip gaji gunakan jalur HRM, termasuk ketika mengakses melalui node atau hubungan dari domain lain.

Hak audit, admin konfigurasi, pembuat hold, peninjau hold, pengaju dan penyetuju disposal berdiri sendiri. MFA/reauth serta pemisahan petugas mengikuti klasifikasi operasi sensitif di katalog. Satu akun superadmin berguna untuk menyiapkan uji, sedangkan keputusan SoD melibatkan dua manusia berbeda. Pasang filter izin **sebelum** hitung total, search, sort atau pagination. Respons daftar memberikan nomor, judul yang diizinkan, jenis, versi, status dan tanggal untuk UI. GID hanya untuk navigasi/refetch.

## Butuh GraphQL Query

| Field GraphQL | RPC DRM | Izin | Input, hasil dan kondisi uji |
|---|---|---|---|
| `documentSettings` | `GetDocumentSettings` | configuration.read | Entity → batas unggahan dan waktu sesi efektif. Sembunyikan rincian KMS, bucket dan konfigurasi provider. |
| `documentLibraries` | `ListDocumentLibraries` | configuration.read | Entity dan owner dalam scope → perpustakaan, label, klasifikasi, status. |
| `documentTypes` | `ListDocumentTypes` | configuration.read | Library sah → tipe, format, batas ukuran, syarat signature dan pembatasan publikasi. |
| `retentionPolicies` | `ListRetentionPolicies` | retention.read | Type/as-of → versi kebijakan, masa efektif, anchor dan basis yang boleh dilihat. |
| `documents` | `SearchDocuments` | metadata.read | Filter library, resource owner, jenis, nomor dan rentang → Connection. Dokumen tersembunyi tidak memengaruhi hasil atau total. |
| `document` | `GetDocument` | metadata.read | GID → identitas dokumen dan metadata yang berhak dilihat. Tautan lintas owner diperiksa sendiri. |
| `documentVersions` | `ListDocumentVersions` | metadata.read | Document → riwayat berurutan. Penerima hanya melihat versi yang diizinkan untuknya. |
| `documentVersion` | `GetDocumentVersion` | metadata.read | Versi persis → nomor versi, jenis, ukuran aman dan status. ID dari dokumen lain tidak membuka relasi. |
| `documentLinks` | `ListDocumentLinks` | links.read | Versi dan konteks → link aktif yang disetujui kedua owner. Redaksi judul/nomor objek lain sesuai visibilitas. |
| `documentUploadStatus` | `GetUploadStatus` | upload.read | Sesi upload → progres, expiry, kode kegagalan aman. Hanya pengunggah berwenang atau staf scoped. |
| `documentPublications` | `ListPublications` | publication.read | Versi → tujuan, status, tanggal penerbitan dan pencabutan. Daftar recipient memerlukan izin sendiri. |
| `publicationRecipients` | `ListPublicationRecipients` | recipient.read | Publication → jenis user/party, scope, expiry, revoked. Tolak enumerasi alamat atau penerima lain dari portal. |
| `myPublishedDocuments` | `ListMyPublications` | publication.read_self | Sesi → publikasi aktif yang memang boleh didaftar oleh owner. Terapkan pembatasan domain sensitif serta perubahan status pegawai/PIC sebelum pagination. |
| `documentContentChunk` | `StreamDocument` | content.read | Version, konteks publikasi/tujuan, offset/limit terbatas → potongan base64, offset berikut dan EOF. Otorisasi ulang per panggilan/range, no-store, audit sebelum byte pertama. Unduhan untuk owner yang mensyaratkan jalur khusus mengikuti kebijakan yang sama. |
| `documentSignatures` | `ListSignatureEvidence` | signature.read | Version → status validasi dan waktu. Sembunyikan data identitas penanda tangan yang tidak diperlukan. |
| `declaredRecord` | `GetRecord` | records.read | Versi → kebijakan frozen, anchor, retain-until, status. Unresolved atau indefinite ditampilkan eksplisit. |
| `retentionHistory` | `ListRetentionEvents` | retention.read | Record → trigger terverifikasi, sumber, correction dan waktu yang disetujui. |
| `documentLegalHolds` | `ListLegalHolds` | hold.read | Entity/dokumen → hold pending/active/released yang dapat dilihat petugas. Alasan legal tidak tampil pada portal. |
| `dispositionQueue` | `ListDispositions` | disposition.read | Entity + status → antrean per record, peninjau dan perkara unresolved. |
| `dispositionStatus` | `GetDispositionStatus` | disposition.read | Request → hasil provider terverifikasi atau executing/unknown dan tindakan berikut. |
| `documentAuditTrail` | `ListDocumentAuditEvents` | audit.read | Resource/waktu → tindakan, outcome dan correlation yang disanitasi. Batasi rentang, hasil dan data aktor. |
| `documentStorageCases` | `ListStorageCases` | operations.read | Antrean scan/seal/purge yang macet, umur dan kode pemulihan. Batas scope, tanpa path storage atau isi file. |

## Butuh GraphQL Mutation

| Field GraphQL | RPC DRM | Izin | Command, perubahan dan penerimaan |
|---|---|---|---|
| `configureDocumentSettings` | `ConfigureSettings` | configuration.manage | Validasi batas file, kuota, TTL dan storage policy terhadap batas deployment. Tidak ada opsi menonaktifkan scan/hold. |
| `createDocumentLibrary` | `CreateLibrary` | configuration.manage | Entity dan owning management tervalidasi, kode unik, floor klasifikasi. |
| `updateDocumentLibrary` | `UpdateLibrary` | configuration.manage | Label/aktif/floor dengan expected version; pertahankan owner, entity dan proteksi dokumen lama. |
| `createDocumentType` | `CreateDocumentType` | configuration.manage | Type per library; allowlist MIME, ukuran, signature, external publication dan floor. |
| `reviseDocumentType` | `ReviseDocumentType` | configuration.manage | Ubah aturan prospektif dengan bukti; aturan baru tidak membuka versi lama secara implisit. |
| `proposeRetentionPolicy` | `ProposeRetentionPolicy` | retention.propose | Pengusul membuat usulan berversi berisi type, trigger, durasi calendar months/indefinite, interval dan basis kebijakan. Simpan digest proposal dan identitas aktual sebelum review; policy yang berlaku tetap versi approved. |
| `approveRetentionPolicy` | `ApproveRetentionPolicy` | retention.approve | Usulan kebijakan terdokumentasi + referensi persetujuan. Verifikasi pengusul asli dan approver independen, periode efektif tanpa overlap. Simpan versi kebijakan beserta evidence. |
| `registerSourceDocument` | `RegisterDocument` | documents.create | Library/type + sumber dan key owner → identitas stabil setelah owner memverifikasi draft/objek. Replay mengembalikan dokumen yang sama. |
| `createDocumentVersion` | `CreateDocumentVersion` | versions.create | Document, source revision dan predecessor → percobaan versi baru, nomor dialokasikan di bawah lock. File lama tetap utuh. |
| `beginDocumentUpload` | `BeginUpload` | upload.create | Version pending, tipe dan ukuran deklarasi → satu session berumur pendek. Retry mengembalikan sesi identik. |
| `uploadDocumentChunk` | `WriteUploadChunk` | upload.write | Session, offset, chunk terbatas → jumlah diterima. Duplikat dengan byte identik idempoten, overlap/gap/digest konflik ditolak. |
| `finalizeDocumentUpload` | `FinalizeUpload` | upload.finalize | Bekukan generation, verifikasi ukuran dan MIME aktual, antrekan scan. Hasil awal boleh scanning. |
| `cancelDocumentUpload` | `CancelUpload` | upload.cancel | Batalkan sesi yang masih boleh dibatalkan. Cleanup staging menghormati hold dan generation persis. |
| `linkDocumentVersion` | `LinkDocumentVersion` | links.manage | Versi sealed + resource/relasi terverifikasi kedua management → link scoped, tanpa grant akses baru. |
| `revokeDocumentLink` | `RevokeDocumentLink` | links.manage | Nonaktifkan relasi kontekstual berdasarkan mandat dan alasan; simpan histori serta bukti resmi yang sudah dipatok. |
| `declareDocumentRecord` | `DeclareRecord` | records.declare | Versi available → satu record dengan kebijakan approved, digest, anchor/trigger. Replay tidak mengganti kebijakan. |
| `archiveDocument` | `ArchiveDocument` | records.archive | Pindahkan lifecycle sesuai policy; byte, retention, hold dan ACL tetap ditegakkan. |
| `publishDocumentVersion` | `PublishVersion` | publication.publish | Record, exact version, bukti issuance owner, purpose dan penerima terverifikasi → publication. Untuk payslip hanya jalur terdelegasi HRM/Finance; admin generik tidak boleh menerbitkannya. |
| `revokeDocumentPublication` | `RevokePublication` | publication.revoke | Alasan dari owner → revoke publikasi spesifik dan tolak pembacaan berikutnya untuk tujuan itu. |
| `authorizeDocumentRecipient` | `AuthorizePublicationRecipient` | recipient.manage | Validasi penerima persis dari keputusan distribusi owner: IAM user atau EF party + role/PIC, TTL. Staf tidak dapat menambahkan dirinya tanpa mandat owner. |
| `revokeDocumentRecipient` | `RevokePublicationRecipient` | recipient.manage | Cabut satu penerima; uji request chunk berikutnya serta cache setelah revoke. |
| `registerDocumentHold` | `RegisterHold` | hold.request | Entity, daftar target deterministik, reason ref → pending langsung memblokir purge baru. |
| `addDocumentHoldTarget` | `AddHoldTarget` | hold.request | Daftarkan target di lock dokumen; pertahankan status preservation per target, termasuk versi mendatang. |
| `activateDocumentHold` | `ActivateHold` | hold.approve | Reviewer lain + bukti preservation seluruh target. Purge sedang in-flight/unknown menahan konfirmasi aktivasi. |
| `rejectDocumentHold` | `RejectHold` | hold.approve | Review independen, reason dan audit untuk pending; hold lain pada dokumen tetap mengikat. |
| `requestDocumentHoldRelease` | `RequestHoldRelease` | hold.release_request | Ajukan alasan/bukti pelepasan untuk satu hold aktif. Persist intent dan actor; seluruh target tetap terlindungi sampai review sukses. |
| `releaseDocumentHold` | `ReleaseHold` | hold.release | Evidence keputusan pencabutan dan pemeriksa independen; cek seluruh hold tersisa sebelum kelayakan disposal. |
| `requestRecordDisposition` | `RequestDisposition` | disposition.request | Record finite yang eligible + owner clearance + digest kelayakan → request live unik. |
| `approveRecordDisposition` | `ApproveDisposition` | disposition.approve | Orang berbeda, recheck anchor, expiry, hold, clearance, publikasi/dependensi dan exact version; eksekusi melalui job durable. |
| `rejectRecordDisposition` | `RejectDisposition` | disposition.approve | Reviewer menolak request dengan alasan → simpan bukti, tidak membuat job purge. |
| `cancelRecordDisposition` | `CancelDisposition` | disposition.cancel | Cancel sebelum fence/eksekusi, per domain state; unknown harus direkonsiliasi. |
| `requestStorageReconciliation` | `RequestStorageReconciliation` | operations.reconcile | Case + reason → lookup status intent/operation provider. Petugas tidak boleh mengeset scan clean, bytes served atau deleted. |

## Penjelasan operasi privat

| RPC / pekerja usulan | Pemanggil, input dan postcondition |
|---|---|
| `AuthorizeSourceAction` (di owner), `BindSourceEvidence`, `VerifyDocumentVersion` (di DRM) | Workload owner terdaftar + konteks aktor/tujuan/entity/resource/version/digest. Verifier mengembalikan ref immutable, state, bukti dan waktu keputusan. Verifikasi hash/UUID saja tidak membuktikan hak aktor atas transaksi. |
| `ApplyVerifiedScanResult`, `SealScannedObject` | Hanya scanner/worker terautentikasi. Cocokkan sesi, storage generation, SHA-256, ukuran dan versi engine; verifikasi generation tujuan sebelum available. |
| `RecordVerifiedSignature` | Verifier tepercaya mengikat provider assertion, identitas penanda tangan, versi asal dan digest byte bertanda tangan. Invalid/indeterminate memblokir penerbitan yang mensyaratkan signature. |
| `AuthorizeDownload`, `StreamDocument` | Keputusan internal terikat subjek, versi, purpose, scope dan publikasi; streaming memeriksa kebijakan kembali. Hanya adapter download yang mengembalikan byte terbatas ke browser. |
| `StorePrivatePayslip`, `ReadPrivatePayslip`, `RestorePrivatePayslip` | Penuhi port HRM melalui client DRM dengan validasi Finance, subject dan digest. Restore hanya untuk kehilangan fisik terkonfirmasi dengan metadata sah dan content identik; tombstone disposition menolak restore. |
| `ApplyRetentionAnchor`, `ApplyApprovedRetentionCorrection` | Event owner terverifikasi, dedup dan bukti koreksi tertinjau. Anchor ambigu menahan purge; perubahan tidak memperpendek masa secara diam-diam. |
| `DeliverDocumentEvent`, `RelayOutbox`, `ConfirmPreservation`, `ExecuteDisposition`, `ApplyVerifiedStorageDeletion`, `ReconcileStorageJob` | Caller sesuai tujuan dan workload. Job dengan lease/fence; hasil fisik memerlukan bukti versi provider. Duplicate/out-of-order, timeout lalu restart dan conflicting digest diuji per consumer. |

## Pemeliharaan inventaris

Simpan detail baru di `agents/grpc/document_record/{id}-{title}.md` dengan heading `List operasi`, `Penjelasan {operation_name}`, `Butuh GraphQL Query/Mutation`, `RBAC`, contoh outcome dan lokasi bukti. Tiap baris akhir mencantumkan proto/handler/service, permission final, SDL/adapter, hasil HTTPS allowed/denied dan status. Laporkan hitungan root domain, RPC privat, worker, dan operasi domain lain yang dipakai ulang secara terpisah. Preview, cetak, pencarian, akses lewat node dan konsumsi oleh management lain memakai policy objek yang sama.
