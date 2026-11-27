# List operasi

`human_resource.payroll.v1.PayrollService` memiliki 120 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/payroll/v1/payroll.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreatePayrollProfile` | Membuat entri `payroll_profiles`. | Mutation | `CreatePayrollProfileRequest` → `PayrollProfileResponse` | `human_resource.payroll_profiles.create` |
| `UpdatePayrollProfile` | Mengubah entri `payroll_profiles` sesuai field mask. | Mutation | `UpdatePayrollProfileRequest` → `PayrollProfileResponse` | `human_resource.payroll_profiles.update` |
| `DeletePayrollProfile` | Menghapus entri `payroll_profiles` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.payroll_profiles.delete` |
| `GetPayrollProfileByID` | Mengambil entri `payroll_profiles` berdasarkan ID. | Query | `IDRequest` → `PayrollProfileResponse` | `human_resource.payroll_profiles.read` |
| `GetAllPayrollProfiles` | Mendaftar entri `payroll_profiles` dengan paginasi. | Query | `PageRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByWorkerID` | Menyaring daftar `payroll_profiles` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByPayrollCountryCode` | Menyaring daftar `payroll_profiles` berdasarkan `payroll_country_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByPayrollJurisdictionCode` | Menyaring daftar `payroll_profiles` berdasarkan `payroll_jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByPayrollGroupID` | Menyaring daftar `payroll_profiles` berdasarkan `payroll_group_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByEmployingLegalEntityReferenceID` | Menyaring daftar `payroll_profiles` berdasarkan `employing_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByCurrencyCode` | Menyaring daftar `payroll_profiles` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByTaxProfileReferenceID` | Menyaring daftar `payroll_profiles` berdasarkan `tax_profile_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByBankAccountTokenReferenceID` | Menyaring daftar `payroll_profiles` berdasarkan `bank_account_token_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByTaxStatusCode` | Menyaring daftar `payroll_profiles` berdasarkan `tax_status_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `FindPayrollProfileByWithholdingElectionReferenceID` | Menyaring daftar `payroll_profiles` berdasarkan `withholding_election_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollProfilesResponse` | `human_resource.payroll_profiles.read` |
| `CreatePayComponent` | Membuat entri `pay_components`. | Mutation | `CreatePayComponentRequest` → `PayComponentResponse` | `human_resource.pay_components.create` |
| `UpdatePayComponent` | Mengubah entri `pay_components` sesuai field mask. | Mutation | `UpdatePayComponentRequest` → `PayComponentResponse` | `human_resource.pay_components.update` |
| `DeletePayComponent` | Menghapus entri `pay_components` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.pay_components.delete` |
| `GetPayComponentByID` | Mengambil entri `pay_components` berdasarkan ID. | Query | `IDRequest` → `PayComponentResponse` | `human_resource.pay_components.read` |
| `GetAllPayComponents` | Mendaftar entri `pay_components` dengan paginasi. | Query | `PageRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByComponentCode` | Menyaring daftar `pay_components` berdasarkan `component_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByComponentType` | Menyaring daftar `pay_components` berdasarkan `component_type`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByTaxTreatmentCode` | Menyaring daftar `pay_components` berdasarkan `tax_treatment_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByJurisdictionCode` | Menyaring daftar `pay_components` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByCurrencyCode` | Menyaring daftar `pay_components` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByRuleSetID` | Menyaring daftar `pay_components` berdasarkan `rule_set_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `FindPayComponentByFormulaReferenceID` | Menyaring daftar `pay_components` berdasarkan `formula_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayComponentsResponse` | `human_resource.pay_components.read` |
| `CreatePayrollPeriod` | Membuat entri `payroll_periods`. | Mutation | `CreatePayrollPeriodRequest` → `PayrollPeriodResponse` | `human_resource.payroll_periods.create` |
| `UpdatePayrollPeriod` | Mengubah entri `payroll_periods` sesuai field mask. | Mutation | `UpdatePayrollPeriodRequest` → `PayrollPeriodResponse` | `human_resource.payroll_periods.update` |
| `DeletePayrollPeriod` | Menghapus entri `payroll_periods` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.payroll_periods.delete` |
| `GetPayrollPeriodByID` | Mengambil entri `payroll_periods` berdasarkan ID. | Query | `IDRequest` → `PayrollPeriodResponse` | `human_resource.payroll_periods.read` |
| `GetAllPayrollPeriods` | Mendaftar entri `payroll_periods` dengan paginasi. | Query | `PageRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByPayrollGroupID` | Menyaring daftar `payroll_periods` berdasarkan `payroll_group_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByEmployingLegalEntityReferenceID` | Menyaring daftar `payroll_periods` berdasarkan `employing_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByCurrencyCode` | Menyaring daftar `payroll_periods` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByJurisdictionCode` | Menyaring daftar `payroll_periods` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByPeriodCode` | Menyaring daftar `payroll_periods` berdasarkan `period_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByPeriodStatus` | Menyaring daftar `payroll_periods` berdasarkan `period_status`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `FindPayrollPeriodByCalendarReferenceID` | Menyaring daftar `payroll_periods` berdasarkan `calendar_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollPeriodsResponse` | `human_resource.payroll_periods.read` |
| `CreatePayrollRun` | Membuat entri `payroll_runs`. | Mutation | `CreatePayrollRunRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.create` |
| `UpdatePayrollRun` | Mengubah entri `payroll_runs` sesuai field mask. | Mutation | `UpdatePayrollRunRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.update` |
| `DeletePayrollRun` | Menghapus entri `payroll_runs` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.payroll_runs.delete` |
| `GetPayrollRunByID` | Mengambil entri `payroll_runs` berdasarkan ID. | Query | `IDRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.read` |
| `GetAllPayrollRuns` | Mendaftar entri `payroll_runs` dengan paginasi. | Query | `PageRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByPayrollPeriodID` | Menyaring daftar `payroll_runs` berdasarkan `payroll_period_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByPayrollGroupID` | Menyaring daftar `payroll_runs` berdasarkan `payroll_group_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByEmployingLegalEntityReferenceID` | Menyaring daftar `payroll_runs` berdasarkan `employing_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByCurrencyCode` | Menyaring daftar `payroll_runs` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByRuleSetID` | Menyaring daftar `payroll_runs` berdasarkan `rule_set_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByOriginalRunID` | Menyaring daftar `payroll_runs` berdasarkan `original_run_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByRunNumber` | Menyaring daftar `payroll_runs` berdasarkan `run_number`, dengan paginasi. | Query | `Int32FinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByRunType` | Menyaring daftar `payroll_runs` berdasarkan `run_type`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByStatus` | Menyaring daftar `payroll_runs` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByJurisdictionCode` | Menyaring daftar `payroll_runs` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByInputSnapshotReferenceID` | Menyaring daftar `payroll_runs` berdasarkan `input_snapshot_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByApprovedByPrincipalID` | Menyaring daftar `payroll_runs` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByFinanceJournalReferenceID` | Menyaring daftar `payroll_runs` berdasarkan `finance_journal_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByPaymentBatchReferenceID` | Menyaring daftar `payroll_runs` berdasarkan `payment_batch_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByFinanceSettlementEventReferenceID` | Menyaring daftar `payroll_runs` berdasarkan `finance_settlement_event_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByStartedByPrincipalID` | Menyaring daftar `payroll_runs` berdasarkan `started_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `FindPayrollRunByIdempotencyKey` | Menyaring daftar `payroll_runs` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRunsResponse` | `human_resource.payroll_runs.read` |
| `CalculatePayrollRun` | Menghitung run draf melalui adapter kalkulator dan menyimpan snapshot serta komponennya secara atomik. | Mutation | `IDRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.calculate` |
| `CancelPayrollRun` | Membatalkan run dalam review dengan alasan dan menjaga snapshot audit. | Mutation | `CancelPayrollRunRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.cancel` |
| `SubmitPayrollRunForReview` | Penginisiasi mengajukan run yang telah dihitung dan direkonsiliasi untuk review. | Mutation | `IDRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.submit_review` |
| `RequestPayrollPosting` | Mencatat permintaan jurnal satu kali untuk run yang disetujui dan menerbitkan event outbox. | Mutation | `IDRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.request_posting` |
| `ApprovePayrollRun` | Menyetujui run setelah hasil payroll dan baris hasil lengkap serta lolos validasi rekonsiliasi. | Mutation | `IDRequest` → `PayrollRunResponse` | `human_resource.payroll_runs.approve` |
| `CreatePayrollResult` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `CreatePayrollResultRequest` → `PayrollResultResponse` | — |
| `UpdatePayrollResult` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `UpdatePayrollResultRequest` → `PayrollResultResponse` | — |
| `DeletePayrollResult` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `IDRequest` → `EmptyResponse` | — |
| `GetMyPayslip` | Membaca metadata slip sendiri setelah verifikasi penerbitan dan akses privat DRM. | Query | `GetMyPayslipRequest` → `GetMyPayslipResponse` | `human_resource.payslips.read_self` · scope self |
| `ListMyPayslips` | Mendaftar slip terbit milik pekerja, dengan paginasi dan filter periode opsional. | Query | `ListMyPayslipsRequest` → `ListMyPayslipsResponse` | `human_resource.payslips.read_self` · scope self |
| `ReissueEmployeePayslip` | Membuka PDF slip terbit untuk petugas HR dan menyiapkan versi PDF bila belum tertaut. | Mutation | `ReissueEmployeePayslipRequest` → `ReissueEmployeePayslipResponse` | `human_resource.payslips.reissue` · MFA dan scope badan hukum |
| `GetPayrollResultByID` | Mengambil entri `payroll_results` berdasarkan ID. | Query | `IDRequest` → `PayrollResultResponse` | `human_resource.payroll_results.read` |
| `GetAllPayrollResults` | Mendaftar entri `payroll_results` dengan paginasi. | Query | `PageRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByPayrollRunID` | Menyaring daftar `payroll_results` berdasarkan `payroll_run_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByPayrollGroupID` | Menyaring daftar `payroll_results` berdasarkan `payroll_group_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByWorkerID` | Menyaring daftar `payroll_results` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByPayrollProfileReferenceID` | Menyaring daftar `payroll_results` berdasarkan `payroll_profile_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByCurrencyCode` | Menyaring daftar `payroll_results` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByResultStatus` | Menyaring daftar `payroll_results` berdasarkan `result_status`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByPaymentReferenceID` | Menyaring daftar `payroll_results` berdasarkan `payment_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `FindPayrollResultByPayslipDocumentReferenceID` | Menyaring daftar `payroll_results` berdasarkan `payslip_document_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultsResponse` | `human_resource.payroll_results.read` |
| `CreatePayrollResultLine` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `CreatePayrollResultLineRequest` → `PayrollResultLineResponse` | — |
| `UpdatePayrollResultLine` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `UpdatePayrollResultLineRequest` → `PayrollResultLineResponse` | — |
| `DeletePayrollResultLine` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `IDRequest` → `EmptyResponse` | — |
| `GetPayrollResultLineByID` | Mengambil entri `payroll_result_lines` berdasarkan ID. | Query | `IDRequest` → `PayrollResultLineResponse` | `human_resource.payroll_result_lines.read` |
| `GetAllPayrollResultLines` | Mendaftar entri `payroll_result_lines` dengan paginasi. | Query | `PageRequest` → `PayrollResultLinesResponse` | `human_resource.payroll_result_lines.read` |
| `FindPayrollResultLineByPayrollResultID` | Menyaring daftar `payroll_result_lines` berdasarkan `payroll_result_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultLinesResponse` | `human_resource.payroll_result_lines.read` |
| `FindPayrollResultLineByComponentID` | Menyaring daftar `payroll_result_lines` berdasarkan `component_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollResultLinesResponse` | `human_resource.payroll_result_lines.read` |
| `CreatePayrollAdjustment` | Membuat entri `payroll_adjustments`. | Mutation | `CreatePayrollAdjustmentRequest` → `PayrollAdjustmentResponse` | `human_resource.payroll_adjustments.create` |
| `UpdatePayrollAdjustment` | Mengubah entri `payroll_adjustments` sesuai field mask. | Mutation | `UpdatePayrollAdjustmentRequest` → `PayrollAdjustmentResponse` | `human_resource.payroll_adjustments.update` |
| `DeletePayrollAdjustment` | Menghapus entri `payroll_adjustments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.payroll_adjustments.delete` |
| `GetPayrollAdjustmentByID` | Mengambil entri `payroll_adjustments` berdasarkan ID. | Query | `IDRequest` → `PayrollAdjustmentResponse` | `human_resource.payroll_adjustments.read` |
| `GetAllPayrollAdjustments` | Mendaftar entri `payroll_adjustments` dengan paginasi. | Query | `PageRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByWorkerID` | Menyaring daftar `payroll_adjustments` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByPayrollPeriodID` | Menyaring daftar `payroll_adjustments` berdasarkan `payroll_period_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByComponentID` | Menyaring daftar `payroll_adjustments` berdasarkan `component_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByAdjustmentType` | Menyaring daftar `payroll_adjustments` berdasarkan `adjustment_type`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByCurrencyCode` | Menyaring daftar `payroll_adjustments` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByRequestedByPrincipalID` | Menyaring daftar `payroll_adjustments` berdasarkan `requested_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByApprovedByPrincipalID` | Menyaring daftar `payroll_adjustments` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `FindPayrollAdjustmentByIdempotencyKey` | Menyaring daftar `payroll_adjustments` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollAdjustmentsResponse` | `human_resource.payroll_adjustments.read` |
| `CreatePayrollGroup` | Membuat entri `payroll_groups`. | Mutation | `CreatePayrollGroupRequest` → `PayrollGroupResponse` | `human_resource.payroll_groups.create` |
| `UpdatePayrollGroup` | Mengubah entri `payroll_groups` sesuai field mask. | Mutation | `UpdatePayrollGroupRequest` → `PayrollGroupResponse` | `human_resource.payroll_groups.update` |
| `DeletePayrollGroup` | Menghapus entri `payroll_groups` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.payroll_groups.delete` |
| `GetPayrollGroupByID` | Mengambil entri `payroll_groups` berdasarkan ID. | Query | `IDRequest` → `PayrollGroupResponse` | `human_resource.payroll_groups.read` |
| `GetAllPayrollGroups` | Mendaftar entri `payroll_groups` dengan paginasi. | Query | `PageRequest` → `PayrollGroupsResponse` | `human_resource.payroll_groups.read` |
| `FindPayrollGroupByGroupCode` | Menyaring daftar `payroll_groups` berdasarkan `group_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollGroupsResponse` | `human_resource.payroll_groups.read` |
| `FindPayrollGroupByEmployingLegalEntityReferenceID` | Menyaring daftar `payroll_groups` berdasarkan `employing_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollGroupsResponse` | `human_resource.payroll_groups.read` |
| `FindPayrollGroupByJurisdictionCode` | Menyaring daftar `payroll_groups` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollGroupsResponse` | `human_resource.payroll_groups.read` |
| `FindPayrollGroupByCurrencyCode` | Menyaring daftar `payroll_groups` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollGroupsResponse` | `human_resource.payroll_groups.read` |
| `CreatePayrollRuleSet` | Membuat entri `payroll_rule_sets`. | Mutation | `CreatePayrollRuleSetRequest` → `PayrollRuleSetResponse` | `human_resource.payroll_rule_sets.create` |
| `UpdatePayrollRuleSet` | Mengubah entri `payroll_rule_sets` sesuai field mask. | Mutation | `UpdatePayrollRuleSetRequest` → `PayrollRuleSetResponse` | `human_resource.payroll_rule_sets.update` |
| `DeletePayrollRuleSet` | Menghapus entri `payroll_rule_sets` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.payroll_rule_sets.delete` |
| `GetPayrollRuleSetByID` | Mengambil entri `payroll_rule_sets` berdasarkan ID. | Query | `IDRequest` → `PayrollRuleSetResponse` | `human_resource.payroll_rule_sets.read` |
| `GetAllPayrollRuleSets` | Mendaftar entri `payroll_rule_sets` dengan paginasi. | Query | `PageRequest` → `PayrollRuleSetsResponse` | `human_resource.payroll_rule_sets.read` |
| `FindPayrollRuleSetByRuleCode` | Menyaring daftar `payroll_rule_sets` berdasarkan `rule_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRuleSetsResponse` | `human_resource.payroll_rule_sets.read` |
| `FindPayrollRuleSetByJurisdictionCode` | Menyaring daftar `payroll_rule_sets` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRuleSetsResponse` | `human_resource.payroll_rule_sets.read` |
| `FindPayrollRuleSetByArtifactDocumentVersionReferenceID` | Menyaring daftar `payroll_rule_sets` berdasarkan `artifact_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRuleSetsResponse` | `human_resource.payroll_rule_sets.read` |
| `FindPayrollRuleSetByLegalBasisDocumentVersionReferenceID` | Menyaring daftar `payroll_rule_sets` berdasarkan `legal_basis_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRuleSetsResponse` | `human_resource.payroll_rule_sets.read` |
| `FindPayrollRuleSetByApprovedByPrincipalID` | Menyaring daftar `payroll_rule_sets` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PayrollRuleSetsResponse` | `human_resource.payroll_rule_sets.read` |

## Butuh GraphQL

88 kandidat Query dan 27 kandidat Mutation dan 6 RPC nonaktif. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

`ApprovePayrollRun` hanya berlaku pada periode payroll berstatus `open`. HRM memeriksa status periode dalam transaksi persetujuan.

`SubmitPayrollRunForReview` membutuhkan MFA, penginisiasi run, periode terbuka, dan hasil payroll yang lengkap serta konsisten. Command mencatat status `review` dan outbox. Persetujuan merupakan aksi terpisah dan hanya dapat dijalankan setelah run masuk status `review`.

`RequestPayrollPosting` membutuhkan MFA dan run yang sudah disetujui. Mutation Gateway `requestPayrollPosting` meneruskan ID run dan menampilkan tanda terima. Respons gRPC berisi `postingRequestedAt` dan `postingRequestedByPrincipalId`. Status tetap `approved` sampai bukti jurnal Finance diterima.

Event outbox `hr_payroll_runs.request_posting` menyertakan `details` dengan periode, grup, badan hukum, mata uang, arah settlement, dan digest SHA-256 snapshot hasil. Payload tidak berisi rincian gaji pekerja.

## Kebutuhan sinkronisasi Finance

Finance belum tersedia. Poin berikut adalah kebutuhan untuk menyepakati kontrak saat Finance dibangun.

[Handoff penerbitan dan settlement untuk Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md).

- **Batas tanggung jawab.** HRM menghitung, merekonsiliasi, dan mengajukan snapshot nominal beserta rincian payroll yang disetujui. Finance Accounting menerbitkan payroll dan memiliki otorisasi pencairan, eksekusi transfer per pekerja melalui bank atau penyedia pembayaran pihak ketiga, serta rekonsiliasi. HRM mencatat status dan referensi konfirmasi dari Finance.
- **Permintaan jurnal.** Finance perlu menerima event `hr_payroll_runs.request_posting` dari HRM dan menggunakan `event_id` sebagai kunci deduplikasi. Korelasikan permintaan dengan ID run, badan hukum, mata uang, arah settlement, dan digest snapshot yang disetujui.
- **Data jurnal.** Sepakati kontrak privat untuk memperoleh nominal agregat yang dibutuhkan Finance, termasuk pemetaan akun, unit biaya, dan perlakuan atas koreksi atau recovery. Tentukan satuan, pembulatan, serta rekonsiliasi terhadap snapshot yang dirujuk. Event permintaan saat ini tidak membawa nominal.
- **Bukti posting.** Finance perlu mengembalikan konfirmasi autentik yang mengikat ID event, ID run, ID jurnal, badan hukum, mata uang, dan digest snapshot. HRM menggunakan bukti tersebut untuk transisi `approved → posted`.
- **Penerbitan payroll.** Finance mengunci versi yang disetujui, mencatat kewajiban gaji, menetapkan jadwal pembayaran, dan mengirim konfirmasi penerbitan eksplisit. Konfirmasi memuat ID dan versi penerbitan, ID run, digest snapshot, referensi jurnal, tanggal penerbitan, serta tanggal pembayaran terjadwal. HRM memvalidasi penerbitan terhadap bukti posting sebelum membuat slip. Jika bukti posting belum diterima, tahan pemrosesan sampai bukti cocok. Finance mengembalikan pengajuan yang perlu dikoreksi beserta alasannya.
- **Pembayaran.** Petugas Finance menyiapkan instruksi pembayaran dan petugas Finance berwenang lain menyetujuinya. Finance mengirim instruksi melalui bank atau penyedia pembayaran dan merekonsiliasi hasil per pekerja. Finance menyediakan status terjadwal, diproses, dibayar, atau gagal dari konfirmasi tepercaya. Retry harus menghindari transfer ganda, termasuk ketika hasil percobaan sebelumnya belum diketahui.
- **Bukti settlement.** HRM memproyeksikan konfirmasi pembayaran per hasil payroll. Pembayaran sebagian tidak mengubah seluruh run menjadi `paid`. Transisi agregat `posted → paid` hanya berlaku setelah semua pembayaran yang diwajibkan selesai dan Finance mengonfirmasi settlement final. Skema HRM memerlukan referensi jurnal, batch pembayaran, event settlement, dan waktu settlement. Koreksi dan reversal menyimpan jejak audit.
- **Pengiriman dan rekonsiliasi.** Sepakati identitas workload, versi payload, urutan event, retry, dan penanganan event berulang dengan digest berbeda. Simpan receipt inbox dan perubahan status HRM dalam satu transaksi. Uji ulang pengiriman, event di luar urutan, referensi yang tidak cocok, dan pemulihan setelah gangguan.
- **Gateway.** Tampilkan permintaan posting, konfirmasi jurnal, penerbitan payroll, dan status pembayaran sebagai tahap yang dapat dibedakan. Pengguna HRM meminta posting melalui Mutation yang sudah ada. Bukti jurnal dan settlement berasal dari integrasi privat Finance.

Setelah run masuk `review`, PostgreSQL membekukan snapshot run, hasil, dan baris komponen. Persetujuan hanya menambah status dan metadata persetujuan pada snapshot tersebut.

Jika reviewer menemukan kesalahan, `CancelPayrollRun` memungkinkan pembatalan dari status `review`. Principal manusia dengan MFA harus berbeda dari penginisiasi dan memberi alasan. Respons membawa alasan, pelaku, dan waktu. Snapshot dibekukan dan run pengganti memakai nomor serta snapshot baru. Koreksi setelah approval mengikuti alur correction atau reversal.

`CalculatePayrollRun` memerlukan penginisiasi run, MFA, periode terbuka, grup payroll aktif, serta adapter kalkulator yang membuktikan kecocokan snapshot input dan rule set. Runtime tanpa adapter mengembalikan `Unavailable`.
Adapter menerima ID versi dokumen dan digest SHA-256 artefak aturan. Adapter harus memverifikasi bytes artefak yang dipakai dan mengembalikan digest tersebut bersama hasil. HRM menolak hasil jika digest berbeda dari rule set yang disetujui.
Adapter juga menerima tanggal awal dan akhir periode serta tanggal pembayaran dari periode yang dikunci dalam transaksi. Gunakan tanggal kalender tersebut sebagai input kalkulasi.

HRM memeriksa bahwa profil pekerja dan komponen aktif serta memiliki masa berlaku yang bersinggungan dengan periode run. Nama, kode, dan tipe komponen pada hasil harus cocok dengan versi katalog yang dirujuk.

Penyesuaian payroll memerlukan periode terbuka, pekerja aktif yang sudah mulai bekerja paling lambat pada akhir periode, serta profil aktif yang mencakup periode dan cocok dengan badan hukum, yurisdiksi, dan mata uang.

Validasi menelusuri profil aktif per halaman sehingga riwayat profil tidak membatasi pencarian pada halaman pertama.

Pembuatan dan perubahan profil payroll memeriksa benturan masa berlaku dengan riwayat profil pekerja pada grup yang sama, termasuk ketika hasil pencarian mencakup beberapa halaman.

Tanggal pembayaran dapat diubah selama periode terbuka dan belum dirujuk oleh run atau penyesuaian payroll. Setelah itu, HRM menolak perubahan tanggal melalui `UpdatePayrollPeriod`.

Adapter kalkulator menerima `CurrencyScale` dari grup payroll sebagai presisi pembulatan. HRM menolak perubahan presisi grup setelah dirujuk profil, periode, atau run.

HRM menolak hasil ketika nominal bersih melebihi presisi grup. Angka nol di akhir pecahan tidak menambah presisi nominal.

Rekonsiliasi seluruh run menolak ID baris komponen yang dipakai ulang, walaupun baris tersebut berada pada hasil pekerja yang berbeda.

Tes gRPC mencakup field mask tanggal pembayaran dan penolakan perubahan setelah penyesuaian dibuat. [Laporan pengujian](../../e2e/human_resource/36-payroll-mutation-update-payroll-period.md)

## Slip gaji PDF mandiri

Setelah Finance Accounting menerbitkan payroll, sistem membuat satu PDF slip gaji per pekerja dari snapshot hasil yang telah dibekukan. Konfirmasi penerbitan dari Finance memicu pembuatan PDF. Begitu versi privat DRM siap dan referensinya tersimpan, karyawan dapat mengakses slipnya meskipun pembayaran masih terjadwal. Tampilkan status pembayaran secara terpisah di portal. Status `Dibayar` hanya muncul setelah Finance mengonfirmasi keberhasilan transfer untuk karyawan tersebut. Status run `posted` atau `paid` tetap menjadi prasyarat penautan PDF di HRM.

- HRM membentuk slip dari hasil pekerja, periode, badan hukum, mata uang, serta rincian penghasilan dan potongan yang perlu tampil. DRM menyimpan versi PDF immutable dengan klasifikasi data sensitif. HRM mengaitkan versi tersebut ke hasil payroll melalui `payslip_document_reference_id`. Kunci idempotensi memakai ID hasil payroll dan versi penerbitan Finance agar retry setelah kegagalan DRM atau penyimpanan referensi tidak membuat PDF duplikat. Koreksi payroll menghasilkan slip pengganti dengan riwayat versi yang dapat ditelusuri.
- **Butuh GraphQL Query** untuk daftar slip pekerja sendiri dan akses PDF dari profil. Gateway dan DRM memverifikasi hubungan principal IAM dengan worker pemilik hasil di server. Employee dapat membuka, mengunduh, dan mencetak PDF langsung. ID dokumen atau filter yang dikirim klien tidak memberikan akses ke slip pekerja lain.
- **RBAC** untuk akses mandiri perlu disepakati dengan IAM. Grant baca payroll staf dan akses slip pribadi memiliki cakupan berbeda. Pembuatan PDF berjalan otomatis dari event penerbitan, sehingga employee tidak perlu mengajukan Mutation untuk meminta slip.
- Uji bahwa PDF belum dapat diakses sebelum diterbitkan, retry tidak membuat versi duplikat, akses sendiri berhasil saat pembayaran masih terjadwal, akses pekerja lain ditolak, pembayaran parsial hanya memperbarui status pekerja yang berhasil, dan slip koreksi dapat ditelusuri.

`FindPayrollResultByPayslipDocumentReferenceID` adalah finder internal dengan izin `human_resource.payroll_results.read`. Query mandiri memerlukan izin tersendiri dan filter pemilik yang diterapkan sebelum paginasi maupun hitungan. Derivasikan worker dari principal IAM terverifikasi dan kaitannya di HRM. Respons mandiri cukup memuat metadata slip yang terbit dan referensi akses dokumen DRM.

Baseline PostgreSQL mengizinkan pengisian awal referensi slip hanya setelah run berstatus `posted` atau `paid`, lalu melarang penggantian referensi. Jalur penerbitan perlu memeriksa konfirmasi penerbitan dari Finance dan menyimpan referensi versi yang berhasil diverifikasi. Slip koreksi perlu dirujuk melalui hasil koreksi, dengan tautan riwayat ke slip sebelumnya. [Invariant PostgreSQL](../../../backend/services/human_resource/db/schema/postgresql-invariants.sql)

Tes penyimpanan referensi dan pembacaan staf menggunakan fixture Finance dan DRM. [Laporan pengujian](../../e2e/human_resource/37-payroll-query-get-payroll-result-payslip.md)

Satu referensi versi slip DRM hanya dapat dipakai oleh satu hasil payroll. Hasil yang slipnya belum tersedia tetap dapat menyimpan nilai kosong.

Tes policy HRM mengonfirmasi finder hasil payroll tetap membutuhkan izin staf dalam scope yang didukung dan MFA terbaru. Grant `self` untuk kode izin staf ditolak, termasuk saat MFA masih berlaku.

Pengujian mencakup pembuatan run pengganti pada periode yang sama tanpa membuka kembali run lama. [Laporan pengujian pembatalan](../../e2e/human_resource/39-payroll-mutation-cancel-payroll-run.md)

Sebelum memanggil kalkulator, HRM memeriksa waktu persetujuan rule set. Rule set tanpa waktu persetujuan atau dengan waktu persetujuan setelah waktu kalkulasi ditolak.

`CreatePayrollRun` menolak rule set yang belum disetujui pada waktu pembuatan run. Rule set dengan persetujuan tepat pada waktu pembuatan memenuhi batas waktu tersebut.

Kalkulator hanya dipanggil jika yurisdiksi rule set cocok dan masa berlakunya mencakup seluruh periode payroll. Startup produksi masih memerlukan adapter yang dapat memperoleh snapshot input dan artefak rule set terverifikasi.

`SubmitPayrollRunForReview` dan `ApprovePayrollRun` memeriksa ulang apakah rule set telah disetujui pada waktu kalkulasi yang tercatat di run. Snapshot lama yang mendahului persetujuan rule set ditolak dengan `FailedPrecondition`.

`RequestPayrollPosting` memeriksa ulang waktu kalkulasi terhadap waktu persetujuan rule set. Jika snapshot dihitung sebelum rule set disetujui, HRM menolak permintaan posting sebelum mencatat event outbox Finance.

Review, persetujuan, dan pengajuan posting memerlukan digest input dan output serta digest kalkulasi setiap hasil pekerja, masing-masing 32 byte untuk SHA-256. Snapshot lama dengan panjang berbeda ditolak sebelum hasil direkonsiliasi.

Rekonsiliasi sebelum review, persetujuan, dan posting menerima hasil pekerja berstatus `calculated` atau `approved`. Referensi pembayaran dan slip baru boleh muncul setelah proses Finance yang sesuai.

### Kontrak Query slip pribadi untuk Gateway

- `myPayslips` menerima paginasi dan filter periode opsional. `myPayslip` menerima ID hasil payroll milik pekerja. Kedua Query memerlukan sesi employee dan grant IAM `human_resource.payslips.read_self` dengan scope `self`. Kedua query tersedia pada Gateway. Konten PDF diminta melalui `myPayslipPdf` dengan grant SELF yang sama.
- HRM mencari worker melalui `FindByIAMPrincipalReferenceID` menggunakan principal dari sesi yang telah diverifikasi. Beberapa hubungan kerja dapat dipakai jika semuanya menunjuk party orang yang sama. API menolak akses jika hubungan tidak ditemukan atau menunjuk party orang berbeda. Gateway tidak menerima worker ID sebagai dasar otorisasi.
- Daftar dan hitungan memakai filter worker pemilik serta slip yang telah siap. Query detail mengambil metadata periode dan dokumen hanya setelah filter kepemilikan diterapkan. ID versi DRM yang diberikan klien tidak menjadi kunci untuk membaca slip.
- DRM memverifikasi akses ke versi PDF yang tertaut pada hasil tersebut melalui kontrak akses privat. Respons mandiri hanya memuat periode, tanggal terbit, tanggal pembayaran terjadwal, status pembayaran per pekerja, dan akses PDF. Hak akses staf atas hasil payroll tidak memberikan akses mandiri ke slip pekerja lain.
- RPC privat `GetMyPayslip` dan `ListMyPayslips` tersedia. Query Gateway `myPayslip` dan `myPayslips` terhubung dan telah diuji dengan Payroll gRPC fixture. Respons memuat metadata dan ID versi DRM. Unduhan PDF mengikuti kontrak akses privat DRM. Adapter daftar menerima seluruh ID hubungan kerja milik satu orang dan harus menghitung total setelah membatasi cakupan tersebut, periode, penerbitan, dan kesiapan DRM. HRM memeriksa ulang setiap hasil terhadap worker, periode, dan referensi dokumen yang tersimpan. Jalur runtime memerlukan bukti penerbitan Finance, referensi PDF DRM, dan provisioning grant IAM yang disetujui. Status `posted` sendiri tidak cukup untuk membuka akses slip.

Baseline HRM menyediakan indeks pencarian principal IAM yang terisi. Indeks mempertahankan kemungkinan beberapa hubungan kerja historis. Resolver mandiri tetap harus mendeteksi pemetaan ambigu sebelum membaca slip.

Service internal `PublishPayslipForResult` memproses satu hasil dari konfirmasi penerbitan Finance. Service memverifikasi ID penerbitan, run, jurnal, badan hukum, dan digest sebelum memanggil adapter PDF–DRM. Referensi DRM disimpan pada hasil payroll setelah total dan rincian diperiksa. Retry setelah referensi tersimpan mengembalikan versi yang sama. Jika DRM sudah menyimpan PDF dan transaksi HRM gagal, DRM harus mengembalikan versi yang sama untuk kunci penerbitan dan hasil payroll tersebut. Consumer Finance dan adapter produksi PDF–DRM masih diperlukan.

Renderer HRM menghasilkan PDF A4 dari snapshot final dan font TrueType Unicode yang disediakan konfigurasi deployment. PDF memuat periode, pekerja, badan hukum, rincian penghasilan serta potongan, dan nominal bersih. Adapter menyertakan SHA-256 konten dan kunci idempotensi saat menyerahkan PDF ke penyimpanan privat DRM. Font, client DRM, dan verifier penerbitan Finance masih perlu dipasang pada startup produksi.

Consumer privat Finance dapat menjalankan `PublishPayslipsForRun` dengan ID penerbitan dan ID run. HRM memproses setiap hasil dalam transaksi tersendiri. Jika proses berhenti di tengah run, pengiriman ulang melanjutkan hasil yang referensinya belum tersimpan. ACK event dilakukan setelah semua hasil selesai. Consumer jaringan dan kontrak pengiriman Finance masih terbuka.

Verifier penerbitan mengembalikan timestamp Finance yang stabil untuk ID penerbitan. HRM menolak timestamp kosong atau sebelum persetujuan run. Renderer mencetak tanggal penerbitan dalam UTC dan menggunakan timestamp yang sama pada metadata PDF agar retry menghasilkan konten identik.

Handler internal inbox menerima bentuk awal event `payroll.published` v1 dari `finance_accounting` dengan ID penerbitan dan ID run. Kontrak transport wajib mengisi identitas sumber dari autentikasi workload. Handler memvalidasi jenis event dan meneruskan event ke pemroses run. Bukti penerbitan tetap diperiksa oleh verifier Finance. Nama event dan mekanisme pengiriman perlu disinkronkan saat Finance dibangun. Handler ini belum terhubung ke transport durable atau tabel receipt inbox.

Referensi versi DRM dan ID penerbitan Finance baru disimpan bersamaan. ID penerbitan tidak dapat diganti setelah slip tertaut. Replay hanya mengembalikan versi lama jika ID penerbitan cocok. Referensi historis tanpa ID penerbitan tetap bisa dibaca melalui jalur verifikasi akses yang berlaku, tetapi pemroses penerbitan menolaknya sampai korelasi historis diselesaikan.

Verifier baca slip harus mengembalikan ID versi DRM dan ID penerbitan Finance. HRM mencocokkan keduanya dengan hasil payroll saat ID penerbitan sudah tercatat. Untuk referensi historis yang belum menyimpan ID penerbitan, verifier tetap harus mengembalikan UUID penerbitan Finance yang sah. ID penerbitan internal tidak perlu diterbitkan ke GraphQL.

Daftar dan detail slip menolak timestamp penerbitan yang mendahului persetujuan run atau belum tercapai menurut waktu HRM. Tanggal pembayaran terjadwal tetap boleh di masa depan.

Respons daftar dan detail slip memuat ID periode, kode periode, serta tanggal awal dan akhir dari data HRM. HRM memeriksa kembali status run dan kesesuaian filter periode sebelum mengembalikan respons.

Jika satu pekerja sudah dibayar saat run masih `posted`, penerbitan slip tetap bisa dilanjutkan setelah bukti penerbitan diverifikasi. Hasil pekerja berstatus `paid` memerlukan referensi pembayaran Finance. Status run `paid` dapat menyusul setelah penyelesaian pembayaran seluruh run.

## Query status run pada Gateway

`payrollRun` dan `payrollRuns` memanggil RPC baca run dengan izin global `human_resource.payroll_runs.read`. Respons merangkum status HRM, waktu pengajuan jurnal, referensi jurnal Finance dan waktu settlement aktual jika tersedia. Status publikasi slip tetap diperiksa oleh RPC slip mandiri. [Matriks pengujian](../../e2e/gateway/20-hrm-query-status-run-payroll.md).

`submitPayrollRunForReview` dan `approvePayrollRun` meneruskan command review dan persetujuan ke HRM. Persetujuan tetap terikat pemisahan tugas di HRM. Jika respons pengajuan jurnal hilang, baca ulang status run melalui `payrollRun` untuk memeriksa waktu pengajuan sebelum mencoba kembali. [Matriks pengujian](../../e2e/gateway/22-hrm-mutation-review-dan-approve-payroll.md).

`cancelPayrollRun` meneruskan ID run dan alasan pembatalan ke HRM. Respons GraphQL mencantumkan status dan `cancelled_at`. [Matriks pengujian](../../e2e/gateway/23-hrm-mutation-cancel-payroll-run.md).

`calculatePayrollRun` telah terhubung ke gRPC dan meneruskan `UNAVAILABLE` bila runtime HRM belum memasang kalkulator. `calculated_at` membedakan hasil kalkulasi dari tahap pengajuan jurnal dan settlement. [Matriks pengujian](../../e2e/gateway/24-hrm-mutation-calculate-payroll-run.md).

`createPayrollRun` menerima referensi snapshot dan digest SHA-256 berformat heksadesimal. Ulangi kunci idempotensi yang sama untuk permintaan yang sama. HRM membedakan duplikasi dengan conflict dan menetapkan status draf. [Matriks pengujian](../../e2e/gateway/25-hrm-mutation-create-payroll-draft.md).

Jika respons pembuatan draf terputus, gunakan `payrollRuns(idempotency_key: ...)` dengan kunci asli untuk memeriksa apakah HRM telah mencatat run. Query menerapkan izin baca run global dan mengembalikan pasangan ID serta status run melalui field yang diminta. [Matriks pengujian](../../e2e/gateway/30-hrm-query-pemulihan-draf-payroll.md).

Kebutuhan snapshot input, artefak dan engine untuk aktivasi kalkulator runtime dicatat pada [handoff kalkulator payroll](27-kebutuhan-kalkulator-payroll-runtime.md).

Detail dan daftar slip menggunakan periode dari run HRM. Jika verifier menyertakan ID periode, HRM memeriksa bahwa ID itu cocok sebelum mengisi metadata tampilan. Respons dengan periode berbeda atau rentang tanggal terbalik ditolak. Verifier dapat menyerahkan ID periode kosong jika metadata periode berasal sepenuhnya dari HRM.

RPC privat `GetMyPayslip` mendukung `include_pdf` untuk membaca satu versi PDF melalui verifier yang memiliki kemampuan baca DRM. Opsi bawaan tetap metadata. Uji gRPC dan PostgreSQL membuktikan penolakan versi berbeda, digest salah, konten melewati batas, dan permintaan pekerja lain sebelum DRM dipanggil. [Kontrak DRM](24-kebutuhan-integrasi-drm.md#kontrak-baca-pdf-pada-hrm).

## Butuh GraphQL Query untuk PDF

`myPayslipPdf` mengembalikan isi PDF base64 melalui `/api/graphql`. Akses memakai izin `human_resource.payslips.read_self` pada principal sesi dan hasil payroll yang dimiliki pekerja. Portal membuka dokumen di penampil browser setelah PDF berhasil disiapkan. List dan detail metadata tetap terpisah dari pembacaan konten.

## Periode payroll dan badan hukum

Pembuatan dan perubahan periode payroll memeriksa ke EF bahwa badan hukum masih aktif serta rentang awal sampai akhir periode berada dalam rentang pendirian dan pembubaran yang tercatat. Tanggal pembayaran tetap mengikuti validasi periode tersendiri. Verifikasi memakai RPC badan hukum EF yang sudah ada. Referensi kalender eksternal masih menghasilkan `UNAVAILABLE` sampai ada kontrak dari pemilik kalender. [Adapter EF](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go)

Validasi run yang sebelumnya hanya memeriksa status badan hukum kini memeriksa ulang rentang periode. Ini mencakup run dengan periode yang tersimpan sebelum validasi tanggal diberlakukan. Regresi gRPC memakai verifier EF fixture untuk menolak periode di luar rentang sebelum membuat periode yang sah dengan kode yang sama.

Pengajuan review dan persetujuan memeriksa ulang skala gaji bersih dari grup payroll, termasuk hasil lama yang sudah tersimpan. Nilai dengan presisi fraksi melebihi skala ditolak tanpa menyimpan perubahan status atau event. Presisi baris komponen dan basis tarif tetap mengikuti kontrak penyimpanan.

`RequestPayrollPosting` menggunakan pemeriksaan hasil yang sama. Regresi dengan run lama yang sudah berstatus approved membuktikan bahwa gaji bersih di luar skala grup tidak menulis waktu/pelaku pengajuan atau event ke Finance.

## Penjelasan ConfirmPayrollPosting

HRM memproyeksikan status `posted` dan referensi jurnal setelah verifier Finance mencocokkan bukti dengan snapshot yang telah diajukan. Handler privat menolak identitas management lain. Retry dengan jurnal yang sudah tercatat berhasil secara idempoten, dan jurnal berbeda ditolak. Operasi ini tidak memerlukan Query atau Mutation GraphQL. Regresi PostgreSQL memasang verifier fixture dan menggunakan handler konfirmasi, lalu melanjutkan penerbitan dan pemrosesan PDF melalui fixture DRM. [Kontrak Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md)

## Penjelasan ConfirmPayrollPayment

Operasi internal HRM ini hanya mencatat status paid dan referensi pembayaran individual setelah verifier Finance mengonfirmasi bukti. Replay referensi yang sama berhasil dan referensi berbeda ditolak. Status run tetap mengikuti konfirmasi agregat terpisah. Tidak diperlukan Query atau Mutation baru. Regresi PostgreSQL memakai verifier fixture dan membaca status hasil serta run melalui gRPC. [Catatan Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md)

## Penjelasan ConfirmPayrollSettlement

Setelah semua hasil memiliki bukti individual, HRM menerima verifikasi settlement Finance untuk mencatat status agregat `paid`. Batch dan ID settlement yang sama dapat diulang. Penggantian referensi ditolak. Operasi privat ini tidak menambah API GraphQL. Regresi PostgreSQL membuktikan status tetap `posted` saat baru satu hasil dikonfirmasi dan menjadi `paid` hanya setelah bukti agregat fixture diverifikasi. [Kontrak Finance](../../notes/finance_accounting/3-human_resource-payroll-publication.md)
