# List operasi

Terdapat **176 Query** dan **326 Mutation** khusus HRM di schema Gateway (`86` tipe `Node` dan `86` `Connection` Relay), menutup seluruh operasi aktif dari **1.016 RPC bisnis** pada **22 service gRPC privat** (dikurangi 18 RPC CRUD nonaktif untuk 6 tabel read-only serta RPC privat/internal seperti `LinkWorkerPrincipal`, `DeliverPayrollEvent`, `VerifyTerminationAccess`, `VerifyWorkerAccountLink`). [Inventaris RPC](1-inventaris-operasi-grpc-hrm.md) · [Laporan E2E Penuh](../../e2e/gateway/185-hrm-query-mutation-penuh-dan-guard-https.md).

Seluruh **176/176 Query** dan **326/326 Mutation** telah lulus guard SDL (`check_hrm_schema_guards.py`: 0 prefiks `hrm`, kontrak `Node`/`Connection`/`MutationResult`) dan guard HTTPS 401/403 pada `erp.localhost:18443`, serta ter-introspect pada `https://erp.localhost:8443/api/graphql`. Jalur bisnis yang bergantung pada runtime Finance & DRM tetap fail-closed sampai service pemilik siap. [Query](31-bukti-https-query-gateway.md) · [Mutation](30-bukti-https-mutation-gateway.md)


Lihat [audit cakupan per service](29-audit-cakupan-wiring-gateway.md) untuk perbandingan dengan inventaris gRPC.

## Penjelasan Query

| Query | RPC | Izin dengan prefiks `human_resource.` | Scope |
| --- | --- | --- | --- |
| `workerEmployeeAccountRequest` | `GetWorkerByID`, `FindAccessRequestByIdempotencyKey` | `access_requests.read` dan `workers.read` | Global untuk pengajuan, pekerja mengikuti scope baca. Mengembalikan status IAM atau null jika belum pernah diajukan. |
| `employeeAccountRequest` | `GetAccessRequestByID`, `GetWorkerByID` | `access_requests.read` dan `workers.read` | Global untuk pengajuan, worker mengikuti scope baca |
| `worker`, `node` dengan GID pekerja | `GetWorkerByID` | `workers.read` | global atau badan hukum |
| `employmentContract` | `GetEmploymentContractByID` | `employment_contracts.read` | global dan MFA |
| `employmentContracts` | `FindEmploymentContractByWorkerID` | `employment_contracts.read` | global dan MFA |
| `offboardingPlan` | `GetOffboardingPlanByID` | `offboarding_plans.read` | global |
| `offboardingPlans` | `FindOffboardingPlanByWorkerID` | `offboarding_plans.read` | global |
| `offboardingTask` | `GetOffboardingTaskByID` | `offboarding_tasks.read` | global |
| `offboardingTasks` | `FindOffboardingTaskByOffboardingPlanID` | `offboarding_tasks.read` | global |
| `workerAssignment` | `GetWorkerAssignmentByID` | `worker_assignments.read` | global |
| `workerAssignments` | `FindWorkerAssignmentByWorkerID`, `FindWorkerAssignmentByPositionID` | `worker_assignments.read` | global |
| `workers` | `GetAllWorkers`, `FindWorkerByWorkerNumber`, `FindWorkerByPrimaryLegalEntityReferenceID` | `workers.read` | global atau badan hukum |
| `workforcePlan` | `GetWorkforcePlanByID` | `workforce_plans.read` | global atau badan hukum |
| `workforcePlans` | `GetAllWorkforcePlans` | `workforce_plans.read` | global atau badan hukum |
| `job` | `GetJobByID` | `jobs.read` | global |
| `jobs` | `GetAllJobs`, `FindJobByJobFamilyID` | `jobs.read` | global |
| `jobFamily` | `GetJobFamilyByID` | `job_families.read` | global |
| `jobFamilies` | `GetAllJobFamilies` | `job_families.read` | global |
| `position` | `GetPositionByID` | `positions.read` | global |
| `positions` | `GetAllPositions`, `FindPositionByLegalEntityReferenceID` | `positions.read` | global |
| `workforcePlanLine` | `GetWorkforcePlanLineByID` | `workforce_plan_lines.read` | global atau badan hukum |
| `workforcePlanLines` | `GetAllWorkforcePlanLines`, `FindWorkforcePlanLineByWorkforcePlanID` | `workforce_plan_lines.read` | global atau badan hukum |
| `payrollRun` | `GetPayrollRunByID` | `payroll_runs.read` | global |
| `payrollResult` | `GetPayrollResultByID` | `payroll_results.read` | global dengan MFA terkini |
| `payrollResults` | `FindPayrollResultByPayrollRunID` | `payroll_results.read` | global dengan MFA terkini |
| `payrollResultLine` | `GetPayrollResultLineByID` | `payroll_result_lines.read` | global dengan MFA terkini |
| `payrollResultLines` | `FindPayrollResultLineByPayrollResultID` | `payroll_result_lines.read` | global dengan MFA terkini |
| `payrollRuns` | `GetAllPayrollRuns`, `FindPayrollRunByIdempotencyKey`, `FindPayrollRunByPayrollPeriodID` | `payroll_runs.read` | global |
| `payrollPeriod` | `GetPayrollPeriodByID` | `payroll_periods.read` | global |
| `payrollPeriods` | `GetAllPayrollPeriods`, `FindPayrollPeriodByPayrollGroupID` | `payroll_periods.read` | global |
| `payrollGroup` | `GetPayrollGroupByID` | `payroll_groups.read` | global |
| `payrollGroups` | `GetAllPayrollGroups` | `payroll_groups.read` | global |
| `payrollRuleSet` | `GetPayrollRuleSetByID` | `payroll_rule_sets.read` | global |
| `payrollRuleSets` | `GetAllPayrollRuleSets` | `payroll_rule_sets.read` | global |
| `myPayslip` | `GetMyPayslip` | `payslips.read_self` | self |
| `myPayslipPdf` | `GetMyPayslip` dengan `include_pdf` | `payslips.read_self` | self, respons tanpa cache |
| `myPayslips` | `ListMyPayslips` | `payslips.read_self` | self |
| `leavePolicy` | `GetLeavePolicyByID` | `leave_policies.read` | global |
| `leavePolicies` | `GetAllLeavePolicies` | `leave_policies.read` | global |
| `leaveEntitlement` | `GetLeaveEntitlementByID` | `leave_entitlements.read` | global |
| `leaveEntitlements` | `FindLeaveEntitlementByWorkerID` | `leave_entitlements.read` | global |
| `leaveLedgerEntry` | `GetLeaveLedgerEntryByID` | `leave_ledger_entries.read` | global |
| `leaveLedgerEntries` | `FindLeaveLedgerEntryByEntitlementID`, `FindLeaveLedgerEntryByAdjustmentID` | `leave_ledger_entries.read` | global |
| `leaveAdjustment` | `GetLeaveAdjustmentByID` | `leave_adjustments.read` | global |
| `leaveAdjustments` | `FindLeaveAdjustmentByEntitlementID`, `FindLeaveAdjustmentByIdempotencyKey` | `leave_adjustments.read` | global |

