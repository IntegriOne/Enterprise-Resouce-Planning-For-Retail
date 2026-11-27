# List operasi

## Butuh GraphQL Mutation

Schema HRM memiliki 61 Mutation. Pengujian memakai grant uji yang diberikan melalui GraphQL IAM. Matriks ini menyebutkan bukti yang lebih spesifik daripada [uji guard 47 Mutation sebelumnya](../../e2e/gateway/96-hrm-mutation-seluruh-schema-superadmin-https.md).

Sebanyak 53 Mutation memiliki bukti jalur berhasil melalui HTTPS. Delapan Mutation masih memerlukan verifikasi runtime lengkap. Sejumlah uji jalur berhasil memakai data prasyarat sintetis melalui SQL pada database disposable. Persetujuan, assignment role, dan pemasangan permission tetap lewat GraphQL.

## Penjelasan operasi dengan bukti berhasil

| Mutation | Laporan HTTPS |
| --- | --- |
| `activateWorker` | [135](../../e2e/gateway/135-hrm-mutation-aktivasi-dan-provisioning-akun.md), dengan fixture kontrak bertanda tangan |
| `requestEmployeeAccount` | [135](../../e2e/gateway/135-hrm-mutation-aktivasi-dan-provisioning-akun.md), termasuk pengajuan ulang |
| `activatePosition` | [112](../../e2e/gateway/112-hrm-query-mutation-lifecycle-posisi-mfa.md) |
| `applyPrimaryAssignment` | [113](../../e2e/gateway/113-hrm-query-mutation-assignment-primer-mfa.md) |
| `approveLeaveAdjustment` | [110](../../e2e/gateway/110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md) |
| `approveOffboardingPlan` | [104](../../e2e/gateway/104-hrm-mutation-persetujuan-dan-pembatalan-offboarding-mfa.md) |
| `approvePayrollRun` | [108](../../e2e/gateway/108-hrm-query-mutation-review-approval-dan-pengajuan-payroll.md) |
| `approveWorkerAssignment` | [113](../../e2e/gateway/113-hrm-query-mutation-assignment-primer-mfa.md) |
| `cancelOffboardingPlan` | [104](../../e2e/gateway/104-hrm-mutation-persetujuan-dan-pembatalan-offboarding-mfa.md) |
| `cancelPayrollRun` | [109](../../e2e/gateway/109-hrm-mutation-pembatalan-payroll-saat-review.md) |
| `cancelWorkforcePlan` | [115](../../e2e/gateway/115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md) |
| `closeRequisition` | [176](../../e2e/gateway/176-hrm-mutation-close-requisition-https.md), dengan fixture anggaran pada tahap pembukaan |
| `closeLeaveEntitlement` | [111](../../e2e/gateway/111-hrm-mutation-penutupan-entitlement-cuti-mfa.md) |
| `createCandidate` | [144](../../e2e/gateway/144-hrm-mutation-pembuatan-kandidat-ef.md) |
| `deleteCandidate` | [147](../../e2e/gateway/147-hrm-mutation-penghapusan-kandidat.md) |
| `updateCandidate` | [146](../../e2e/gateway/146-hrm-mutation-koreksi-nama-kandidat.md) |
| `createEmploymentContract` | [101](../../e2e/gateway/101-hrm-query-mutation-draf-kontrak-mfa.md) |
| `createLeaveAdjustment` | [110](../../e2e/gateway/110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md) |
| `createOffboardingPlan` | [103](../../e2e/gateway/103-hrm-query-mutation-draf-offboarding-mfa.md) |
| `createOffboardingTask` | [103](../../e2e/gateway/103-hrm-query-mutation-draf-offboarding-mfa.md) |
| `createPayrollRun` | [107](../../e2e/gateway/107-hrm-query-mutation-draf-payroll-mfa.md) |
| `createPosition` | [112](../../e2e/gateway/112-hrm-query-mutation-lifecycle-posisi-mfa.md) |
| `createWorker` | [98](../../e2e/gateway/98-hrm-query-mutation-pekerja-dan-guard-offboarding.md) |
| `createWorkerAssignment` | [113](../../e2e/gateway/113-hrm-query-mutation-assignment-primer-mfa.md) |
| `createWorkforcePlan` | [97](../../e2e/gateway/97-hrm-mutation-referensi-badan-hukum-workforce.md) |
| `createWorkforcePlanLine` | [115](../../e2e/gateway/115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md) |
| `decideWorkerAssignmentTransfer` | [114](../../e2e/gateway/114-hrm-mutation-transfer-assignment-mfa.md) |
| `deleteEmploymentContract` | [101](../../e2e/gateway/101-hrm-query-mutation-draf-kontrak-mfa.md) |
| `createRequisition` | [149](../../e2e/gateway/149-hrm-mutation-buat-draf-requisition.md) |
| `deletePosition` | [112](../../e2e/gateway/112-hrm-query-mutation-lifecycle-posisi-mfa.md) |
| `deleteWorkforcePlan` | [100](../../e2e/gateway/100-hrm-query-mutation-mfa-payroll-workforce.md) · [116](../../e2e/gateway/116-hrm-mutation-persetujuan-dan-penggantian-workforce-mfa.md) |
| `deleteWorkforcePlanLine` | [115](../../e2e/gateway/115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md) |
| `postLeaveAdjustment` | [110](../../e2e/gateway/110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md) |
| `rejectLeaveAdjustment` | [110](../../e2e/gateway/110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md) |
| `requestPayrollPosting` | [108](../../e2e/gateway/108-hrm-query-mutation-review-approval-dan-pengajuan-payroll.md) |
| `startOffboardingPlan` | [105](../../e2e/gateway/105-hrm-mutation-mulai-offboarding-dan-pengecualian-tugas.md) |
| `submitPayrollRunForReview` | [108](../../e2e/gateway/108-hrm-query-mutation-review-approval-dan-pengajuan-payroll.md) |
| `submitWorkerAssignmentTransfer` | [114](../../e2e/gateway/114-hrm-mutation-transfer-assignment-mfa.md) |
| `submitWorkforcePlan` | [115](../../e2e/gateway/115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md) |
| `supersedeWorkforcePlan` | [116](../../e2e/gateway/116-hrm-mutation-persetujuan-dan-penggantian-workforce-mfa.md) |
| `terminateWorker` | [106](../../e2e/gateway/106-hrm-mutation-terminasi-pekerja-mfa.md) · [182](../../e2e/gateway/182-hrm-mutation-terminasi-pencabutan-sesi-iam.md) (deaktivasi IAM dan pencabutan sesi) |
| `updateEmploymentContract` | [101](../../e2e/gateway/101-hrm-query-mutation-draf-kontrak-mfa.md) |
| `deleteRequisition` | [152](../../e2e/gateway/152-hrm-mutation-hapus-draf-requisition.md) |
| `rejectRequisition` | [151](../../e2e/gateway/151-hrm-mutation-penolakan-requisition-dua-aktor.md) |
| `submitRequisition` | [150](../../e2e/gateway/150-hrm-mutation-submit-requisition-mfa.md) |
| `updateRequisition` | [153](../../e2e/gateway/153-hrm-query-mutation-nama-sdl-tanpa-prefix.md) |
| `updatePosition` | [112](../../e2e/gateway/112-hrm-query-mutation-lifecycle-posisi-mfa.md) |
| `updateWorker` | [98](../../e2e/gateway/98-hrm-query-mutation-pekerja-dan-guard-offboarding.md) |
| `updateWorkforcePlan` | [97](../../e2e/gateway/97-hrm-mutation-referensi-badan-hukum-workforce.md) |
| `updateWorkforcePlanLine` | [115](../../e2e/gateway/115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md) |
| `waiveOffboardingTask` | [105](../../e2e/gateway/105-hrm-mutation-mulai-offboarding-dan-pengecualian-tugas.md) |
| `createJobFamily`, `createJob` | [177](../../e2e/gateway/177-hrm-mutation-katalog-jabatan-untuk-requisition.md) |

