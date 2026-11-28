# Matriks operasi Finance Accounting pada Gateway dan gRPC

Baseline kode produksi 29 November 2026 WIB. Hitungan hanya mencakup field root milik Finance dan RPC yang dideklarasikan oleh service Finance. Input, payload, Node global dan field bersarang tidak menjadi operasi tambahan. Satu field dapat memanggil lebih dari satu RPC.

| Query | Mutation | Total root | Service gRPC | RPC | RPC dipakai Gateway | RPC tanpa root Finance |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 11 | 34 | 45 | 11 | 74 | 43 | 31 |

## Operasi dengan field Gateway

Prefix SDL adalah `backend/gateway/graphql/finance_accounting/`. Resolver berada di `backend/gateway/graph/resolvers/`, dan adapter di `backend/gateway/internal/adapter/finance_accounting/`. Tabel merekam wiring source. Publikasi runtime dibuktikan terpisah melalui HTTPS.

| Jenis | Field GraphQL | Method adapter | RPC yang dipanggil | SDL / resolver |
| --- | --- | --- | --- | --- |
| Query | `accountingSettings` | `AccountingSettings` | `GetAccountingSettings` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Query | `glAccount` | `GLAccount` | `GetGLAccount` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Query | `glAccounts` | `GLAccounts` | `ListGLAccounts` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Query | `journalEntry` | `JournalEntry` | `GetJournalEntry` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Query | `journalEntries` | `JournalEntries` | `ListJournalEntries` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `configureAccounting` | `ConfigureAccounting` | `ConfigureAccounting` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `createFiscalYear` | `CreateFiscalYear` | `CreateFiscalYear` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `createAccountingPeriod` | `CreateAccountingPeriod` | `CreateAccountingPeriod` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `openAccountingPeriod` | `OpenAccountingPeriod` | `OpenAccountingPeriod` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `closeAccountingPeriod` | `CloseAccountingPeriod` | `CloseAccountingPeriod` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `reopenAccountingPeriod` | `ReopenAccountingPeriod` | `ReopenAccountingPeriod` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `closeFiscalYear` | `CloseFiscalYear` | `CloseFiscalYear` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `createGLAccount` | `CreateGLAccount` | `CreateGLAccount` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `configurePostingRule` | `ConfigurePostingRule` | `ConfigurePostingRule` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `configurePayrollPostingRule` | `ConfigurePayrollPostingRule` | `ConfigurePostingRule` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `configureFinancePostingRule` | `ConfigureFinancePostingRule` | `ConfigurePostingRule` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `createJournalDraft` | `CreateJournalDraft` | `CreateJournalDraft` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `approveJournal` | `ApproveJournal` | `ApproveJournal` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `postJournal` | `PostJournal` | `PostJournal` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Mutation | `reverseJournal` | `ReverseJournal` | `ReverseJournal` | finance_accounting.graphql / finance_accounting.resolvers.go |
| Query | `apInvoice` | `APInvoice` | `GetAPInvoice` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Query | `apInvoices` | `APInvoices` | `ListAPInvoices` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `postServiceAcceptanceCreditNote` | `ToMutationErrors`, `PostServiceAcceptanceCreditNote` | `PostServiceAcceptanceCreditNote` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `approveServiceAcceptanceCreditNote` | `ToMutationErrors`, `ApproveServiceAcceptanceCreditNote` | `ApproveServiceAcceptanceCreditNote` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `createServiceAcceptanceCreditNote` | `ToMutationErrors`, `CreateServiceAcceptanceCreditNote` | `CreateServiceAcceptanceCreditNote` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `createAPInvoice` | `ToMutationErrors`, `CreateAPInvoice` | `CreateAPInvoice` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `performThreeWayMatching` | `ToMutationErrors`, `PerformThreeWayMatching` | `PerformThreeWayMatching` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `approveAPInvoice` | `ToMutationErrors`, `ApproveAPInvoice` | `ApproveAPInvoice` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `postAPInvoice` | `ToMutationErrors`, `PostAPInvoice` | `PostAPInvoice` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `createPaymentDraft` | `ToMutationErrors`, `CreatePaymentDraft` | `CreatePaymentDraft` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `approvePayment` | `ToMutationErrors`, `ApprovePayment` | `ApprovePayment` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `confirmPaymentEvidence` | `ToMutationErrors`, `ConfirmPaymentEvidence` | `ConfirmPaymentEvidence` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `postPayment` | `ToMutationErrors`, `PostPayment` | `PostPayment` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `allocatePayrollPayment` | `ToMutationErrors`, `AllocatePayrollPayment` | `AllocatePayrollPayment` | finance_accounts_payable.graphql / finance_accounts_payable.resolvers.go |
| Mutation | `createBudgetDraft` | `CreateBudgetDraft` | `CreateBudgetDraft`, `GetBudget` | finance_budget.graphql / finance_budget.resolvers.go |
| Mutation | `approveBudget` | `ApproveBudget` | `ApproveBudget`, `GetBudget` | finance_budget.graphql / finance_budget.resolvers.go |
| Mutation | `issueBudgetAuthorization` | `IssueBudgetAuthorization`, `ParseTypedID` | `IssueHRBudgetAuthorization` | finance_budget.graphql / finance_budget.resolvers.go |
| Mutation | `revokeBudgetAuthorization` | `RevokeBudgetAuthorization` | `RevokeHRBudgetAuthorization` | finance_budget.graphql / finance_budget.resolvers.go |
| Query | `budget` | `Budget` | `GetBudget` | finance_budget.graphql / finance_budget.resolvers.go |
| Mutation | `reviewFinanceClearance` | `ReviewFinanceClearance` | `ReviewFinanceClearance` | finance_payroll.graphql / finance_payroll.resolvers.go |
| Mutation | `publishPayrollRun` | `PublishPayrollRun` | `PublishPayroll` | finance_payroll.graphql / finance_payroll.resolvers.go |
| Mutation | `completePayrollSettlement` | `CompletePayrollSettlement` | `CompletePayrollSettlement` | finance_payroll.graphql / finance_payroll.resolvers.go |
| Query | `trialBalance` | `TrialBalance` | `GetTrialBalance` | finance_reporting.graphql / finance_reporting.resolvers.go |
| Query | `profitAndLoss` | `ProfitAndLoss` | `GetProfitAndLoss` | finance_reporting.graphql / finance_reporting.resolvers.go |
| Query | `taxSummary` | `TaxSummary` | `GetTaxSummary` | finance_reporting.graphql / finance_reporting.resolvers.go |

