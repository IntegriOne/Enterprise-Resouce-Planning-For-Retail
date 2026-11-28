# Audit implementasi Human Resource terhadap usulan logika bisnis

**Kesimpulan.** Dua belas temuan memerlukan tindakan. Yang langsung menahan perjalanan gaji dari HR ke Finance adalah perbedaan status kompensasi dan kalkulator yang belum terpasang di binary. Isolasi kasus hubungan karyawan serta perubahan mandat IAM perlu ditangani paralel. Prioritas tinggi pada laporan ini mencakup cacat keamanan, risiko kewajiban keuangan dan alur kerja yang terhenti.

## Dasar pemeriksaan

- Baseline checkout bersama per 29 November 2026 WIB, HEAD `738c5490db08c81d8d748ccf07b950939a6edfc5`. Sejumlah perubahan Finance, integrasi DRM/IAM dan HRM masih lokal. Path bertanda *checkout* perlu dibaca dari working tree yang sama. Sumber usulan `agents/logic/human_resource/logic.md` dipakai sebagai bahan pertimbangan untuk 43 bagian, termasuk prioritas dan skenario penerimaan.
- AST SDL menunjukkan **176 query dan 326 mutation**. Seluruh 502 nama mempunyai method resolver root dan pemanggilan method adapter HRM yang dideklarasikan pada kode produksi. Pemeriksaan handoff melalui `python3 backend/services/human_resource/tools/check_gateway_handoff.py` menghasilkan **22 service, 1.018 RPC domain, 674 RPC baca, 326 RPC mutation, 18 nonaktif, 0 selisih**. Dua kontrak workload terpisah memiliki empat RPC privat. Sejumlah finder gRPC dilayani satu connection GraphQL. Angka ini menjelaskan cakupan antarmuka. [Matriks root, RPC dan kebutuhan baru](../../../agents/grpc/hrm/45-matriks-operasi-publik-dan-kebutuhan-domain.md).
- Audit business logic menelusuri jalur produksi yang dirujuk pada masing-masing temuan, termasuk SQL, validator, runtime wiring dan owner eksternal. Pencocokan simbol dilakukan dengan graph dan pemeriksaan source langsung. SQL migrasi, snapshot dan invariant diperiksa dari file sumber. Laporan ini tidak mengklaim pengujian dinamis seluruh 502 operasi, insiden kebocoran, atau pembayaran ganda aktual. Tidak ada source test yang dinilai dan tidak ada E2E atau transfer uang yang dijalankan untuk audit ini.

| ID | Prioritas | Dampak |
| --- | --- | --- |
| HRL-01 | Tinggi | Kompensasi yang disetujui dapat hilang dari snapshot gaji |
| HRL-02 | Tinggi | Kalkulator payroll belum terpasang di binary HRM |
| HRL-03 | Tinggi | Rekonsiliasi memeriksa baris hasil tetapi belum membuktikan kelengkapan pekerja dan digest keluaran |
| HRL-04 | Tinggi | Nomor run berbeda dapat mewakili kewajiban gaji reguler yang sama |
| HRL-05 | Tinggi | Slip untuk posting sah tanpa jurnal tertahan sewaktu terbit maupun dicetak ulang |
| HRL-06 | Tinggi | Settlement menyamakan saldo nol dengan pembayaran dan belum memproyeksikan koreksi settlement |
| HRL-07 | Tinggi | Kelas kerahasiaan kasus hubungan karyawan belum membatasi pembacaan per kasus |
| HRL-08 | Tinggi | Sejumlah mutation tersedia di Gateway tetapi verifier bisnisnya belum dipasang pada startup |
| HRL-09 | Sedang | Offer yang diterima belum memiliki perintah hire yang menghubungkan lamaran dan worker |
| HRL-10 | Sedang | Pegawai dan HR kekurangan query lookup dengan scope yang sesuai |
| HRL-11 | Tinggi | Perubahan mandat setelah transfer belum mempunyai jalur privat setara provisioning dan terminasi |
| HRL-12 | Tinggi | Pemakaian cuti yang disetujui Workforce belum dapat didebit ke ledger HRM |