## Penjelasan operasi yang memerlukan dependency runtime

| Mutation | Prasyarat untuk uji berhasil | Bukti penolakan |
| --- | --- | --- |
| `approveRequisition` | Mandat IAM untuk unit tepat dan [verifier anggaran Finance](37-kebutuhan-finance-anggaran-requisition.md) | [173](../../e2e/gateway/173-hrm-mutation-guard-approve-open-requisition.md) |
| `openRequisition` | Mandat IAM peninjau tersimpan dan [verifier anggaran Finance](37-kebutuhan-finance-anggaran-requisition.md) | [173](../../e2e/gateway/173-hrm-mutation-guard-approve-open-requisition.md) |
| `approveWorkforcePlan` | [Verifier keputusan anggaran Finance](../../notes/finance_accounting/2-human_resource-workforce-budget.md) | [116](../../e2e/gateway/116-hrm-mutation-persetujuan-dan-penggantian-workforce-mfa.md) |
| `calculatePayrollRun` | [Kalkulator HRM dan akses artefak aturan](27-kebutuhan-kalkulator-payroll-runtime.md) | [107](../../e2e/gateway/107-hrm-query-mutation-draf-payroll-mfa.md) |
| `completeOffboardingPlan` | [Verifier clearance Finance](38-kebutuhan-finance-clearance-offboarding.md), [IAM](13-operasi-offboarding.md#handoff-verifier-iam) dan payroll final | [118](../../e2e/gateway/118-hrm-mutation-guard-penyelesaian-offboarding-mfa.md) |
| `completeOffboardingTask` | [Verifier konfirmasi tugas dari pemilik tugas](13-operasi-offboarding.md#prasyarat-penyelesaian) | [118](../../e2e/gateway/118-hrm-mutation-guard-penyelesaian-offboarding-mfa.md) |
| `recordSignedEmploymentContract` | [Verifier versi kontrak bertanda tangan dari DRM](24-kebutuhan-integrasi-drm.md) | [102](../../e2e/gateway/102-hrm-mutation-guard-tanda-tangan-dan-aktivasi.md) |
| `reissueEmployeePayslip` | Adapter Finance dan DRM untuk membuka atau membuat PDF | [181](../../e2e/gateway/181-hrm-mutation-cetak-ulang-slip-gaji-https.md) · [Handoff](39-cetak-ulang-slip-gaji-mantan-karyawan.md). |


## RBAC dan batas bukti

Kebutuhan role, scope, dan MFA per operasi mengikuti [matriks wiring](26-status-wiring-gateway-hrm.md) dan [handoff IAM](28-izin-iam-untuk-wiring-gateway.md). Pengajuan jurnal berhasil hingga outbox HRM. Penerbitan dan pembayaran tetap dikerjakan Finance. Akses PDF setelah penerbitan bergantung pada integrasi DRM.


Browser HTTPS juga menguji pembuatan dan edit draf dengan referensi EF dan HRM, pengajuan serta penghapusan draf, dan penolakan oleh peninjau terpisah. [Referensi draf](../../frontend/e2e/human_resource/11-mutation-referensi-draf-requisition.md) · [Ajukan dan hapus](../../frontend/e2e/human_resource/12-mutation-ajukan-hapus-draf-requisition.md) · [Penolakan](../../frontend/e2e/human_resource/13-mutation-tolak-requisition.md).

`approveRequisition` dan `openRequisition` juga memiliki bukti transisi positif pada HTTPS dengan verifier anggaran fixture. Kedua mutation tetap berada dalam kelompok yang membutuhkan dependency runtime untuk verifikasi Finance lintas layanan. [Hasil uji fixture](../../e2e/gateway/175-hrm-mutation-approve-open-requisition-fixture-anggaran.md).