## RPC yang belum memiliki field Gateway

17 RPC berpolicy workload-only dan 14 RPC menerima manusia menurut `backend/services/finance_accounting/internal/security/policy.go`. Field di kolom rencana adalah usulan. File kontrak tiap baris berada di `backend/services/finance_accounting/contract/proto/finance_accounting/v1/`.

| File proto | RPC | Rencana untuk Gateway | Kontrol sebelum dipakai |
| --- | --- | --- | --- |
| `accounting.proto` | `VerifyFunctionalCurrency` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `fixed_assets.proto` | `RegisterFixedAsset` | Mutation `registerFixedAsset` | FA-11 · register → aktif → depresiasi |
| `fixed_assets.proto` | `PostDepreciation` | Mutation `postDepreciation` | FA-11 · register → aktif → depresiasi |
| `budget.proto` | `ReleaseHRBudgetReservation` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `budget.proto` | `CommitHRBudgetReservation` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `budget.proto` | `VerifyHRBudgetDecision` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `budget.proto` | `VerifyWorkforceBudgetAuthorization` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `budget.proto` | `VerifyRequisitionBudgetAuthorization` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `accounts_receivable.proto` | `CreateARInvoice` | Mutation `createARInvoice` | FA-01/02/03 · scope dan posting pajak |
| `accounts_receivable.proto` | `ApproveARInvoice` | Mutation `approveARInvoice` | FA-01/02/03 · scope dan posting pajak |
| `accounts_receivable.proto` | `PostARInvoice` | Mutation `postARInvoice` | FA-01/02/03 · scope dan posting pajak |
| `accounts_receivable.proto` | `GetTradeOutstanding` | Query `tradeOutstanding` | FA-01/02/03 · scope dan posting pajak |
| `accounts_receivable.proto` | `RecordRetailTaxSnapshot` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `expenses.proto` | `CreateExpenseClaim` | Mutation `createExpenseClaim` | Instruksi 8 · bukti, reimbursement HR dan SoD |
| `expenses.proto` | `ApproveExpenseClaim` | Mutation `approveExpenseClaim` | Instruksi 8 · bukti, reimbursement HR dan SoD |
| `expenses.proto` | `PostExpenseClaim` | Mutation `postExpenseClaim` | Instruksi 8 · bukti, reimbursement HR dan SoD |
| `expenses.proto` | `ReverseExpenseClaim` | Mutation `reverseExpenseClaim` | Instruksi 8 · bukti, reimbursement HR dan SoD |
| `payroll.proto` | `PostPayrollAccrual` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `payroll.proto` | `VerifyOffboardingFinanceClearance` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `payroll.proto` | `VerifyPayrollPosting` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `payroll.proto` | `VerifyPayrollPublication` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `payroll.proto` | `VerifyPayrollPayment` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `payroll.proto` | `VerifyPayrollSettlement` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `payroll.proto` | `FindPayrollPublicationByRun` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `treasury.proto` | `ImportBankStatement` | Mutation `importBankStatement` | FA-08 · bukti bank, approval dan idempotensi |
| `treasury.proto` | `ScheduleSupplierPayout` | Mutation `scheduleSupplierPayout` | FA-08 · bukti bank, approval dan idempotensi |
| `treasury.proto` | `ApprovePayoutSchedule` | Mutation `approvePayoutSchedule` | FA-08 · bukti bank, approval dan idempotensi |
| `treasury.proto` | `DispatchScheduledPayout` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `inbox.proto` | `DeliverInboxEvent` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `inventory_cost.proto` | `ApplyInventoryCostEvent` | Privat antar-workload | Allowlist identitas workload, scope sumber dan replay aman |
| `inventory_cost.proto` | `GetInventoryBookBalance` | Query `inventoryBookBalance` | FA-06 · scope dan buku valuasi |