## HRL-01 — Kompensasi yang disetujui dapat hilang dari snapshot gaji

ApproveCompensationRecord berakhir pada approved. Validator perubahan berikutnya mensyaratkan state sebelumnya proposed. SQL snapshot hanya membaca active dan superseded. Maka persetujuan gaji pokok melalui alur tersedia tidak membuatnya tercakup, meski tanggal efektifnya sudah masuk periode.

**Bukti kode.** [service/compensation/approval.go L31](../../../backend/services/human_resource/internal/service/compensation/approval.go#L31) · [infrastructure/security/compensation.go L164](../../../backend/services/human_resource/infrastructure/security/compensation.go#L164) · [repository/payroll/input_snapshot_capture.sql L18](../../../backend/services/human_resource/repository/payroll/input_snapshot_capture.sql#L18) · [graphql/human_resource/hrm_compensation_commands.graphql L10](../../../backend/gateway/graphql/human_resource/hrm_compensation_commands.graphql#L10).

**Kriteria penutupan.** Periode setelah tanggal efektif memuat kompensasi approved tepat sekali, periode sebelumnya tidak memuatnya.

## HRL-02 — Kalkulator payroll belum terpasang di binary HRM

Registry menyediakan WithPayrollCalculator, tetapi komposisi produksi hanya memasang penyimpan snapshot. CalculatePayrollRun menghentikan proses ketika Calculator kosong. Snapshot v1 membekukan profil, kompensasi dan adjustment. Jam kerja, lembur, bukti cutoff Workforce maupun enrollment benefit belum menjadi bagian kontrak itu.

**Bukti kode.** [human_resource/cmd/main.go L59](../../../backend/services/human_resource/cmd/main.go#L59) · [human_resource/cmd/wiring.go L147](../../../backend/services/human_resource/cmd/wiring.go#L147) · [internal/service/registry.go L167](../../../backend/services/human_resource/internal/service/registry.go#L167) · [service/payroll/calculation.go L55](../../../backend/services/human_resource/internal/service/payroll/calculation.go#L55) · [service/payroll/verified_calculator.go L35](../../../backend/services/human_resource/internal/service/payroll/verified_calculator.go#L35) · [repository/payroll/input_snapshot_capture.sql L38](../../../backend/services/human_resource/repository/payroll/input_snapshot_capture.sql#L38).

**Kriteria penutupan.** Binary resmi dapat bergerak dari freeze ke calculated lalu review/approve dengan aturan yang sah. Snapshot pendukung yang diwajibkan tetapi belum tersedia menghentikan run.

## HRL-03 — Rekonsiliasi memeriksa baris hasil tetapi belum membuktikan kelengkapan pekerja dan digest keluaran

Service memeriksa panjang digest hasil/run dari calculator, referensi profil setiap hasil, duplikasi worker dan jumlah baris yang diserahkan. Saat approval, jumlah pembanding berasal dari hasil tersimpan. Tidak ada perbandingan dengan himpunan worker yang seharusnya digaji dari snapshot immutable atau perhitungan ulang digest keluaran di jalur itu. Sebagian pekerja yang terlewat dapat tetap menghasilkan total internal yang konsisten.

**Bukti kode.** [service/payroll/calculation.go L112](../../../backend/services/human_resource/internal/service/payroll/calculation.go#L112) · [service/payroll/calculation.go L157](../../../backend/services/human_resource/internal/service/payroll/calculation.go#L157) · [service/payroll/approval.go L44](../../../backend/services/human_resource/internal/service/payroll/approval.go#L44) · [service/payroll/service.go L82](../../../backend/services/human_resource/domain/service/payroll/service.go#L82) · `backend/services/human_resource/internal/service/payroll/finance_snapshot.go:20` (checkout).

**Kriteria penutupan.** Menghapus satu pekerja, menambah pekerja di luar snapshot atau mengubah isi tanpa mengganti digest menghentikan approval. Urutan baca yang berbeda memberi digest yang sama.

## HRL-04 — Nomor run berbeda dapat mewakili kewajiban gaji reguler yang sama

HRM mengunikkan nomor run per periode, idempotency key, dan worker per run. Validasi approval merekonsiliasi isi run itu sendiri. Finance memakai identitas HR run untuk posting. Dua run reguler berbeda yang mencakup pekerja, pemberi kerja dan periode yang sama tidak berbagi kunci kewajiban. Tidak ada bukti pembayaran ganda aktual dalam audit ini.

**Bukti kode.** [db/migrations/000001_hrm.up.sql L1687](../../../backend/services/human_resource/db/migrations/000001_hrm.up.sql#L1687) · [infrastructure/security/payroll.go L632](../../../backend/services/human_resource/infrastructure/security/payroll.go#L632) · [service/payroll/approval.go L19](../../../backend/services/human_resource/internal/service/payroll/approval.go#L19) · `backend/services/finance_accounting/internal/service/payroll/service.go:112` (checkout).

**Kriteria penutupan.** Dua approval paralel untuk gaji reguler yang sama menghasilkan paling banyak satu kewajiban efektif. Koreksi tidak melunasi kewajiban asli dua kali.

## HRL-05 — Slip untuk posting sah tanpa jurnal tertahan sewaktu terbit maupun dicetak ulang

Posting HRM mendukung finance_no_journal_posting_reference_id. Finance dapat menerbitkan payroll.published untuk posting itu. PublishPayslipForResult dan reissue mensyaratkan FinanceJournalReferenceID. Bukti publikasi yang sah tidak dapat melewati kedua pemeriksaan ini.

**Bukti kode.** [service/payroll/posting_confirmation.go L1](../../../backend/services/human_resource/internal/service/payroll/posting_confirmation.go#L1) · [service/payroll/payslip_publication.go L148](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go#L148) · [service/payroll/payslip_reissue.go L78](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L78) · `backend/services/human_resource/infrastructure/integration/finance/payroll_verifier.go:92` (checkout) · `backend/services/finance_accounting/internal/service/payroll/service.go:569` (checkout) · `backend/services/finance_accounting/internal/service/payroll/service.go:891` (checkout).

**Kriteria penutupan.** Posting sah dengan atau tanpa jurnal menghasilkan slip yang sama-sama dapat dicetak setelah Finance terbit. Retry tidak menggandakan dokumen.

## HRL-06 — Settlement menyamakan saldo nol dengan pembayaran dan belum memproyeksikan koreksi settlement

Pada net nol Finance memakai ID settlement item sebagai payment ID di result_paid. HRM menyimpannya sebagai PaymentReferenceID dan memberi status paid. Penutupan batch mensyaratkan semua hasil berstatus paid dan ber-ID pembayaran. Untuk nilai positif, Finance memeriksa total penuh. Jalur HRM yang diperiksa tidak memproyeksikan jumlah parsial atau pembatalan pembayaran yang sudah dikonfirmasi.

**Bukti kode.** `backend/services/finance_accounting/internal/service/payroll/service.go:679` (checkout) · `backend/services/finance_accounting/internal/service/payroll/service.go:977` (checkout) · [service/payroll/payment_confirmation.go L55](../../../backend/services/human_resource/internal/service/payroll/payment_confirmation.go#L55) · [service/payroll/settlement_confirmation.go L119](../../../backend/services/human_resource/internal/service/payroll/settlement_confirmation.go#L119) · [payroll/v1/payroll_event.proto L1](../../../backend/services/human_resource/contract/proto/human_resource/workload/payroll/v1/payroll_event.proto#L1).

**Kriteria penutupan.** Net nol dapat selesai tanpa referensi transfer. Sebagian pembayaran tidak menutup kewajiban. Reversal membuka kembali sisa yang benar dengan histori utuh.

## HRL-07 — Kelas kerahasiaan kasus hubungan karyawan belum membatasi pembacaan per kasus

Model membedakan restricted dan highly_restricted. Get/list memakai Resource dan ReadScope generik. Grant global yang valid mengembalikan All, tanpa filter investigator atau kelas kasus. SDL menampilkan kaitan pelapor, subjek, investigator, outcome dan referensi ringkasan. Kasus yang hanya boleh dibaca investigator tertentu belum bisa dibedakan melalui policy tersebut. Isi dokumen DRM tidak ditinjau sebagai bagian klaim ini.

**Bukti kode.** [service/employeerelations/service.go L91](../../../backend/services/human_resource/internal/service/employeerelations/service.go#L91) · [service/operation/resource.go L155](../../../backend/services/human_resource/internal/service/operation/resource.go#L155) · [infrastructure/security/policy.go L176](../../../backend/services/human_resource/infrastructure/security/policy.go#L176) · [infrastructure/security/employee_relations.go L316](../../../backend/services/human_resource/infrastructure/security/employee_relations.go#L316) · [graphql/human_resource/hrm_employeerelations_entities.graphql L1](../../../backend/gateway/graphql/human_resource/hrm_employeerelations_entities.graphql#L1).

**Kriteria penutupan.** Dua investigator dengan hak dasar serupa mendapatkan hasil dan total berbeda sesuai penugasan, konsisten di semua jalur baca.

## HRL-08 — Sejumlah mutation tersedia di Gateway tetapi verifier bisnisnya belum dipasang pada startup

Registry memiliki port untuk bukti kontrak, rekrutmen, onboarding, resolusi tiket dan suksesi. Pada jalur startup resmi tidak ditemukan pemasangan grup verifier tersebut. Handler menolak ketika dependency yang dibutuhkannya kosong. Berbeda dengan kasus ini, mandate/budget requisition dan verifier slip memang mempunyai jalur wiring.

**Bukti kode.** [internal/service/registry.go L60](../../../backend/services/human_resource/internal/service/registry.go#L60) · [human_resource/cmd/main.go L66](../../../backend/services/human_resource/cmd/main.go#L66) · [human_resource/cmd/wiring.go L155](../../../backend/services/human_resource/cmd/wiring.go#L155) · [service/personnel/contract_signature.go L44](../../../backend/services/human_resource/internal/service/personnel/contract_signature.go#L44) · [service/recruitment/offer_issue.go L89](../../../backend/services/human_resource/internal/service/recruitment/offer_issue.go#L89) · [service/recruitment/offer_response.go L32](../../../backend/services/human_resource/internal/service/recruitment/offer_response.go#L32) · [service/onboarding/complete_task.go L36](../../../backend/services/human_resource/internal/service/onboarding/complete_task.go#L36) · [service/hrservice/resolution.go L23](../../../backend/services/human_resource/internal/service/hrservice/resolution.go#L23) · [service/succession/approval.go L55](../../../backend/services/human_resource/internal/service/succession/approval.go#L55).

**Kriteria penutupan.** Setiap operasi yang membutuhkan bukti memiliki adapter konkret dan satu jalur sukses dari binary produksi. Bukti untuk pekerja, kandidat, dokumen atau versi lain ditolak.

## HRL-09 — Offer yang diterima belum memiliki perintah hire yang menghubungkan lamaran dan worker

acceptOffer merekam keputusan serta bukti respons. Schema mengenal hired_worker_id dan finder yang membacanya. Pada jalur produksi yang ditinjau tidak ditemukan perintah yang mengonsumsi offer accepted, menghubungkan worker dengan lamaran dan menyelesaikan status hired secara idempoten. createWorker/activateWorker merupakan urusan lifecycle tersendiri.

**Bukti kode.** [service/recruitment/offer_response.go L36](../../../backend/services/human_resource/internal/service/recruitment/offer_response.go#L36) · [infrastructure/security/recruitment_application.go L68](../../../backend/services/human_resource/infrastructure/security/recruitment_application.go#L68) · [service/recruitment/service.go L613](../../../backend/services/human_resource/internal/service/recruitment/service.go#L613) · [service/personnel/worker_activation.go L24](../../../backend/services/human_resource/internal/service/personnel/worker_activation.go#L24) · [graphql/human_resource/hrm_recruitment_commands.graphql L108](../../../backend/gateway/graphql/human_resource/hrm_recruitment_commands.graphql#L108).

**Kriteria penutupan.** Dua permintaan bersamaan atas offer yang sama menghasilkan satu hubungan kerja dan satu transisi hired. Hubungan lama tetap dapat ditelusuri.

## HRL-10 — Pegawai dan HR kekurangan query lookup dengan scope yang sesuai

SDL HRM memuat tiga query myPayslip*. Untuk cetak ulang, HR memerlukan worker_id dan payroll_result_id, sedangkan finder payrollResults mengakses data payroll umum. Belum ada lookup metadata slip terbit bagi HR dengan mandat cetak ulang saja. Usulan profil, kontrak dan saldo cuti pribadi juga belum memiliki root self khusus.

**Bukti kode.** [graphql/human_resource/hrm_payslip.graphql L32](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L32) · [graphql/human_resource/hrm_payroll_result.graphql L29](../../../backend/gateway/graphql/human_resource/hrm_payroll_result.graphql#L29) · [adapter/human_resource/payroll_result.go L17](../../../backend/gateway/internal/adapter/human_resource/payroll_result.go#L17) · [infrastructure/security/policy.go L126](../../../backend/services/human_resource/infrastructure/security/policy.go#L126) · [service/payroll/payslip.go L134](../../../backend/services/human_resource/internal/service/payroll/payslip.go#L134) · [service/payroll/payslip_reissue.go L31](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L31).

**Kriteria penutupan.** Dengan identitas worker saja, HR yang berwenang dapat memilih periode dan mencetak tanpa melihat nominal pegawai lain. Self-service berfungsi hanya pada data milik pegawai aktif sesuai kebijakan.

## HRL-11 — Perubahan mandat setelah transfer belum mempunyai jalur privat setara provisioning dan terminasi

HRM dapat mencatat request_kind provision/change/revoke dan transfer menghasilkan lifecycle event. Router privat employee provisioning hanya memilih create berjenis provision. Jalur terminasi memiliki kontrak tersendiri. Pencatatan change/revoke umum belum membuktikan IAM mengubah scope lama, dan keberhasilan publish umum bukan bukti hak efektif.

**Bukti kode.** [service/personnel/transfer_decision.go L103](../../../backend/services/human_resource/internal/service/personnel/transfer_decision.go#L103) · [infrastructure/security/access_request.go L108](../../../backend/services/human_resource/infrastructure/security/access_request.go#L108) · [infrastructure/outbox/publisher.go L216](../../../backend/services/human_resource/infrastructure/outbox/publisher.go#L216) · [infrastructure/outbox/store.go L108](../../../backend/services/human_resource/infrastructure/outbox/store.go#L108) · `backend/services/human_resource/infrastructure/integration/iam/employee_provisioning.go:44` (checkout).

**Kriteria penutupan.** Scope lama yang diterima untuk dicabut tidak dapat dipakai setelah acknowledgement efektif IAM. Worker tidak memperoleh grant otomatis hanya karena dipindahkan.

## HRL-12 — Pemakaian cuti yang disetujui Workforce belum dapat didebit ke ledger HRM

PostLedgerEntry memeriksa sumber lintas management dan mengunci entitlement. Validator produksi hanya menerima source_management hrm dan tipe opening, adjustment, expiry atau reversal. Usage dari Workforce tidak lolos. Policy sesi normal juga menghasilkan aktor human. Persetujuan cuti saat ini belum mempunyai jalur privat yang menghasilkan debit authoritative di HRM.

**Bukti kode.** [service/leave/ledger.go L21](../../../backend/services/human_resource/internal/service/leave/ledger.go#L21) · [infrastructure/security/leave_ledger.go L36](../../../backend/services/human_resource/infrastructure/security/leave_ledger.go#L36) · [infrastructure/security/leave_entitlement.go L151](../../../backend/services/human_resource/infrastructure/security/leave_entitlement.go#L151) · [infrastructure/security/policy.go L169](../../../backend/services/human_resource/infrastructure/security/policy.go#L169) · [leave/v1/leave.proto L1](../../../backend/services/human_resource/contract/proto/human_resource/leave/v1/leave.proto#L1).

**Kriteria penutupan.** Satu cuti approved mengurangi entitlement tepat sekali. Cancellation mengembalikan jumlah terkait, dengan referensi ke debit asli. Dua keputusan paralel tidak dapat menghabiskan saldo yang sama.

## Keputusan atas usulan desain

- **Payroll dan slip.** Pertahankan otoritas publikasi pada Finance, perhitungan dan konten slip pada HRM, penyimpanan PDF privat pada DRM, dan otorisasi pada IAM. Pegawai aktif mengambil slip setelah publikasi meskipun transfer masih pending. Mantan pegawai dilayani HR. Histori slip yang dikoreksi memerlukan versi baru yang mengacu versi sebelumnya. Pembacaan slip hilang mengikuti alur restore yang sudah tersedia. [Reissue dan pemulihan](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go) · `backend/services/finance_accounting/internal/service/payroll/service.go:569` (checkout).
- **HR, EF dan IAM.** HRM memakai party individual aktif dari EF dan menyimpan hubungan kerja. Pertahankan provisioning melalui IAM `createEmployeeUser` dan verifikasi exact worker. Reviewer requisition membutuhkan mandat IAM unit serta otorisasi anggaran Finance. Terminasi efektif memicu pencabutan tanpa menunggu settlement akhir. Penutupan offboarding menuntut bukti owner. Petugas HR berwenang mencatat konfirmasi untuk pekerjaan milik HRM. [Verifikasi party EF](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L198) · [Terminasi](../../../backend/services/human_resource/domain/service/personnel/termination.go) · [Konfirmasi HR](../../../backend/services/human_resource/internal/service/offboarding/complete_task.go) · [Clearance owner](../../../backend/services/human_resource/internal/service/offboarding/owner_clearance.go).
- **Workforce, benefit dan cuti.** HRM mengelola hak. Workforce menguasai waktu kerja aktual dan permohonan cuti. Hubungkan sumber yang diwajibkan aturan perusahaan melalui kontrak privat. Tanpa owner yang sudah berjalan, pertahankan pending untuk proses yang memerlukannya. Pembentukan management baru berada di luar remediasi ini.
- **Performance, career, learning, succession, talent dan analytics.** Pertahankan akses hasil sesuai fase release. Pisahkan ringkasan untuk pegawai dari kalibrasi internal, data kasus dan dokumen rahasia. Dashboard membedakan nominal calculated HRM, posting Finance dan realisasi pembayaran. Usulan variasi report atau 28 service konseptual belum menjadi kebutuhan menambah semua root baru. Pemetaan semua domain beserta status nonaktif tersedia di katalog. [Katalog domain HRM](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md).
- **Pekerjaan terdahulu.** Kalkulator produksi pernah dicatat dalam handoff 27. Instruksi 8–10 membahas restore, IAM, anggaran dan publikasi. Temuan kali ini menyebut titik kode yang masih memerlukan perubahan. Keberadaan port, katalog atau laporan historis tidak digunakan sebagai bukti runtime berhasil. [Kalkulator versi sebelumnya](../../../agents/grpc/hrm/27-kebutuhan-kalkulator-payroll-runtime.md) · [Handoff 8](../../../agents/instructions/hrm/8-leader-remediasi-code-inti-hrm-instructions.md) · [Handoff 9](../../../agents/instructions/hrm/9-leader-remediasi-hubungan-iam-hrm-instructions.md) · [Handoff 10](../../../agents/instructions/hrm/10-leader-remediasi-procurement-hrm-finance-instructions.md).

Semua HRL berstatus **terbuka untuk implementasi**. [Instruksi agent dan urutan integrasi](../../../agents/instructions/hrm/11-leader-remediasi-logika-operasi-dan-otorisasi-hrm-instructions.md).