## Penjelasan Mutation

Seluruh Mutation HRM mengembalikan payload non-null yang mengimplementasikan `MutationResult`. Baca `was_successful` dan `errors { type message }` sebelum memakai field hasil. Penolakan bisnis dan akses pada resolver menggunakan error payload. Validasi GraphQL serta penolakan pada guard transport, seperti CSRF, tetap muncul pada error GraphQL tingkat atas.

| Mutation | RPC | Izin dengan prefiks `human_resource.` | Payload dan field hasil |
| --- | --- | --- | --- |
| `requestEmployeeAccount` | `CreateAccessRequest` | `access_requests.create` dan `workers.read` | `RequestEmployeeAccountPayload!` · `request` |
| `createWorker` | `CreateWorker` | `workers.create` | `CreateWorkerPayload!` · `worker` |
| `createEmploymentContract` | `CreateEmploymentContract` | `employment_contracts.create` | `CreateEmploymentContractPayload!` · `employment_contract` |
| `updateEmploymentContract` | `UpdateEmploymentContract` | `employment_contracts.update` | `UpdateEmploymentContractPayload!` · `employment_contract` |
| `deleteEmploymentContract` | `DeleteEmploymentContract` | `employment_contracts.delete` | `DeleteEmploymentContractPayload!` · `deleted` |
| `recordSignedEmploymentContract` | `RecordSignedEmploymentContract` | `employment_contracts.record_signature` | `RecordSignedEmploymentContractPayload!` · `employment_contract` |
| `updateWorker` | `UpdateWorker` | `workers.update` | `UpdateWorkerPayload!` · `worker` |
| `createPosition` | `CreatePosition` | `positions.create` | `CreatePositionPayload!` · `position` |
| `updatePosition` | `UpdatePosition` | `positions.update` | `UpdatePositionPayload!` · `position` |
| `deletePosition` | `DeletePosition` | `positions.delete` | `DeletePositionPayload!` · `deleted` |
| `activatePosition` | `ActivatePosition` | `positions.activate` | `ActivatePositionPayload!` · `position` |
| `activateWorker` | `ActivateWorker` | `workers.activate` | `ActivateWorkerPayload!` · `worker` |
| `terminateWorker` | `TerminateWorker` | `workers.terminate` | `TerminateWorkerPayload!` · `worker` |
| `createOffboardingPlan` | `CreateOffboardingPlan` | `offboarding_plans.create` | `CreateOffboardingPlanPayload!` · `offboarding_plan` |
| `createOffboardingTask` | `CreateOffboardingTask` | `offboarding_tasks.create` | `CreateOffboardingTaskPayload!` · `offboarding_task` |
| `completeOffboardingTask` | `CompleteOffboardingTask` | `offboarding_tasks.complete` | `CompleteOffboardingTaskPayload!` · `offboarding_task` |
| `waiveOffboardingTask` | `WaiveOffboardingTask` | `offboarding_tasks.waive` | `WaiveOffboardingTaskPayload!` · `offboarding_task` |
| `approveOffboardingPlan` | `ApproveOffboardingPlan` | `offboarding_plans.approve` | `ApproveOffboardingPlanPayload!` · `offboarding_plan` |
| `startOffboardingPlan` | `StartOffboardingPlan` | `offboarding_plans.start` | `StartOffboardingPlanPayload!` · `offboarding_plan` |
| `cancelOffboardingPlan` | `CancelOffboardingPlan` | `offboarding_plans.cancel` | `CancelOffboardingPlanPayload!` · `offboarding_plan` |
| `completeOffboardingPlan` | `CompleteOffboardingPlan` | `offboarding_plans.complete` | `CompleteOffboardingPlanPayload!` · `offboarding_plan` |
| `applyPrimaryAssignment` | `ApplyPrimaryAssignment` | `workers.apply_primary_assignment` | `ApplyPrimaryAssignmentPayload!` · `worker` |
| `createWorkerAssignment` | `CreateWorkerAssignment` | `worker_assignments.create` | `CreateWorkerAssignmentPayload!` · `worker_assignment` |
| `approveWorkerAssignment` | `ApproveWorkerAssignment` | `worker_assignments.approve` | `ApproveWorkerAssignmentPayload!` · `worker_assignment` |
| `submitWorkerAssignmentTransfer` | `SubmitWorkerAssignmentTransfer` | `worker_assignments.create` | `SubmitWorkerAssignmentTransferPayload!` · `worker_assignment_transfer` |
| `decideWorkerAssignmentTransfer` | `DecideWorkerAssignmentTransfer` | `worker_assignments.transfer` | `DecideWorkerAssignmentTransferPayload!` · `worker_assignment_transfer` |
| `freezePayrollInput` | `FreezePayrollInput` | `payroll_runs.create` | `FreezePayrollInputPayload!` · `frozen_payroll_input` |
| `createPayrollRun` | `CreatePayrollRun` | `payroll_runs.create` | `CreatePayrollRunPayload!` · `payroll_run` |
| `calculatePayrollRun` | `CalculatePayrollRun` | `payroll_runs.calculate` | `CalculatePayrollRunPayload!` · `payroll_run` |
| `submitPayrollRunForReview` | `SubmitPayrollRunForReview` | `payroll_runs.submit_review` | `SubmitPayrollRunForReviewPayload!` · `payroll_run` |
| `approvePayrollRun` | `ApprovePayrollRun` | `payroll_runs.approve` | `ApprovePayrollRunPayload!` · `payroll_run` |
| `cancelPayrollRun` | `CancelPayrollRun` | `payroll_runs.cancel` | `CancelPayrollRunPayload!` · `payroll_run` |
| `requestPayrollPosting` | `RequestPayrollPosting` | `payroll_runs.request_posting` | `RequestPayrollPostingPayload!` · `payroll_posting_request` |
| `createLeaveAdjustment` | `CreateLeaveAdjustment` | `leave_adjustments.create` | `CreateLeaveAdjustmentPayload!` · `leave_adjustment` |
| `approveLeaveAdjustment` | `ApproveLeaveAdjustment` | `leave_adjustments.approve` | `ApproveLeaveAdjustmentPayload!` · `leave_adjustment` |
| `rejectLeaveAdjustment` | `RejectLeaveAdjustment` | `leave_adjustments.reject` | `RejectLeaveAdjustmentPayload!` · `leave_adjustment` |
| `closeLeaveEntitlement` | `CloseLeaveEntitlement` | `leave_entitlements.close` | `CloseLeaveEntitlementPayload!` · `leave_entitlement` |
| `postLeaveAdjustment` | `PostLedgerEntry` untuk jenis adjustment | `leave_entitlements.post` | `PostLeaveAdjustmentPayload!` · `leave_entitlement` |
| `createWorkforcePlan` | `CreateWorkforcePlan` | `workforce_plans.create` | `CreateWorkforcePlanPayload!` · `workforce_plan` |
| `updateWorkforcePlan` | `UpdateWorkforcePlan` | `workforce_plans.update` | `UpdateWorkforcePlanPayload!` · `workforce_plan` |
| `createWorkforcePlanLine` | `CreateWorkforcePlanLine` | `workforce_plan_lines.create` | `CreateWorkforcePlanLinePayload!` · `workforce_plan_line` |
| `updateWorkforcePlanLine` | `UpdateWorkforcePlanLine` | `workforce_plan_lines.update` | `UpdateWorkforcePlanLinePayload!` · `workforce_plan_line` |
| `deleteWorkforcePlanLine` | `DeleteWorkforcePlanLine` | `workforce_plan_lines.delete` | `DeleteWorkforcePlanLinePayload!` · `deleted` |
| `submitWorkforcePlan` | `SubmitWorkforcePlan` | `workforce_plans.submit` | `SubmitWorkforcePlanPayload!` · `workforce_plan` |
| `cancelWorkforcePlan` | `CancelWorkforcePlan` | `workforce_plans.cancel` | `CancelWorkforcePlanPayload!` · `workforce_plan` |
| `deleteWorkforcePlan` | `DeleteWorkforcePlan` | `workforce_plans.delete` | `DeleteWorkforcePlanPayload!` · `deleted` |
| `approveWorkforcePlan` | `ApproveWorkforcePlan` | `workforce_plans.approve` | `ApproveWorkforcePlanPayload!` · `workforce_plan` |
| `supersedeWorkforcePlan` | `SupersedeWorkforcePlan` | `workforce_plans.supersede` | `SupersedeWorkforcePlanPayload!` · `workforce_plan` |