## Kontrak baru yang diusulkan

| Jenis | Operasi GraphQL / gRPC | Pemanggil dan scope | Data, alur dan prasyarat |
| --- | --- | --- | --- |
| Query | periodCloseReadiness / GetPeriodCloseReadiness | Controller · entity/period | Blocker, jumlah per sumber, cutoff, versi dan exception. FA-07. |
| Query | balanceSheet / GetBalanceSheet, cashFlow / GetCashFlow | Analis Finance · scope reporting | Saldo awal/akhir, klasifikasi dan drilldown terotorisasi. FA-01/07. |
| Query | bankReconciliations / ListBankReconciliations | Treasury · rekening dan entity | Connection dengan statement, residual, status dan tanggal. FA-08. |
| Mutation | matchBankStatementLine / MatchBankStatementLine, approveBankReconciliation / ApproveBankReconciliation | Treasury · maker/checker | Jumlah per pasangan, split, bukti jurnal, versi dan persetujuan independen. FA-08. |
| Query | supplierPayoutSchedule / GetSupplierPayoutSchedule, payoutReconciliationQueue / ListPayoutExceptions | Treasury · entity/instrumen | Jadwal, nominal, kode mata uang dan status teredaksi. FA-08. |
| Query | arInvoice / GetARInvoice, arInvoices / ListARInvoices | AR · entity/customer | Detail dan Connection invoice piutang, nomor, tanggal, jatuh tempo, outstanding. FA-02/03. |
| Query | payrollSettlementProgress / GetPayrollSettlementProgress | Finance payroll · entity/payroll group | Per hasil, alokasi posted, outstanding, bukti no-transfer dan reversal. FA-09. |
| Mutation | capitalizeFixedAsset / CapitalizeFixedAsset, placeFixedAssetInService / PlaceFixedAssetInService | Akuntan aset · entity/lokasi | Bukti acquisition/AP, tanggal, nilai dasar, kebijakan depresiasi. FA-11. |
| Query / Mutation | productionBatchCost / GetProductionBatchCost, postProductionCostAllocation / PostProductionCostAllocation | Cost accountant · entity/batch | Rekonsiliasi WIP/vendor/QC, selisih dan versi fakta. FA-12. |
| gRPC privat | CalculateSalesTax, RequestCustomerPayment, GetCustomerFundingStatus, ReserveFulfillmentFunding, ReleaseFulfillmentFunding, ConsumeFulfillmentFunding, RequestCustomerRefund | Workload Commerce/fulfillment/Finance sesuai perintah | Kunci per order, bagian fulfillment, key, hash, jumlah dan mata uang. Status untuk pembeli lewat operasi pemilik Commerce. FA-10. |

Gunakan pola SDL dari `backend/gateway/graphql/schema.graphql` dan SDL Finance aktual. Untuk daftar objek pakai Connection, edges, nodes, pageInfo dan cursor yang stabil. Tampilkan nomor dokumen, nama pihak yang diizinkan, nominal, mata uang, tanggal dan status. ID opaque dipakai aplikasi sebagai referensi, sedangkan pengguna melihat label bisnis. Catalog IAM, pemetaan field sensitif, resolver Node serta filter total mengikuti scope objek yang sama. [Instruksi agent](../../../agents/instructions/finance_accounting/9-leader-remediasi-logika-keuangan-dan-integrasi-instructions.md) · [Bukti audit](../../../agents/audit/finance_accounting/10-2026-11-29-010010-audit-logika-keuangan-dan-integrasi.md)
