# List operasi

## Butuh GraphQL Query

Gateway memiliki 52 Query sumber daya manusia. Sebanyak 49 Query memiliki bukti berhasil mengembalikan data melalui HTTPS, termasuk data sintetis pada database disposable. Tiga Query slip gaji masih menunggu pembuktian jalur positif lintas Finance dan DRM. Pengujian `node` merupakan bukti tambahan untuk root bersama di luar 52 Query domain.

## Penjelasan operasi dan bukti HTTPS

| Query | Bukti |
| --- | --- |
| `employeeAccountRequest` | [135](../../e2e/gateway/135-hrm-mutation-aktivasi-dan-provisioning-akun.md), status akhir dari IAM |
| `workerEmployeeAccountRequest` | [Uji browser](../../frontend/e2e/human_resource/1-mutation-pengajuan-akun-pekerja.md), status sebelum dan sesudah penautan |
| `workerAssignmentTransfers` | [157](../../e2e/gateway/157-hrm-query-riwayat-transfer-penugasan.md) |
| `workerAssignmentTransfer` | [156](../../e2e/gateway/156-hrm-query-detail-transfer-penugasan.md) |
| `worker`, `workers` | [34](../../e2e/gateway/34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md) (data terisi) |
| `workforcePlan`, `workforcePlans` | [99](../../e2e/gateway/99-hrm-query-workforce-plan-scope-badan-hukum.md) (data terisi) |
| `employmentContract`, `employmentContracts` | [101](../../e2e/gateway/101-hrm-query-mutation-draf-kontrak-mfa.md) (data terisi) |
| `offboardingPlan`, `offboardingPlans`, `offboardingTask`, `offboardingTasks` | [103](../../e2e/gateway/103-hrm-query-mutation-draf-offboarding-mfa.md) (data terisi) |
| `positions` | [112](../../e2e/gateway/112-hrm-query-mutation-lifecycle-posisi-mfa.md) (data terisi) |
| `workerAssignment`, `workerAssignments` | [113](../../e2e/gateway/113-hrm-query-mutation-assignment-primer-mfa.md) (data terisi) |
| `workforcePlanLine`, `workforcePlanLines` | [115](../../e2e/gateway/115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md) (data terisi) |
| `job`, `jobs`, `jobFamily`, `jobFamilies`, `position` | [120](../../e2e/gateway/120-hrm-query-katalog-jabatan-dan-posisi.md) (data terisi) |
| `payrollRun`, `payrollRuns`, `payrollPeriod`, `payrollPeriods`, `payrollGroup`, `payrollGroups`, `payrollRuleSet`, `payrollRuleSets` | [121](../../e2e/gateway/121-hrm-query-konfigurasi-payroll-dengan-data.md) (data terisi) |
| `leavePolicy`, `leavePolicies`, `leaveEntitlement`, `leaveEntitlements`, `leaveLedgerEntry`, `leaveLedgerEntries`, `leaveAdjustment`, `leaveAdjustments` | [122](../../e2e/gateway/122-hrm-query-kebijakan-entitlement-dan-ledger-cuti.md) (data terisi) |
| `payrollResult`, `payrollResults`, `payrollResultLine`, `payrollResultLines` | [123](../../e2e/gateway/123-hrm-query-detail-hasil-dan-komponen-payroll-mfa.md) (data terisi) |
| `candidate`, `candidates` | [145](../../e2e/gateway/145-hrm-query-daftar-dan-detail-kandidat.md) (kontak EF, filter persis, paginasi dan pencabutan role) |
| `requisition`, `requisitions` | [153](../../e2e/gateway/153-hrm-query-mutation-nama-sdl-tanpa-prefix.md), [154](../../e2e/gateway/154-hrm-query-koneksi-relay-requisition.md), [browser](../../frontend/e2e/human_resource/8-query-daftar-detail-requisition.md) |
| `jobsByIds` | [166](../../e2e/gateway/166-hrm-query-label-jabatan-batch.md) (label batch) |

## Penjelasan query slip gaji yang menunggu runtime

| Query | Bukti saat ini |
| --- | --- |
| `myPayslip`, `myPayslips` | [117](../../e2e/gateway/117-hrm-query-slip-gaji-scope-self-https.md) (guard), [138](../../e2e/gateway/138-hrm-query-slip-dengan-grant-self-dan-global.md) (grant global dan SELF bersamaan) |
| `myPayslipPdf` | Gateway dengan fixture gRPC memeriksa konten dan respons `no-store`. Stack HTTPS terisolasi menolak superadmin dengan GLOBAL saja dan mengembalikan `NOT_FOUND` untuk hasil yang tidak ada pada sesi pegawai tertaut dengan grant SELF. [Uji portal memakai fixture GraphQL](../../frontend/e2e/human_resource/3-query-slip-gaji-mandiri.md). |

## Penjelasan root bersama

`node` diuji untuk pekerja dengan grant badan hukum yang sesuai dan di luar scope. [Uji HTTPS](../../e2e/gateway/34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md).

## RBAC

Role uji diberikan melalui GraphQL IAM. Scope, izin, dan kebutuhan MFA mengikuti [matriks wiring](26-status-wiring-gateway-hrm.md). Ketiga Query slip memerlukan `human_resource.payslips.read_self` dengan scope SELF, [bukti penerbitan Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md), dan [akses DRM sesuai kepemilikan](24-kebutuhan-integrasi-drm.md) agar mengembalikan data.

Regresi renderer, gRPC PostgreSQL, dan GraphQL Gateway untuk slip mandiri lulus ulang dengan race detector. [Hasil uji](../../e2e/human_resource/60-payroll-query-regresi-slip-gaji-mandiri.md).