## Penjelasan createWorkforcePlan

Mutation ini membuat draf rencana tenaga kerja. HRM menetapkan pembuat dari sesi dan status awal `draft`. Referensi anggaran Finance digunakan pada tahap persetujuan. [Uji pembuatan draf](../../e2e/gateway/59-hrm-mutation-buat-draf-rencana.md).

## Penjelasan createWorkforcePlanLine

Mutation ini menambahkan kebutuhan FTE, referensi jabatan dan unit, periode, tindakan, serta estimasi biaya opsional pada draf. HRM memeriksa status induk dan validitas referensi. [Uji penambahan rincian](../../e2e/gateway/60-hrm-mutation-tambah-rincian-rencana.md).

## Penjelasan deleteWorkforcePlanLine

Mutation ini menghapus rincian pada draf dan mengembalikan `deleted: true` pada payload setelah RPC berhasil. HRM menolak penghapusan rincian yang dirujuk rekrutmen atau berada pada rencana di luar status draf. [Uji penghapusan rincian](../../e2e/gateway/61-hrm-mutation-hapus-rincian-rencana.md).

## Penjelasan updateWorkforcePlanLine

Mutation ini mengoreksi field yang dikirim. Field yang absen dipertahankan dan `null` mengosongkan field yang nullable. ID rencana induk tetap milik rincian. HRM memeriksa status draf dan referensi rekrutmen. [Uji koreksi rincian](../../e2e/gateway/62-hrm-mutation-koreksi-rincian-rencana.md).

