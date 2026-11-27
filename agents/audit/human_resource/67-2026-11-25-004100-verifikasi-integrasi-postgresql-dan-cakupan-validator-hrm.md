# Audit 67 — Verifikasi Cakupan Validator dan Integrasi gRPC–PostgreSQL HRM

Tanggal audit: **25 November 2026 WIB**  
Scope: `backend/services/human_resource/` (`backend/gateway/` tidak tersentuh sesuai instruksi).

## Ringkasan Cakupan & Implementasi

1. **Cakupan Validator Integritas (`cmd/wiring.go` & `cmd/wiring_test.go`):**
   - `buildMutationValidators` kini mendaftarkan seluruh **34 builder validator** mencakup **78/78 resource mutasi katalog** ditambah `hr_leave_ledger_entries` (**79 validator total**).
   - `TestBuildMutationValidatorsCoversAllCatalogMutatableResources` memastikan tidak ada resource mutasi pada `PermissionCatalog()` yang lolos tanpa validator terdaftar, sedangkan `hr_policies` & `hr_policy_acknowledgements` dipisahkan bersih ke `infrastructure/security/hr_policy.go` agar `infrastructure/security/policy.go` tetap khusus policy otorisasi IAM.
   - Adapter Enterprise Foundation (`FoundationVerifier`) terhubung secara eksplisit pada `cmd/wiring.go`.

2. **Temuan Bug Nyata Selama Pengujian PostgreSQL & Perbaikannya:**
   - **`infrastructure/security/talent.go` (`talentPoolMembershipValidator`):** Perbandingan `entity.CompareDates(membership.ValidFrom, *assessment.FinalizedAt)` sebelumnya mengambil `.Date()` dari `TIMESTAMPTZ` dalam zona waktu lokal proses (`+07:00`), sedangkan `SystemClock.Now()` dan `DATE` memakai UTC. Hal ini menyebabkan penolakan salah saat pukul 00:00–06:59 WIB. Diperbaiki menggunakan `assessment.FinalizedAt.UTC()`.
   - **`transport/grpc/handler/policy/handler.go` (`CreatePolicy`):** Pemanggilan `response.Timestamp(in.ApprovedAt)` menolak request gRPC ketika `approved_at` dikosongkan meskipun `specPolicy().Initialize` mengisi nilai tersebut dari clock server. Diperbaiki agar `approved_at` opsional pada request create.
   - **`transport/grpc/handler/recruitment/test/detail_integration_test.go`:** Penyesuaian tujuan verifikasi bukti asesmen (`"recruitment_assessment_evidence"`) agar cocok dengan kontrak `CompleteApplicationAssessment` serta penegasan pengecekan error pada cleanup.
   - **`transport/grpc/handler/compensation/record_integration_test.go`:** Penambahan `t.Cleanup` eksplisit untuk menghapus fixture relasional (`hr_worker_assignments`, `hr_workers`, `hr_positions`, `hr_jobs`, `hr_job_families`) agar uji paket berikutnya (`TestPlannedPositionRPCWithPostgreSQL`) tidak melihat baris sisa pada database disposable bersama.

3. **Bukti Eksekusi Pasca-Edit Terakhir:**
   - `gofmt -s -w .` & `GOWORK=off go vet ./...` & `GOWORK=off go build ./...`: **Lulus (exit 0)**.
   - Unit tests `GOWORK=off go test -race -count=1 ./...`: **47/47 paket lulus** (0 gagal, 0 skip).
   - Integrasi PostgreSQL `GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...`: **66/66 paket lulus** terhadap PostgreSQL 18.7 disposable `hrm_disposable_e2e_test` (0 gagal, 0 skip; database sekali pakai dibersihkan sesudahnya), mencakup Batch A, Batch B (`payroll`, `talent`, `performance`, `promotion`), dan Batch C (`employeerelations`, `policy`, `hrservice`, `engagement`, `accessrequest`, `learning`).
   - Container Docker `integrione-retail-human-resource-1`: rebuilt dan berjalan **`healthy`** di `127.0.0.1:9003` dengan log `HRM mutation validators registered for all catalog resources` serta healthcheck gRPC `OK`.

## Batas Terbuka
- Pengujian gRPC `bufconn` tetap menggunakan stub policy IAM dan stub DRM/Finance; pembuktian grant IAM nyata, scope parsial non-worker, serta jalur GraphQL Gateway `/api/graphql` menjadi fokus tahap berikutnya.
