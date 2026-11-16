# Application service HRM

Status 16 November 2026: 22 package domain pada `internal/service/`, 84 resource tabel bisnis, dan composition root `internal/service/registry.go`. Dua tabel integrasi tetap milik repository inbox/outbox. Keberadaan method bukan klaim seluruh workflow HRM sudah selesai atau API sudah aktif.

## Urutan review

1. `internal/service/registry.go`: binding eksplisit seluruh dependency repository dan pembaca performance untuk talent.
2. `<domain>/service.go`: constructor dan method publik per entity; `input.go`: kontrak input operasi; `mapper.go`: pemetaan field, default, audit dan guard lifecycle.
3. `operation/ports.go`: kontrak policy, integrity dan repository yang dikonsumsi; `resource.go`: alur CRUD; `command.go`: transaksi command; `outbox.go`: event minimum.
4. `leave/ledger.go`, `payroll/approval.go`, `talent/finalization.go`: command yang memanggil aturan agregat pada `domain/service/`.
5. `<domain>/test/`: mock dependency, tes perilaku publik dan benchmark. Common harness pada `testsupport/` hanya dipakai tes; tidak dipasang pada registry runtime.

Kerangka awal service/DTO/tes sempat dibuat dengan alat bantu sebelum owner meminta penulisan langsung. Selanjutnya berkas tersebut dipelihara lewat patch langsung; generator service tidak menjadi bagian tooling modul. Catatan asal kerangka dipertahankan secara eksplisit. Generator entity/repository yang sudah disepakati pada task sebelumnya tetap terpisah dari application service.

## Kontrak CRUD dan pagination

Method publik mencakup `Create<Entity>`, `Update<Entity>`, `Delete<Entity>`, `Get<Entity>ByID`, `GetAll<Entities>` dan finder bertipe `Find<Entity>By<Field>`. Service meneruskan operasi ke port repository `Insert`, `Update`, `DeleteByID`, `SelectByID`, `SelectWhere` dan `CountWhere`; tidak mengimpor GORM selain lewat composition root repository.

`GetAll`/finder menerima `limit, offset` dan mengembalikan `operation.Page[T]`: `Items`, `Total`, `Limit`, `Offset`. Default 20, maksimum 100, nilai negatif ditolak. Scope tepercaya diterapkan bersama filter pencarian **sebelum** pagination dan count. Scope kosong ditolak; `All` hanya boleh berasal dari adapter policy yang membuktikan hak akses penuh terhadap resource terkait. Pemeriksaan objek tetap dilakukan pada setiap hasil. Tidak ada penyaringan di memori yang membuat halaman atau total bocor.

Filter saat ini berbentuk AND antar kolom; beberapa nilai dalam satu filter berarti IN. Nama kolom di-allowlist dari metadata GORM, nilai memakai parameter SQL. Adapter harus menolak policy yang tidak bisa direpresentasikan dengan aman; jangan mengubah pasangan scope berkorelasi menjadi dua IN yang memperluas akses. Count/list pada READ COMMITTED tidak menjanjikan snapshot atomik.

DTO create tidak menerima primary key UUID atau audit actor server. Patch `operation.Field[T]` membedakan field tidak dikirim, explicit NULL, dan nilai nol/false. Kolom wajib menolak explicit NULL. Salinan model mempertahankan SQL NULL pada JSON nullable, berbeda dengan literal JSON `null`.

## Transaksi dan otorisasi

Constructor mewajibkan repository, transactor, outbox, `Policy` dan `Integrity`, termasuk penolakan typed nil. Jam UTC dan UUID server memiliki default; policy dan pemeriksaan referensi tidak mempunyai fallback sukses.

`Policy.Authorize` bertanggung jawab atas principal/sesi terverifikasi, permission, scope, SoD serta freshness/revisi keamanan sesuai D01–D12. Actor harus human/workload; human memerlukan session ID valid. Service memeriksa akses sebelum operasi, terhadap objek, dan sebelum callback transaksi berakhir. Perubahan actor/revisi atau revoke pada recheck membatalkan transaksi. Ini tidak menjanjikan pembatalan mutlak setelah pemeriksaan terakhir.

`Integrity.ValidateMutation` adalah dependency wajib untuk invariant lintas record dan referensi owner: IAM, Foundation, DRM, Finance serta relasi lokal HRM. Implementasi harus memakai transaksi context yang sama, mengunci root agregat terkait untuk mutasi child, memeriksa parent immutable/final, temporal validity serta bukti yang dibutuhkan. Service memberi validator salinan model agar adapter tidak mengubah write intent. **Adapter produksi ini belum diimplementasikan**, sehingga registry belum dapat dipakai sebagai runtime aman hanya dengan database.