## Penjelasan updateWorkforcePlan

Mutation ini mengoreksi tanggal awal, tanggal akhir, atau skenario pada draf. HRM memeriksa bahwa periode baru tetap mencakup rincian yang sudah ada. [Uji koreksi draf](../../e2e/gateway/63-hrm-mutation-koreksi-draf-rencana.md).

## Penjelasan approveWorkforcePlan dan supersedeWorkforcePlan

Persetujuan memerlukan approver berbeda dan bukti anggaran Finance untuk snapshot rencana. Penggantian menautkan rencana disetujui ke versi pengganti yang memenuhi aturan HRM. Kedua mutation memerlukan grant global masing-masing, MFA, dan CSRF. [Uji persetujuan dan penggantian](../../e2e/gateway/64-hrm-mutation-setujui-dan-ganti-rencana.md) · [Handoff anggaran Finance](../../notes/finance_accounting/2-human_resource-workforce-budget.md).

## Penjelasan deleteWorkforcePlan

Mutation ini menghapus draf kosong dengan grant delete global dan MFA. HRM mengunci rencana dan menolak penghapusan jika masih ada rincian atau status sudah berubah. [Uji penghapusan draf](../../e2e/gateway/65-hrm-mutation-hapus-draf-rencana.md).

## Penjelasan createWorker dan updateWorker

Pendaftaran pekerja menghasilkan status preboarding. Koreksi menerima perubahan tanggal rekrut dan awal masa kerja pada record preboarding sesuai pemeriksaan referensi HRM. Status, assignment primer, dan pengaitan principal mengikuti workflow terpisah. [Uji preboarding](../../e2e/gateway/66-hrm-mutation-preboarding-pekerja.md).

## Penjelasan activateWorker dan terminateWorker

Aktivasi memerlukan status preboarding, tanggal yang sudah efektif, dan kontrak ditandatangani yang masih berlaku. HRM mengaktifkan pekerja sebelum assignment primer dan principal IAM ditautkan. Terminasi menggunakan rencana offboarding yang disetujui untuk pekerja terkait. HRM menutup kontrak dan assignment terkait dalam transaksi terminasi. [Uji aktivasi dan terminasi](../../e2e/gateway/67-hrm-mutation-aktivasi-dan-terminasi.md).

## Penjelasan applyPrimaryAssignment

Penetapan assignment primer memerlukan pekerja aktif dan assignment yang lolos aturan HRM. Penautan akun berlangsung melalui jalur provisioning IAM dan RPC privat HRM. [Alur pengajuan akun](33-alur-pengajuan-akun-employee.md).

## Penjelasan activatePosition

