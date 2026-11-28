# Remediasi HRM dari database sampai Gateway dan pemilik bukti

- **owner** — agent HRM memimpin. Agent Finance, IAM, EF, DRM dan Gateway memegang implementasi boundary masing-masing. Libatkan owner Workforce untuk fakta waktu dan cuti.
- **priority** — tutup HRL-01–06 untuk alur payroll. Kerjakan HRL-07 dan HRL-11 paralel karena menyangkut kerahasiaan dan pencabutan akses.
- **status** — siap dikerjakan. Semua temuan masih terbuka.
- **baseline** — checkout bersama pada `738c5490db08c81d8d748ccf07b950939a6edfc5`, 29 November 2026 WIB. [Laporan audit dan bukti produksi](../../../agents/audit/human_resource/199-2026-11-29-000433-audit-logika-operasi-dan-otorisasi-hrm.md).

## Task

Selesaikan 12 temuan sampai kriteria penerimaan masing-masing terbukti pada jalur pemanggil nyata. Gunakan izin dan kontrak yang sudah ada bila memadai. Katalog baru membedakan operasi publik sekarang, usulan query/mutation baru, dan jalur workload. [Matriks operasi](../../../agents/grpc/hrm/45-matriks-operasi-publik-dan-kebutuhan-domain.md).

## Request

### HRL-01 — Kompensasi yang disetujui dapat hilang dari snapshot gaji