Update/delete mengunci record melalui `FindByIDForUpdate`. Mutasi, validasi integrity dan event tersimpan dalam transaksi yang sama; error dependency/outbox/commit dikembalikan tanpa hasil sukses. Payload outbox hanya identitas record, actor dan jenis perubahan; tidak memuat gaji, narasi kasus atau salinan entity.

Versi event saat ini dihitung dari outbox agregat yang masih tersimpan; unique aggregate-version membatasi konflik konkuren. Jangan purge sumber versi terakhir sebelum ada strategi versi persisten. Pembacaan riwayat dibatasi 10.000 event; konflik atau batas terlampaui menyebabkan rollback, tanpa retry otomatis. Ini batas implementasi awal, bukan strategi throughput/retensi final.

## Guard lifecycle dan command

CRUD generik bukan jalur approval, settlement, pengikatan akun IAM, finalisasi snapshot, atau perubahan jabatan resmi. Mapper menolak perubahan status/field workflow terlindungi dan mutasi riwayat immutable dengan `ErrWorkflowRequired`. Beberapa create khusus juga sengaja ditolak, termasuk ledger, hasil payroll, rule set/framework dan keputusan yang memerlukan alur terverifikasi. Method tetap ada sebagai kontrak bertipe; operasi hanya berhasil pada state yang diizinkan.

| Command | Perilaku yang tersedia | Batas |
|---|---|---|
| `leave.PostLedgerEntry` | Lock entitlement, ledger lengkap, sumber/digest unik, replay/reversal eksak, proyeksi saldo dan outbox atomik | Human hanya sumber HRM; sumber luar membutuhkan workload dan adapter integrity. Sumber yang sama dikembalikan sebagai AlreadyExists/Conflict, tanpa posting ulang. |
| `payroll.ApprovePayrollRun` | Status calculated/review, larangan self approval, pembacaan seluruh result/line dan rekonsiliasi decimal sebelum evidence approval/outbox | Tidak menghitung tarif pajak/BPJS, membuat hasil kalkulasi, jurnal atau pembayaran. Mandat approver diverifikasi policy/integrity. |
| `talent.FinalizeTalentAssessment` | Lock assessment/cycle, rubrik dan skor lengkap, validasi bukti performance melalui service pemilik, evidence finalisasi/outbox | Tidak mengubah skor kalibrasi, assignment, promosi atau grant IAM. |

Ketiga command menolak dataset parsial akibat scope dan membatasi pengumpulan 10.000 record. Mutasi child dari semua jalur harus mengunci parent yang sama melalui adapter integrity sebelum runtime diaktifkan; mock concurrency memakai transaksi yang diserialkan dan belum membuktikan concurrency PostgreSQL.

## Bukti pengujian dan pekerjaan tersisa

Tes CRUD/read/pagination serta kegagalan authorize/integrity/insert/outbox/commit tersedia pada **satu resource representatif di setiap 22 domain**. Ini bukan bukti positive CRUD seluruh 84 tabel. Tes khusus mencakup SQL NULL/JSON null, filter scope sebelum pagination/count, pencarian AND scope, revoke sebelum commit, larangan approval melalui CRUD, concurrency update/ledger, serta approval/finalisasi dan rollback command. Mock tidak membuktikan constraint SQL atau perilaku IAM nyata.

Verifikasi modul mandiri: `GOWORK=off go fmt ./...`, `go vet ./...`, `go test -race ./...` dan `go build ./...` lulus. [Benchmark mock service](../../../../agents/benchmark/human_resource/1-service-domain-pagination-dan-payroll-approval.md) mencatat enam sampel per benchmark. Query `SelectWhere`/`CountWhere` baru belum diuji terhadap PostgreSQL pada task ini; pengujian SQL repository pada task sebelumnya tidak dianggap bukti bagi query baru.

Belum tersedia: adapter keamanan/referensi produksi, bootstrap DB/config, katalog permission disetujui, persistensi promotion/kalibrasi/finalisasi performance dan workflow lifecycle lengkap, kontrak/handler gRPC privat, wiring Gateway, outbox relay, serta E2E nyata. Tuntaskan bertahap tanpa memberi policy allow-all atau menghubungkan mutasi HRM langsung ke tabel IAM. Endpoint publik tetap hanya `/api/graphql`.