Mutation ini mengaktifkan posisi berstatus planned. HRM memeriksa prasyarat dan mengembalikan posisi berstatus active. Operasi ini memakai grant global, MFA, dan CSRF. [Uji aktivasi posisi](../../e2e/gateway/69-hrm-mutation-aktivasi-posisi.md).

## Penjelasan approveWorkerAssignment

Persetujuan assignment menggunakan grant global, sesi employee, MFA, dan CSRF. HRM memeriksa pemisahan pengusul dan approver, kelayakan assignment, serta kapasitas posisi. Persetujuan menghasilkan waktu persetujuan dan FTE. Penetapan assignment primer dilakukan melalui operasi tersendiri. [Uji persetujuan assignment](../../e2e/gateway/70-hrm-mutation-persetujuan-assignment.md).

## Penjelasan workerAssignment dan workerAssignments

Operator melihat detail atau halaman per pekerja atau posisi dengan grant baca assignment. Daftar menerima tepat satu filter. Worker dapat memakai UUID atau Relay ID. Gateway memeriksa hasil RPC terhadap ID filter. Waktu pengajuan, persetujuan, dan pembatalan dapat kosong sesuai kondisi record. [Uji Query assignment](../../e2e/gateway/71-hrm-query-assignment-pekerja.md).

## Penjelasan createWorkerAssignment

Mutation ini mencatat pengajuan assignment. HRM menetapkan pengusul dan waktu pengajuan dari sesi. Persetujuan dan penetapan assignment primer memakai mutation tersendiri. [Uji pengajuan assignment](../../e2e/gateway/72-hrm-mutation-pengajuan-assignment.md).

## Penjelasan transfer assignment pekerja

Pengajuan membawa kunci idempotensi dan referensi assignment asal serta posisi tujuan. Pengulangan dengan kunci dan isi yang sama dapat mengembalikan status terakhir, termasuk applied atau rejected. Otorisasi keputusan memakai grant transfer dan MFA. HRM menolak keputusan oleh pengusul atau pekerja terkait dan menerapkan persetujuan secara atomik. [Uji transfer](../../e2e/gateway/73-hrm-mutation-transfer-assignment.md).

## Penjelasan createPosition

Mutation ini mencatat posisi planned dengan grant global, MFA, dan CSRF. HRM memeriksa jabatan dan referensi organisasi. Aktivasi memakai mutation tersendiri. [Uji pembuatan posisi](../../e2e/gateway/74-hrm-mutation-buat-posisi.md).

## Penjelasan updatePosition

Operator mengoreksi judul, kapasitas, lokasi kerja, atau periode posisi planned. HRM memeriksa referensi sebelum menerima perubahan. Lokasi dan tanggal akhir dapat dikosongkan dengan null eksplisit. Cost center memerlukan verifier pemilik sebelum dapat dipakai dalam input Gateway. [Uji koreksi posisi](../../e2e/gateway/75-hrm-mutation-koreksi-posisi.md).

## Penjelasan deletePosition

Mutation ini menghapus posisi planned dengan grant global, MFA, dan CSRF. HRM menolak penghapusan bila posisi dirujuk struktur pelaporan, assignment, requisition, atau succession plan. [Uji penghapusan posisi](../../e2e/gateway/76-hrm-mutation-hapus-posisi.md).

## Penjelasan detail dan daftar kontrak kerja

Kedua Query menampilkan metadata kontrak per pekerja, termasuk status dan periode efektif. Gateway memeriksa seluruh hasil terhadap ID pekerja. Akses dokumen tanda tangan mengikuti DRM. [Uji metadata kontrak](../../e2e/gateway/79-hrm-query-kontrak-pekerja.md).

## Penjelasan createEmploymentContract

Mutation ini membuat kontrak kerja draf dengan grant global, MFA, dan CSRF. HRM memeriksa pekerja, badan hukum, dan referensi perjanjian kolektif bila ada. Pencatatan tanda tangan mengikuti workflow dan bukti DRM tersendiri. [Uji pembuatan kontrak](../../e2e/gateway/80-hrm-mutation-buat-draf-kontrak.md).

## Penjelasan recordSignedEmploymentContract

Mutation ini mengirim ID kontrak dan versi dokumen. HRM memverifikasi bukti DRM sebelum mengubah status draf menjadi signed. Gateway memeriksa ID kontrak dan versi pada respons. Grant global, MFA, dan CSRF wajib. [Uji pencatatan tanda tangan](../../e2e/gateway/81-hrm-mutation-catat-tanda-tangan-kontrak.md).

## Penjelasan deleteEmploymentContract

Mutation ini meminta HRM menghapus kontrak draf yang belum ditandatangani. Grant global delete, MFA, dan CSRF dibutuhkan. Penolakan prasyarat dipetakan menjadi CONFLICT. [Uji penghapusan kontrak](../../e2e/gateway/82-hrm-mutation-hapus-draf-kontrak.md).

## Penjelasan updateEmploymentContract