- file: [service/compensation/approval.go L31](../../../backend/services/human_resource/internal/service/compensation/approval.go#L31) · [infrastructure/security/compensation.go L164](../../../backend/services/human_resource/infrastructure/security/compensation.go#L164) · [repository/payroll/input_snapshot_capture.sql L18](../../../backend/services/human_resource/repository/payroll/input_snapshot_capture.sql#L18) · [graphql/human_resource/hrm_compensation_commands.graphql L10](../../../backend/gateway/graphql/human_resource/hrm_compensation_commands.graphql#L10).
- function: approveCompensationRecord, FreezePayrollInput, Capture.
- logic: Jadikan approval sah dan rentang efektif sebagai sumber hak kompensasi. Pilihan implementasi yang disarankan adalah memasukkan approved yang berlaku ke snapshot. Jika aktivasi terpisah memang diperlukan, sediakan transisi efektif yang dapat dijalankan dan diaudit. Tangani effective_from/to, supersesi di tengah periode, overlap serta data lama tanpa menghitung komponen dua kali.
- error_handling: Tahan freeze untuk pekerja yang wajib menerima komponen tetapi sumbernya ambigu. Kegagalan membatalkan snapshot.
- security: Pertahankan pemisahan pengusul dan penyetuju serta larangan menyetujui kompensasi sendiri.
- expected_result: Periode setelah tanggal efektif memuat kompensasi approved tepat sekali, periode sebelumnya tidak memuatnya.
- testing: Buat proposal, setujui dengan petugas berbeda, freeze dua periode, cocokkan jumlah serta ID komponen. Ulangi untuk supersesi tengah periode.

### HRL-02 — Kalkulator payroll belum terpasang di binary HRM

- file: [human_resource/cmd/main.go L59](../../../backend/services/human_resource/cmd/main.go#L59) · [human_resource/cmd/wiring.go L147](../../../backend/services/human_resource/cmd/wiring.go#L147) · [internal/service/registry.go L167](../../../backend/services/human_resource/internal/service/registry.go#L167) · [service/payroll/calculation.go L55](../../../backend/services/human_resource/internal/service/payroll/calculation.go#L55) · [service/payroll/verified_calculator.go L35](../../../backend/services/human_resource/internal/service/payroll/verified_calculator.go#L35) · [repository/payroll/input_snapshot_capture.sql L38](../../../backend/services/human_resource/repository/payroll/input_snapshot_capture.sql#L38).
- function: calculatePayrollRun, NewVerifiedCalculator, Capture, buildDependenciesWithAllPublishers.
- logic: Pasang reader snapshot dan artefak DRM serta implementasi mesin deterministik sesuai nama/versi rule set. Dokumentasikan komponen yang didukung. Untuk aturan berbasis waktu/benefit, ambil fakta owner yang approved pada cutoff dan simpan versi/digest di snapshot. Payroll gaji tetap boleh berjalan bila kebijakannya tidak memerlukan Workforce. Tidak ada tarif legal atau jam kerja asumsi.
- error_handling: Dependency wajib yang belum tersedia menghasilkan Unavailable atau FailedPrecondition berikut identitas tahap aman. Simpan draf agar retry tidak menggandakan hasil.
- security: Artefak immutable, workload mTLS dengan allowlist, digest dari byte aktual. Batasi format aturan ke operasi yang didukung dan jangan mengeksekusi kode dari dokumen.
- expected_result: Binary resmi dapat bergerak dari freeze ke calculated lalu review/approve dengan aturan yang sah. Snapshot pendukung yang diwajibkan tetapi belum tersedia menghentikan run.
- testing: Jalankan service dengan mock_db dan vektor deterministik, lalu HTTPS pada binary resmi dengan DRM/owner fakta. Sertakan replay, pembulatan, sumber terlambat, digest berbeda, dan bukti hasil per pekerja.

### HRL-03 — Rekonsiliasi memeriksa baris hasil tetapi belum membuktikan kelengkapan pekerja dan digest keluaran

- file: [service/payroll/calculation.go L112](../../../backend/services/human_resource/internal/service/payroll/calculation.go#L112) · [service/payroll/calculation.go L157](../../../backend/services/human_resource/internal/service/payroll/calculation.go#L157) · [service/payroll/approval.go L44](../../../backend/services/human_resource/internal/service/payroll/approval.go#L44) · [service/payroll/service.go L82](../../../backend/services/human_resource/domain/service/payroll/service.go#L82) · `backend/services/human_resource/internal/service/payroll/finance_snapshot.go:20` (checkout).
- function: CalculatePayrollRun, ValidateRunResults, validateRunResults, GetApprovedPayrollSnapshot.
- logic: Definisikan roster dan exclusion per run type saat freeze. Rekonsiliasi dua arah dengan hasil per worker dan alasan pengecualian. Tetapkan format kanonik berversi untuk identitas worker/result, komponen, amount desimal, currency, periode, input dan rule. Hitung ulang digest dari isi yang akan dipersistenkan dan pastikan snapshot privat Finance memuat versi dan isi yang sama.
- error_handling: Roster/digest berbeda menghasilkan konflik integritas, tanpa result, approval atau outbox separuh jadi.
- security: Approval independen dan recheck tetap wajib. Kembalikan jumlah selisih dan correlation ID kepada operator, simpan rincian hanya di kanal terbatas.
- expected_result: Menghapus satu pekerja, menambah pekerja di luar snapshot atau mengubah isi tanpa mengganti digest menghentikan approval. Urutan baca yang berbeda memberi digest yang sama.
- testing: Kasus dua pekerja, output satu pekerja, baris komponen yatim, amount berubah dengan total tetap, urutan terbalik, off-cycle terbatas dan finance read-back.

### HRL-04 — Nomor run berbeda dapat mewakili kewajiban gaji reguler yang sama

- file: [db/migrations/000001_hrm.up.sql L1687](../../../backend/services/human_resource/db/migrations/000001_hrm.up.sql#L1687) · [infrastructure/security/payroll.go L632](../../../backend/services/human_resource/infrastructure/security/payroll.go#L632) · [service/payroll/approval.go L19](../../../backend/services/human_resource/internal/service/payroll/approval.go#L19) · `backend/services/finance_accounting/internal/service/payroll/service.go:112` (checkout).
- function: createPayrollRun, approvePayrollRun, requestPayrollPosting dan PostPayrollAccrual.
- logic: Pada transisi pertama yang menerbitkan kewajiban, klaim kunci bisnis stabil per hubungan kerja, employer, periode dan jenis kewajiban reguler. Lindungi dengan constraint atomik dan transaksi. Tetap izinkan draf simulasi, bonus, off-cycle dan koreksi terotorisasi yang menunjuk kewajiban asal dan delta. Rekonsiliasi histori duplikat sebelum penerapan. Finance memeriksa identitas kewajiban pada snapshot untuk pertahanan tambahan.
- error_handling: Retry intent identik mendapatkan hasil lama. Run berbeda yang mengeklaim kewajiban yang sama mendapat konflik sebelum event posting.
- security: HRM mengajukan rincian approved. Jurnal dan instruksi pembayaran tetap hanya dapat diterbitkan Finance.
- expected_result: Dua approval paralel untuk gaji reguler yang sama menghasilkan paling banyak satu kewajiban efektif. Koreksi tidak melunasi kewajiban asli dua kali.
- testing: Dua run/nomor/key berbeda, worker overlap sebagian, dua legal entity, bonus sah, correction/reversal dan replay posting Finance.

### HRL-05 — Slip untuk posting sah tanpa jurnal tertahan sewaktu terbit maupun dicetak ulang

- file: [service/payroll/posting_confirmation.go L1](../../../backend/services/human_resource/internal/service/payroll/posting_confirmation.go#L1) · [service/payroll/payslip_publication.go L148](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go#L148) · [service/payroll/payslip_reissue.go L78](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L78) · `backend/services/human_resource/infrastructure/integration/finance/payroll_verifier.go:92` (checkout) · `backend/services/finance_accounting/internal/service/payroll/service.go:569` (checkout) · `backend/services/finance_accounting/internal/service/payroll/service.go:891` (checkout).
- function: PublishPayrollRun (Finance), HandlePayrollPublication, PublishPayslipForRun, VerifyPayrollPublication, FindPayrollPublicationByRun, reissueEmployeePayslip.
- logic: Pakai referensi posting terverifikasi yang membedakan journal_id dan no_journal_posting_id secara eksklusif. Teruskan varian yang tepat ke verifier Finance dan lookup reissue. Publikasi Finance memicu pembuatan PDF privat per result di DRM. Pegawai aktif mencetak sendiri, mantan pegawai melalui HR berwenang. Jika byte hilang, pulihkan versi persis setelah verifikasi metadata, digest dan izin.
- error_handling: Bukti tidak cocok ditolak. DRM belum siap menghasilkan pending/retry. Kegagalan sementara, izin ditolak, dan objek benar-benar hilang harus dapat dibedakan.
- security: Grant self berasal dari principal aktif. HR membutuhkan scope employer serta MFA. Publikasi tidak menunggu transfer dan tidak mengubah status paid.
- expected_result: Posting sah dengan atau tanpa jurnal menghasilkan slip yang sama-sama dapat dicetak setelah Finance terbit. Retry tidak menggandakan dokumen.
- testing: Verifikasi dua varian posting, publikasi sebelum bayar, pegawai lain, mantan pegawai melalui HR, dokumen byte hilang serta timeout DRM.

### HRL-06 — Settlement menyamakan saldo nol dengan pembayaran dan belum memproyeksikan koreksi settlement

- file: `backend/services/finance_accounting/internal/service/payroll/service.go:679` (checkout) · `backend/services/finance_accounting/internal/service/payroll/service.go:977` (checkout) · [service/payroll/payment_confirmation.go L55](../../../backend/services/human_resource/internal/service/payroll/payment_confirmation.go#L55) · [service/payroll/settlement_confirmation.go L119](../../../backend/services/human_resource/internal/service/payroll/settlement_confirmation.go#L119) · [payroll/v1/payroll_event.proto L1](../../../backend/services/human_resource/contract/proto/human_resource/workload/payroll/v1/payroll_event.proto#L1).
- function: CompletePayrollSettlement, verifiedPaymentID, ConfirmPayrollPayment, ConfirmPayrollSettlement, DeliverPayrollEvent.
- logic: Pisahkan status kewajiban, jumlah terbayar, sisa, bukti settlement nol dan bukti transaksi uang. Finance menerbitkan fakta per result dengan amount, versi urut, mata uang, ID settlement dan referensi pembayaran hanya jika terjadi transfer. Proyeksikan parsial dan reversal dari owner, pertahankan histori dan kaitan dengan slip koreksi. Tangani event lama secara kompatibel melalui rekonsiliasi Finance.
- error_handling: Cegah duplikasi dan urutan lama menimpa revisi baru. Tahan event jika Finance belum bisa membuktikan nilai. Tutup batch hanya ketika semua sisa nol dengan bukti yang sesuai.
- security: Hanya workload Finance berwenang. Hindari mutation publik untuk mengesahkan pembayaran dari HRM.
- expected_result: Net nol dapat selesai tanpa referensi transfer. Sebagian pembayaran tidak menutup kewajiban. Reversal membuka kembali sisa yang benar dengan histori utuh.
- testing: Uji campuran net nol dan positif, dua cicilan, retur setelah lunas, replay, event tidak berurutan dan keseluruhan batch. Periksa data yang ditampilkan di slip.

### HRL-07 — Kelas kerahasiaan kasus hubungan karyawan belum membatasi pembacaan per kasus

- file: [service/employeerelations/service.go L91](../../../backend/services/human_resource/internal/service/employeerelations/service.go#L91) · [service/operation/resource.go L155](../../../backend/services/human_resource/internal/service/operation/resource.go#L155) · [infrastructure/security/policy.go L176](../../../backend/services/human_resource/infrastructure/security/policy.go#L176) · [infrastructure/security/employee_relations.go L316](../../../backend/services/human_resource/infrastructure/security/employee_relations.go#L316) · [graphql/human_resource/hrm_employeerelations_entities.graphql L1](../../../backend/gateway/graphql/human_resource/hrm_employeerelations_entities.graphql#L1).
- function: employeeRelationsCase, employeeRelationsCases, node, pembacaan event turunan dan total_count.
- logic: Tetapkan siapa boleh membaca tiap kelas. Terapkan ACL kasus/investigator/delegasi atau izin oversight eksplisit yang disetujui, sebelum select, count dan pagination. Event mewarisi kebijakan kasus induk. Redaksi metadata pelapor, legal hold dan referensi bila tidak diperlukan bagi aktor. Catat akses sensitif dan pencabutan penugasan.
- error_handling: Objek di luar lingkup tidak membocorkan keberadaan lewat detail, node, filter maupun count. Tolak ketika data mandat tidak tersedia.
- security: Permission global HR biasa dan MFA tidak menggantikan mandat investigasi terbatas. Tetap sediakan jalur oversight yang eksplisit, tercatat dan dapat dicabut.
- expected_result: Dua investigator dengan hak dasar serupa mendapatkan hasil dan total berbeda sesuai penugasan, konsisten di semua jalur baca.
- testing: Kasus restricted/highly_restricted, investigator ditugaskan/dicabut, petugas HR umum, delegasi sementara dan oversight. Bandingkan detail, node, event, filter dan count.

### HRL-08 — Sejumlah mutation tersedia di Gateway tetapi verifier bisnisnya belum dipasang pada startup

- file: [internal/service/registry.go L60](../../../backend/services/human_resource/internal/service/registry.go#L60) · [human_resource/cmd/main.go L66](../../../backend/services/human_resource/cmd/main.go#L66) · [human_resource/cmd/wiring.go L155](../../../backend/services/human_resource/cmd/wiring.go#L155) · [service/personnel/contract_signature.go L44](../../../backend/services/human_resource/internal/service/personnel/contract_signature.go#L44) · [service/recruitment/offer_issue.go L89](../../../backend/services/human_resource/internal/service/recruitment/offer_issue.go#L89) · [service/recruitment/offer_response.go L32](../../../backend/services/human_resource/internal/service/recruitment/offer_response.go#L32) · [service/onboarding/complete_task.go L36](../../../backend/services/human_resource/internal/service/onboarding/complete_task.go#L36) · [service/hrservice/resolution.go L23](../../../backend/services/human_resource/internal/service/hrservice/resolution.go#L23) · [service/succession/approval.go L55](../../../backend/services/human_resource/internal/service/succession/approval.go#L55).
- function: recordSignedEmploymentContract, publishJobPosting, completeApplicationAssessment, approveOffer, issueOffer, acceptOffer, declineOffer, completeOnboardingTask, resolveServiceRequest, approveSuccessionPlan.
- logic: Inventaris kondisi yang membutuhkan verifier untuk setiap operasi. Sambungkan bukti pemilik tepat sasaran ke registry produksi. DRM memverifikasi versi, digest, pihak dan tujuan dokumen. HRM memverifikasi kompensasi dan mandat suksesi sebagai pemilik, IAM memberi fakta otorisasi. Bedakan bukti opsional dan wajib. Untuk tugas offboarding yang dimiliki HRM, pertahankan kemampuan petugas berwenang mencatat konfirmasi lokal.
- error_handling: Jika dependency wajib sementara tidak tersedia, kembalikan Unavailable dan izinkan retry aman. Jangan mengubah status dengan verifier tiruan.
- security: Identitas workload privat, MFA dan SoD pemutus tetap berlaku. Jangan menganggap UUID bukti dari klien sebagai keputusan owner.
- expected_result: Setiap operasi yang membutuhkan bukti memiliki adapter konkret dan satu jalur sukses dari binary produksi. Bukti untuk pekerja, kandidat, dokumen atau versi lain ditolak.
- testing: Satu happy path dan satu bukti salah per command pada scope terdampak. Uji putus koneksi sesudah owner menerima, replay, bukti expired dan ketidakcocokan digest.

### HRL-09 — Offer yang diterima belum memiliki perintah hire yang menghubungkan lamaran dan worker

- file: [service/recruitment/offer_response.go L36](../../../backend/services/human_resource/internal/service/recruitment/offer_response.go#L36) · [infrastructure/security/recruitment_application.go L68](../../../backend/services/human_resource/infrastructure/security/recruitment_application.go#L68) · [service/recruitment/service.go L613](../../../backend/services/human_resource/internal/service/recruitment/service.go#L613) · [service/personnel/worker_activation.go L24](../../../backend/services/human_resource/internal/service/personnel/worker_activation.go#L24) · [graphql/human_resource/hrm_recruitment_commands.graphql L108](../../../backend/gateway/graphql/human_resource/hrm_recruitment_commands.graphql#L108).
- function: Usulan mutation hireAcceptedApplicant, RPC HireAcceptedApplicant, acceptOffer dan createWorker.
- logic: Kunci application dan offer, cocokkan identitas kandidat/party individual dengan EF, cek requisition/position/headcount dan bukti accepted yang masih sah. Buat atau tautkan relasi kerja yang benar serta tahap awal onboarding, lalu tandai application hired dan catat event atomik. Pisahkan rehire dari duplikasi worker dan lindungi retry dengan constraint. IAM provisioning tetap dimulai dari pengajuan dan createEmployeeUser(worker_id) sesuai alur sekarang.
- error_handling: Offer belum accepted atau referensi tidak cocok menghasilkan FailedPrecondition. Retry identik mengembalikan worker yang sama. Konflik identity masuk peninjauan HR.
- security: Recruiter tidak boleh membuat principal, memberi role, mengaktifkan gaji atau mengesahkan kontrak lewat hire.
- expected_result: Dua permintaan bersamaan atas offer yang sama menghasilkan satu hubungan kerja dan satu transisi hired. Hubungan lama tetap dapat ditelusuri.
- testing: Kasus accepted/expired/withdrawn, kandidat internal, rehire, dua panggilan paralel dan owner EF tidak tersedia.

### HRL-10 — Pegawai dan HR kekurangan query lookup dengan scope yang sesuai

- file: [graphql/human_resource/hrm_payslip.graphql L32](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L32) · [graphql/human_resource/hrm_payroll_result.graphql L29](../../../backend/gateway/graphql/human_resource/hrm_payroll_result.graphql#L29) · [adapter/human_resource/payroll_result.go L17](../../../backend/gateway/internal/adapter/human_resource/payroll_result.go#L17) · [infrastructure/security/policy.go L126](../../../backend/services/human_resource/infrastructure/security/policy.go#L126) · [service/payroll/payslip.go L134](../../../backend/services/human_resource/internal/service/payroll/payslip.go#L134) · [service/payroll/payslip_reissue.go L31](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L31).
- function: Usulan query employeePayslipPeriods(worker_id, pagination), myEmploymentProfile, myPublishedContracts, myLeaveBalances. Pertahankan myPayslips, myPayslipPdf dan reissueEmployeePayslip..
- logic: Lookup HR membatasi periode terbit menurut worker dan employer sebelum count. Berikan result ID opaque, label periode dan publikasi yang dibutuhkan UI. Query self memetakan principal dari sesi ke hubungan kerja efektif dan menyajikan proyeksi minimal, kontrak published serta saldo cuti dengan satuan. Histori rehire mengikuti identitas orang yang sudah dikonfirmasi. Daftar pribadi performance/learning dapat mengikuti setelah aturan release disetujui.
- error_handling: Pengguna di luar scope mendapat respons tidak mengungkap keberadaan data. Ketidakcocokan relasi principal/worker gagal tertutup.
- security: Jangan memperluas permission payroll global untuk HR pencetak atau employee. Lindungi detail/Relay edges dan total dengan filter sama. Mantan karyawan meminta cetak melalui HR.
- expected_result: Dengan identitas worker saja, HR yang berwenang dapat memilih periode dan mencetak tanpa melihat nominal pegawai lain. Self-service berfungsi hanya pada data milik pegawai aktif sesuai kebijakan.
- testing: Uji lookup antar employer, dua worker satu nama, rehire, pegawai aktif vs mantan pegawai, pagination kosong dan akses dari node generik.

### HRL-11 — Perubahan mandat setelah transfer belum mempunyai jalur privat setara provisioning dan terminasi

- file: [service/personnel/transfer_decision.go L103](../../../backend/services/human_resource/internal/service/personnel/transfer_decision.go#L103) · [infrastructure/security/access_request.go L108](../../../backend/services/human_resource/infrastructure/security/access_request.go#L108) · [infrastructure/outbox/publisher.go L216](../../../backend/services/human_resource/infrastructure/outbox/publisher.go#L216) · [infrastructure/outbox/store.go L108](../../../backend/services/human_resource/infrastructure/outbox/store.go#L108) · `backend/services/human_resource/infrastructure/integration/iam/employee_provisioning.go:44` (checkout).
- function: DecideWorkerAssignmentTransfer, CreateAccessRequest, relay privat change/revoke dan lookup hasil IAM.
- logic: Terbitkan intent idempoten untuk perubahan mandat dengan worker/principal, versi hubungan, effective_at, scope lama/baru dan katalog sah. IAM mengevaluasi kebijakan dan menjadi satu-satunya pihak yang mengubah grant. Bedakan submitted, accepted, effective dan refused. Rekonsiliasi hasil exact request setelah timeout dan pertahankan jalur termination khusus. Rehire memakai identity yang diverifikasi ulang.
- error_handling: Outbox yang belum memiliki tujuan privat tetap pending. IAM unavailable tidak menghasilkan status effective. Tutup intent usang saat ada versi hubungan baru.
- security: Jangan menambahkan mutation HR untuk menghubungkan worker ke principal. Grant baru membutuhkan persetujuan yang sesuai dan pencabutan harus memengaruhi sesi yang relevan.
- expected_result: Scope lama yang diterima untuk dicabut tidak dapat dipakai setelah acknowledgement efektif IAM. Worker tidak memperoleh grant otomatis hanya karena dipindahkan.
- testing: Transfer, promosi yang mengubah mandat, rehire, penolakan role baru, akun tidak aktif, replay dan event datang terbalik. Verifikasi dengan sesi lama serta lookup privat IAM.

### HRL-12 — Pemakaian cuti yang disetujui Workforce belum dapat didebit ke ledger HRM

- file: [service/leave/ledger.go L21](../../../backend/services/human_resource/internal/service/leave/ledger.go#L21) · [infrastructure/security/leave_ledger.go L36](../../../backend/services/human_resource/infrastructure/security/leave_ledger.go#L36) · [infrastructure/security/leave_entitlement.go L151](../../../backend/services/human_resource/infrastructure/security/leave_entitlement.go#L151) · [infrastructure/security/policy.go L169](../../../backend/services/human_resource/infrastructure/security/policy.go#L169) · [leave/v1/leave.proto L1](../../../backend/services/human_resource/contract/proto/human_resource/leave/v1/leave.proto#L1).
- function: PostLedgerEntry, validator ledger, usulan DeliverApprovedLeaveUsage dan konfirmasi hasil ke Workforce.
- logic: Workforce tetap pemilik permohonan, jadwal serta keputusan cuti. Sediakan inbox gRPC privat tervalidasi yang mengonsumsi bukti approved/cancelled. Cocokkan worker, employer, policy, periode, satuan, debit dan versi. Terapkan debit/reversal dalam transaksi yang sama dengan idempotency key sumber dan proyeksi saldo. Apabila owner belum tersedia, sediakan kontrak dan pertahankan status pending yang dapat direkonsiliasi. Jalur penyesuaian human tetap terpisah.
- error_handling: Saldo tidak cukup, satuan tidak cocok atau periode tertutup tidak memotong hak. Replay dengan isi sama mengembalikan hasil lama, digest berbeda ditolak.
- security: Hanya workload owner berwenang mengajukan usage. Jangan mengizinkan sumber eksternal dengan hanya mempercayai string source_management dari mutation umum.
- expected_result: Satu cuti approved mengurangi entitlement tepat sekali. Cancellation mengembalikan jumlah terkait, dengan referensi ke debit asli. Dua keputusan paralel tidak dapat menghabiskan saldo yang sama.
- testing: Service menguji debit, reversal, unit berbeda, replay dan balapan saldo terakhir. Jalankan integrasi owner→HRM dan konfirmasi saldo dari query self.

## Action

1. Cocokkan setiap temuan dengan revisi terbaru dan catat SHA serta perubahan lokal yang dipakai. Baca usulan, audit, matriks dan prior handoff. Selaraskan dahulu identitas kewajiban, semantik settlement, versi event dan arti publikasi dengan Finance. Pastikan lookup pribadi serta hak petugas HR jelas di katalog IAM.
2. Mulai dari DBML, invariant PostgreSQL, skema up/down `000001`, model dan repository. Periksa data yang sudah ada, hasil regular ganda, status approved historis, ID payment sintetis dan posting tanpa jurnal. Beri langkah pemulihan yang dapat diulang, pertahankan ID dan bukti versi lama. Lanjutkan ke domain, service, proto, handler, registry, client workload dan binary resmi.
3. Untuk tiap RPC, perbarui daftar `agents/grpc/hrm/`, handler otorisasi serta konsumennya. Di Gateway ikuti SDL dasar yang sudah dipakai IAM/EF untuk root, payload, input, Node, edges, cursor dan total. Hindari nama operasi/type publik berprefiks HRM. UI menampilkan kode, nama dan periode, memakai ID opaque hanya untuk navigasi/aksi. Jangan membuka RPC workload sebagai root publik.
4. Uji gRPC di lapisan **service** menggunakan `mock_db` sesuai repo. Verifikasi efek constraint, concurrency, inbox/outbox dan wiring melalui E2E stack resmi dengan database disposable `*_test`. Jalankan formatter, vet, test race dan generator untuk modul yang diubah. Login superadmin lewat HTTPS `/api/graphql`, atur role, permission dan assignment melalui GraphQL IAM. Pakai pegawai, HR berscope, reviewer independen dan workload privat untuk membuktikan keputusan tiap peran. Pertahankan MFA, CSRF, identitas workload dan recheck sebelum commit.
5. Kerjakan pemulihan retry dan dependency timeout bersama pemilik data. Lampirkan matriks hasil tiap HRL, versi rules/snapshot tersanitasi, status outbox/inbox serta hasil baca kembali. Perbarui instruksi operasi, E2E, history dan konteks domain. Tandai selesai sesudah binary resmi memenuhi semua kriteria, termasuk jalur gaji, slip, hire dan hak akses. Laporkan secara eksplisit dependency yang belum dapat dibuktikan.

## Context

**Urutan baca agent**

1. [Laporan audit dengan file dan fungsi](../../../agents/audit/human_resource/199-2026-11-29-000433-audit-logika-operasi-dan-otorisasi-hrm.md) → [Matriks root dan perubahan kontrak](../../../agents/grpc/hrm/45-matriks-operasi-publik-dan-kebutuhan-domain.md) → [Inventaris lengkap per service](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md). Dokumen usulan `agents/logic/human_resource/logic.md` dan template `agents/instruction_template.md` berada di checkout bersama.
2. [Aturan repositori](../../../AGENTS.md) · [Keputusan D01–D12](../../../agents/arsitektur.md) · [Kepemilikan](../../../agents/boundary.md) · [User, role dan scope](../../../agents/users.md) · [Konteks HRM](../../../agents/context/human_resource/1-context.md) · [Kebutuhan HRM dan DRM](../../../docs/03-managements/documents-and-hrm.md).
3. [how_to_code gRPC](../../../agents/how_to_code_grpc_managements.md) · [how_to_code wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [SDL dasar](../../../backend/gateway/graphql/schema.graphql) · [DBML HRM](../../../backend/services/human_resource/db/schema/human_resource_management.dbml) · [Invariant PostgreSQL](../../../backend/services/human_resource/db/schema/postgresql-invariants.sql).
4. **Payroll/Finance/DRM** — [Kalkulator](../../../agents/grpc/hrm/27-kebutuhan-kalkulator-payroll-runtime.md) · [Slip mantan karyawan](../../../agents/grpc/hrm/39-cetak-ulang-slip-gaji-mantan-karyawan.md) · [Event Finance](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md) · `agents/grpc/hrm/44-snapshot-payroll-untuk-finance.md` (checkout) · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto` (checkout) · `backend/services/human_resource/infrastructure/integration/document_records/drm_client.go` (checkout).
5. **IAM/EF/ketenagakerjaan** — [Proto provisioning IAM](../../../backend/services/iam/contract/proto/iam/v1/employee_provisioning.proto) · [Verifier EF](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go) · [Handoff HRM 8](../../../agents/instructions/hrm/8-leader-remediasi-code-inti-hrm-instructions.md) · [Handoff HRM 9](../../../agents/instructions/hrm/9-leader-remediasi-hubungan-iam-hrm-instructions.md) · [Handoff HRM 10](../../../agents/instructions/hrm/10-leader-remediasi-procurement-hrm-finance-instructions.md). File produksi spesifik setiap pekerjaan ada pada field `file` di atas.
