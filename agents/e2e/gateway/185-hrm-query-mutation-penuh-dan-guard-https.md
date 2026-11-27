# 185 — E2E & Guard HTTPS Penuh Query–Mutation HRM (176 Query + 326 Mutation)

- **Tanggal:** 27 November 2026
- **Target Endpoint:** `https://erp.localhost:8443/api/graphql` (Docker Compose live) dan `https://erp.localhost:18443/api/graphql` (`check_hrm_schema_guards.py` & `run_hrm_employee_provisioning.py`)

## Matriks Pengujian & Hasil Aktual

| Kasus Pengujian | Perintah / Jalur | Hasil |
| --- | --- | --- |
| Validasi RBAC Manifest & Policy gRPC HRM | `go test -race ./...` di `backend/shared/rbac` & `TestHRMPermissionsInManifestMatchGRPCPolicy` | **PASS** — `409` kode izin `human_resource.*` (`434` total RBAC) cocok 100% dengan `security.PermissionCatalog()` & `IsSensitiveOperation()`. |
| Validasi Linter & Unit/Server Gateway (`-race`) | `golangci-lint run ./internal/adapter/human_resource/... ./graph/resolvers/...` & `go test -race ./internal/adapter/... ./internal/server ./graph/resolvers/...` | **PASS** — `0 issues` lint; `enterprise_foundation`, `human_resource` (termasuk `coverage_test.go`: consolidated finder `>1` filter, MFA/IDOR, composite key `LearningCourseCompetency`, `FieldMask`, base64), `iam`, dan `internal/server` semua lulus. |
| Guard SDL & HTTPS Seluruh Operasi HRM | `python3 backend/gateway/internal/e2e/scripts/check_hrm_schema_guards.py` | **PASS** — `SDL: 86 Node types, 86 Relay connections`; **Query `176 / 176`**, **Mutation `326 / 326`** lulus (0 nama `hrm`, kontrak `Node`/`Connection`/`MutationResult`, guard 401/403). |
| Introspection HTTPS Live via Caddy (`8443`) | `POST https://erp.localhost:8443/api/graphql` | **PASS** — `267` total Query & `423` total Mutation aktif melayani `learningCourseCompetency`, `recordTalentCalibrationDecision`, `acknowledgePolicy`, `approveSuccessionPlan`, dll. |
| Runner Provisioning & Freeze Payroll (`--guards --freeze-payroll`) | `python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py --source-root . --freeze-payroll` | **PARSIAL** — `HTTPS freezePayrollInput + createPayrollRun E2E OK`, `createWorker`/`activateWorker`/`requestEmployeeAccount`/MFA/scoped payslip lulus; `TestHRMEmployeeProvisioningOverHTTPS` gagal pada `createEmployeeUser` (`identity service unavailable` pada kontainer IAM terisolasi). `--termination` tidak dijalankan karena overlay termination tidak aktif pada `docker-compose.yml` standar. |

## Catatan Dependensi Pemilik

Operasi produksi yang memerlukan adapter runtime Finance (verifikasi anggaran/clearance/settlement) dan DRM (bukti versi dokumen immutable) tetap fail-closed di Gateway/HRM sampai service pemilik terpasang.