Mutation ini memperbarui field yang dikirim pada kontrak draf. Field opsional bisa dikosongkan dengan null. Pekerja, badan hukum, status dan bukti tanda tangan ditangani oleh aturan HRM. Grant global update, MFA dan CSRF dibutuhkan. [Uji koreksi kontrak](../../e2e/gateway/83-hrm-mutation-koreksi-draf-kontrak.md).

## Penjelasan offboardingPlan dan offboardingPlans

Operator membaca status dan tanggal rencana offboarding berdasarkan ID rencana atau pekerja. Query memakai grant baca offboarding tersendiri. Daftar memeriksa kecocokan pekerja pada hasil RPC. [Uji Query offboarding](../../e2e/gateway/85-hrm-query-rencana-offboarding.md).

## Penjelasan persetujuan dan mulai offboarding

HRM memeriksa prasyarat persetujuan independen. Setelah disetujui, rencana bisa dimulai dalam periode offboarding dan berstatus in_progress. Kedua perintah membutuhkan grant global masing-masing, MFA dan CSRF. [Uji workflow offboarding](../../e2e/gateway/86-hrm-mutation-setujui-dan-mulai-offboarding.md).

## Penjelasan createOffboardingPlan

Mutation ini mencatat draf untuk pekerja dengan nomor rencana, alasan, dan tanggal workflow. HRM menetapkan status dan pembuat. Operator memakai grant global create dan CSRF. [Uji pembuatan draf offboarding](../../e2e/gateway/87-hrm-mutation-buat-draf-offboarding.md).

## Penjelasan createOffboardingTask

Operator menambah tugas pada rencana draf. HRM menetapkan status pending, memeriksa periode, dan memegang aturan transisi tugas. Grant global create dan CSRF dibutuhkan. [Uji pembuatan tugas](../../e2e/gateway/88-hrm-mutation-buat-tugas-offboarding.md).

## Penjelasan offboardingTasks

Query menampilkan tugas per rencana dengan paginasi, status, tenggat, dan penanda wajib. Gateway memeriksa kecocokan ID rencana pada hasil RPC. [Uji daftar tugas](../../e2e/gateway/89-hrm-query-tugas-offboarding.md).

## Penjelasan penyelesaian dan pengecualian tugas offboarding

HRM memeriksa tugas pending milik HRM pada rencana berjalan. Penyelesaian memakai bukti konfirmasi terverifikasi. Pengecualian menyimpan alasan. Kedua mutation memakai grant global terpisah, MFA, dan CSRF. [Uji penyelesaian tugas](../../e2e/gateway/90-hrm-mutation-selesaikan-dan-waive-tugas.md).

## Penjelasan completeOffboardingPlan

Gateway meneruskan ID payroll final dan referensi clearance. HRM memeriksa terminasi, tugas wajib, hasil payroll pekerja, dan bukti Finance serta IAM sebelum menutup rencana. Grant global complete, MFA, dan CSRF dibutuhkan. [Uji penutupan offboarding](../../e2e/gateway/91-hrm-mutation-tutup-offboarding.md).

## Penjelasan cancelOffboardingPlan

HRM menerima pembatalan rencana yang disetujui dan belum berjalan. HRM mencatat alasan serta waktu pembatalan. Grant global cancel, MFA, dan CSRF dibutuhkan. [Uji pembatalan](../../e2e/gateway/92-hrm-mutation-batalkan-offboarding.md).

## Penjelasan offboardingTask

Query detail tugas memanggil `GetOffboardingTaskByID` dengan grant global employee `human_resource.offboarding_tasks.read`. Gateway memvalidasi ID tugas dan ID rencana pada respons. [Uji detail tugas](../../e2e/gateway/93-hrm-query-detail-tugas-offboarding.md).

## RBAC

Seluruh jalur memerlukan sesi employee. Mutasi memakai scope global serta CSRF. Mutasi payroll, keputusan dan pembukuan koreksi, penutupan hak cuti, serta pengajuan, penarikan, persetujuan, penggantian rencana tenaga kerja, dan penghapusan rinciannya juga memerlukan autentikasi terkini dan MFA. Aktivasi, terminasi, penetapan assignment primer, pengaitan principal pekerja, dan penghapusan draf rencana juga memerlukan MFA. Keputusan cakupan terhadap objek dan pemisahan tugas tetap diperiksa di HRM. Provisioning IAM memerlukan manifest dan persetujuan pemilik katalog. [Handoff IAM](../../notes/iam_management/1-human_resource-permission-handoff.md).

## Verifikasi dan dependency
- Seluruh 43 Query dan 47 Mutation pada rollout terbaru dikenali lewat Caddy HTTPS dan mencapai guard sesi atau CSRF. [Matriks guard seluruh schema](../../e2e/gateway/84-hrm-query-mutation-guard-seluruh-schema.md).
- Tes HTTPS otomatis mencakup guard sesi untuk detail dan daftar assignment serta guard CSRF untuk enam mutation assignment dan posisi. [Hasil pengujian](../../e2e/gateway/78-hrm-query-mutation-guard-assignment-posisi.md).
- Kedua replika lokal sudah memakai image terbaru. Caddy HTTPS mengenali Query assignment dan Mutation hapus posisi, serta menerapkan guard sesi dan CSRF. [Bukti rollout](../../e2e/gateway/77-hrm-query-verifikasi-schema-runtime.md#rollout-lokal-dan-verifikasi-https).
- Image terbaru lulus pemeriksaan startup dan pemuatan schema pada container uji loopback. Query assignment menolak akses anonim dan Mutation hapus posisi menolak permintaan tanpa CSRF. [Hasil uji image](../../e2e/gateway/77-hrm-query-verifikasi-schema-runtime.md#image-terbaru-pada-container-uji).
- Pemeriksaan sebelum rollout menemukan schema runtime belum memuat Query assignment dan Mutation hapus posisi. Build Go dari checkout lulus. [Verifikasi schema runtime](../../e2e/gateway/77-hrm-query-verifikasi-schema-runtime.md).

- Uji fixture membuktikan pemetaan GraphQL ke RPC dan penolakan grant yang tidak sesuai. Pada lingkungan lokal, Gateway menolak akses anonim ke Query dan permintaan Mutation tanpa CSRF. Uji positif HTTPS dengan grant IAM resmi masih terbuka. [Query HTTPS](../../e2e/gateway/18-hrm-query-akses-anonim-https.md) · [Status payroll HTTPS](../../e2e/gateway/21-hrm-mutation-request-payroll-posting-https.md) · [Mutation HTTPS](../../e2e/gateway/26-hrm-mutation-workflow-payroll-https.md).
- Kalkulasi real memerlukan adapter kalkulator yang terpasang di HRM. Metadata slip real membutuhkan bukti publikasi Finance dan versi PDF di DRM. Unduh dan cetak mandiri membutuhkan akses privat DRM. [Finance](25-kebutuhan-integrasi-finance-payroll.md) · [DRM](24-kebutuhan-integrasi-drm.md).
- Inventaris per service tetap menjadi sumber kandidat GraphQL dan aturan workflow untuk operasi lain. [Daftar service](1-inventaris-operasi-grpc-hrm.md).

## Pemeriksaan IAM untuk izin slip mandiri

Manifest resmi HRM memuat `human_resource.payslips.read_self`. IAM menyinkronkan definisi izin saat startup. Penugasan tetap melalui GraphQL dengan scope `SELF` yang ID-nya cocok dengan principal sesi. [Uji scope HTTPS](../../e2e/gateway/138-hrm-query-slip-dengan-grant-self-dan-global.md).

Pengenal self-service bersama kini mengenali kode `human_resource.payslips.read_self` secara eksplisit. Evaluator IAM, snapshot bersama, Gateway, dan policy gRPC HRM memerlukan grant `SELF` eksplisit untuk principal sesi pada jalur slip gaji. Grant global untuk izin ini tidak memenuhi target `SELF`, termasuk jika grant global muncul lebih dulu daripada grant yang sah. E2E akses PDF tetap memerlukan bukti Finance dan DRM. [Pengenal self-service](../../../backend/shared/authz/snapshot.go) · [Evaluator IAM](../../../backend/services/iam/domain/service/access_policy.go) · [Policy HRM](../../../backend/services/human_resource/infrastructure/security/policy.go).

Daftar umum HRM membaca item dan total melalui pernyataan terpisah. Saat data berubah di antaranya, jumlah item dapat berbeda dari proyeksi total terbaru. Gateway meneruskan total dari HRM dan memvalidasi batas halaman. Daftar slip terbit mengikuti kontrak konsistensi total milik verifier slip. [Regresi paginasi pekerja](../../e2e/gateway/42-hrm-query-pekerja-berdasarkan-nomor.md).

Pemetaan 67 izin untuk tinjauan provisioning, termasuk 39 izin dengan MFA minimum runtime, tercatat pada [handoff izin IAM](28-izin-iam-untuk-wiring-gateway.md).

## Provisioning akun employee melalui IAM

Petugas HR mengirim `requestEmployeeAccount` dengan `worker_id`. HRM memvalidasi lifecycle perekrutan, lalu IAM memeriksa badan hukum melalui Enterprise Foundation, menerapkan kebijakan akun dasar `iam_selfservice`, dan menyimpan penerimaan pengajuan.

Operator IAM memakai MFA dan izin global `iam.users.write` untuk membuka antrean. Mutation `createEmployeeUser` mewajibkan `worker_id` dan `provisioning_request_id`. Pekerja pada input harus cocok dengan pengajuan dan berada dalam scope baca operator. IAM menyimpan referensi akun saat pengajuan diproses. Penautan memakai RPC HRM privat dan pengajuan ditandai `linked` setelah tautan dikonfirmasi.

Query HRM membaca status dari IAM. Workspace Pekerja menampilkan Menunggu proses IAM atau Akun tertaut berdasarkan hasil. [Kontrak operasi dan pengujian](33-alur-pengajuan-akun-employee.md).

## Pembuatan kandidat

Mutation `createCandidate` memakai RPC `CreateCandidate` dan izin `human_resource.candidates.create` dengan scope GLOBAL. [Uji HTTPS pembuatan kandidat](../../e2e/gateway/144-hrm-mutation-pembuatan-kandidat-ef.md).

Query `candidate` memanggil `GetCandidateByID`. Query `candidates` memanggil `GetAllCandidates` atau `FindCandidateByCandidateNumber`. Keduanya memakai `human_resource.candidates.read` pada scope GLOBAL. [Uji HTTPS daftar dan detail kandidat](../../e2e/gateway/145-hrm-query-daftar-dan-detail-kandidat.md).

Mutation `updateCandidate` memakai RPC `UpdateCandidate` dan izin `human_resource.candidates.update` dengan scope GLOBAL. [Uji HTTPS koreksi nama kandidat](../../e2e/gateway/146-hrm-mutation-koreksi-nama-kandidat.md).

Mutation `deleteCandidate` memakai RPC `DeleteCandidate` dan izin `human_resource.candidates.delete` dengan scope GLOBAL serta MFA terbaru. [Uji HTTPS penghapusan kandidat](../../e2e/gateway/147-hrm-mutation-penghapusan-kandidat.md).

Query `myPayslipPdf(payroll_result_id: ID!)` meminta PDF dari RPC `GetMyPayslip` dengan flag `include_pdf`. Query memakai izin `human_resource.payslips.read_self` dengan scope SELF dan mengembalikan `content_base64` dalam respons yang tidak disimpan cache.

Query `requisition` dan `requisitions` menggunakan `GetRequisitionByID`, `GetAllRequisitions`, atau `FindRequisitionByRequisitionNumber`. Akses membutuhkan `human_resource.requisitions.read` dengan scope GLOBAL. [Uji HTTPS requisition](../../e2e/gateway/148-hrm-query-daftar-detail-requisition.md).

Mutation `createRequisition` menggunakan RPC `CreateRequisition` dan permission `human_resource.requisitions.create`. [Bukti HTTPS](../../e2e/gateway/149-hrm-mutation-buat-draf-requisition.md).

Mutation `submitRequisition` menggunakan RPC `SubmitRequisition`, permission `human_resource.requisitions.submit`, dan MFA. [Bukti HTTPS](../../e2e/gateway/150-hrm-mutation-submit-requisition-mfa.md).

Mutation `rejectRequisition` menggunakan RPC `RejectRequisition`, permission `human_resource.requisitions.reject`, dan MFA. [Bukti HTTPS](../../e2e/gateway/151-hrm-mutation-penolakan-requisition-dua-aktor.md).

Mutation `deleteRequisition` memakai RPC `DeleteRequisition`, grant sensitif dan MFA. [Bukti HTTPS](../../e2e/gateway/152-hrm-mutation-hapus-draf-requisition.md).

Mutation `updateRequisition` menggunakan RPC `UpdateRequisition`. [Bukti HTTPS](../../e2e/gateway/153-hrm-query-mutation-nama-sdl-tanpa-prefix.md).

Query `workerAssignmentTransfer` memakai RPC `GetWorkerAssignmentTransfer` dan izin `human_resource.worker_assignment_transfers.read`. [Uji HTTPS](../../e2e/gateway/156-hrm-query-detail-transfer-penugasan.md).

Query `workerAssignmentTransfers` memakai RPC `ListWorkerAssignmentTransfers` dan izin baca transfer. [Uji HTTPS](../../e2e/gateway/157-hrm-query-riwayat-transfer-penugasan.md).

Query `jobsByIds` menggunakan RPC `GetJobsByIds` dan izin baca jabatan. [Bukti HTTPS](../../e2e/gateway/166-hrm-query-label-jabatan-batch.md).

Mutation `approveRequisition` memakai grant sensitif approval pada unit tepat. `openRequisition` memakai grant sensitif pembukaan GLOBAL. HRM memeriksa mandat dan bukti anggaran sebelum perubahan status. [Uji HTTPS](../../e2e/gateway/173-hrm-mutation-guard-approve-open-requisition.md).

Mutation `closeRequisition` memakai izin sensitif `requisitions.close` GLOBAL. HRM menolak penutupan bila requisition belum open atau masih memiliki posting atau lamaran aktif. [Uji HTTPS](../../e2e/gateway/176-hrm-mutation-close-requisition-https.md).

## Mutation katalog jabatan

`createJobFamily` dan `createJob` memakai RPC `CreateJobFamily` dan `CreateJob`. Keduanya memerlukan grant employee global sesuai operasi. Jabatan yang dibuat dapat dipilih ketika membuat requisition. [Bukti HTTPS](../../e2e/gateway/177-hrm-mutation-katalog-jabatan-untuk-requisition.md).

`reissueEmployeePayslip` memakai RPC `ReissueEmployeePayslip`, grant `human_resource.payslips.reissue` dengan scope badan hukum, serta MFA. [Handoff cetak ulang](39-cetak-ulang-slip-gaji-mantan-karyawan.md).
