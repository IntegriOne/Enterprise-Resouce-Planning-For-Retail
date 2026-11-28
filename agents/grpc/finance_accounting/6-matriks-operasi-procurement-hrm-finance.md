# List operasi Procurement, HRM dan Finance

Snapshot kode produksi pada 28 November 2026 WIB. Perhitungan memakai definisi dan ekstensi root Query/Mutation per direktori SDL Gateway. Setiap field dihitung sekali. Root bersama `node` berada di luar hitungan. Status “ada” menyatakan keberadaan field di source. Kesiapan runtime memerlukan bukti tersendiri.

| Management | Query SDL | Mutation SDL | Total | Selisih terhadap rencana |
| --- | ---: | ---: | ---: | --- |
| Procurement | 46 | 56 | 102 | 46/46 Query dan 55/66 Mutation punya padanan. Ada 1 Mutation tambahan di SDL. |
| HRM | 176 | 326 | 502 | 324 nama Mutation langsung cocok, 1 pemetaan ledger cuti, 1 RPC penautan privat IAM. Ada 1 Mutation pengajuan akun di SDL. |
| Finance | 11 | 33 | 44 | Dari tabel rancangan 12 Query dan 41 Mutation, 10 Query dan 29 Mutation punya padanan. SDL memuat tambahan 1 Query dan 4 Mutation. |
| **Jumlah di SDL** | **233** | **415** | **648** | **25 calon field katalog belum tampil** di Procurement dan Finance. |

## Penjelasan operasi dan butuh GraphQL

Tabel berikut mencantumkan semua baris rancangan publik dan field tambahan di SDL. Nama calon field mengikuti katalog owner. Gunakan ejaan pada SDL saat memanggil API. Posisi sumber menunjuk deklarasi aktual. Perubahan yang diusulkan setelah audit dipisahkan di bagian akhir. [Audit kode inti](../../../agents/audit/finance_accounting/9-2026-11-28-213747-audit-kode-inti-procurement-hrm-finance.md)

## Procurement

Rancangan sumber di [agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md).

| Jenis | Calon field | RPC backend | Field SDL dan lokasi |
| --- | --- | --- | --- |
| Query | `procurementSettings` | `GetProcurementSettings` | [procurement_settings](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L366) |
| Query | `procurementApprovalPolicies` | `ListProcurementApprovalPolicies` | [procurement_approval_policies](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L367) |
| Query | `procurementPartner` | `GetProcurementPartner` | [procurement_partner](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L368) |
| Query | `procurementPartners` | `ListProcurementPartners` | [procurement_partners](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L369) |
| Query | `partnerOffers` | `ListPartnerOffers` | [procurement_partner_offers](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L370) |
| Query | `purchaseRequest` | `GetPurchaseRequest` | [purchase_request](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L371) |
| Query | `purchaseRequests` | `ListPurchaseRequests` | [purchase_requests](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L372) |
| Query | `purchaseRequestOutstanding` | `GetPurchaseRequestOutstanding` | [purchase_request_outstanding](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L373) |
| Query | `procurementApprovals` | `ListProcurementApprovals` | [procurement_approvals](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L374) |
| Query | `pendingProcurementApprovals` | `ListPendingProcurementApprovals` | [pending_procurement_approvals](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L375) |
| Query | `rfq` | `GetRFQ` | [procurement_rfq](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L376) |
| Query | `rfqs` | `ListRFQs` | [procurement_rfqs](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L377) |
| Query | `rfqInvitations` | `ListRFQInvitations` | [procurement_rfq_invitations](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L378) |
| Query | `supplierQuotation` | `GetSupplierQuotation` | [supplier_quotation](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L379) |
| Query | `supplierQuotations` | `ListSupplierQuotations` | [supplier_quotations](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L380) |
| Query | `rfqQuotationComparison` | `CompareRFQQuotations` | [compare_rfq_quotations](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L381) |
| Query | `sourcingAwards` | `ListSourcingAwards` | [sourcing_awards](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L382) |
| Query | `purchaseOrder` | `GetPurchaseOrder` | [purchase_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L347) |
| Query | `purchaseOrders` | `ListPurchaseOrders` | [purchase_orders](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L348) |
| Query | `purchaseOrderRevisions` | `ListPurchaseOrderRevisions` | [purchase_order_revisions](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L349) |
| Query | `purchaseOrderResponses` | `ListPurchaseOrderResponses` | [purchase_order_responses](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L350) |
| Query | `purchaseOrderOutstanding` | `GetPurchaseOrderOutstanding` | [purchase_order_outstanding](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L351) |
| Query | `purchaseReceiptLinks` | `ListPurchaseReceiptLinks` | [purchase_receipt_links](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L353) |
| Query | `serviceAcceptance` | `GetServiceAcceptance` | [service_acceptance](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L354) |
| Query | `serviceAcceptances` | `ListServiceAcceptances` | [service_acceptances](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L355) |
| Query | `purchaseReturn` | `GetPurchaseReturn` | [purchase_return](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L356) |
| Query | `purchaseReturns` | `ListPurchaseReturns` | [purchase_returns](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L357) |
| Query | `purchaseOrderFinancialStatus` | `GetPurchaseOrderFinancialStatus` | [purchase_order_financial_status](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L352) |
| Query | `productionBatch` | `GetProductionBatch` | [production_batch](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L202) |
| Query | `productionBatches` | `ListProductionBatches` | [production_batches](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L203) |
| Query | `productionStages` | `ListProductionStages` | [production_stages](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L204) |
| Query | `vendorWorkOrder` | `GetVendorWorkOrder` | [vendor_work_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L205) |
| Query | `vendorWorkOrders` | `ListVendorWorkOrders` | [vendor_work_orders](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L206) |
| Query | `vendorWorkOrderHandoffs` | `ListVendorWorkOrderHandoffs` | [vendor_work_order_handoffs](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L207) |
| Query | `vendorWorkOrderCharges` | `ListVendorWorkOrderCharges` | [vendor_work_order_charges](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L208) |
| Query | `vendorProgressReports` | `ListVendorProgressReports` | [vendor_progress_reports](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L209) |
| Query | `vendorMaterialReconciliations` | `ListVendorMaterialReconciliations` | [vendor_material_reconciliations](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L210) |
| Query | `vendorQualityActions` | `ListVendorQualityActions` | [vendor_quality_actions](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L211) |
| Query | `materialSubstitutionRequests` | `ListMaterialSubstitutionRequests` | [material_substitution_requests](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L212) |
| Query | `procurementCommitments` | `GetProcurementCommitments` | [procurement_commitments](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L358) |
| Query | `actualProcurementSpending` | `GetActualProcurementSpending` | [actual_procurement_spending](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L359) |
| Query | `overduePurchaseOrders` | `ListOverduePurchaseOrders` | [overdue_purchase_orders](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L360) |
| Query | `supplierPriceHistory` | `ListSupplierPriceHistory` | [supplier_price_history](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L361) |
| Query | `supplierPerformance` | `GetSupplierPerformance` | [supplier_performance](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L362) |
| Query | `materialAtVendor` | `GetMaterialAtVendor` | [material_at_vendor](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L363) |
| Query | `vendorMaterialVariance` | `GetVendorMaterialVariance` | [vendor_material_variance](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L364) |
| Mutation | `configureProcurement` | `ConfigureProcurement` | [configure_procurement](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L386) |
| Mutation | `createProcurementApprovalPolicy` | `CreateProcurementApprovalPolicy` | [create_procurement_approval_policy](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L387) |
| Mutation | `activateProcurementApprovalPolicy` | `ActivateProcurementApprovalPolicy` | [activate_procurement_approval_policy](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L388) |
| Mutation | `retireProcurementApprovalPolicy` | `RetireProcurementApprovalPolicy` | [retire_procurement_approval_policy](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L389) |
| Mutation | `createProcurementPartnerProfile` | `CreateProcurementPartnerProfile` | [create_procurement_partner_profile](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L390) |
| Mutation | `qualifyPartner` | `QualifyPartner` | [qualify_procurement_partner](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L391) |
| Mutation | `suspendPartner` | `SuspendPartner` | [suspend_procurement_partner](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L392) |
| Mutation | `disqualifyPartner` | `DisqualifyPartner` | [disqualify_procurement_partner](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L393) |
| Mutation | `createPartnerOffer` | `CreatePartnerOffer` | [create_partner_offer](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L395) |
| Mutation | `revisePartnerOffer` | `RevisePartnerOffer` | Belum ada field |
| Mutation | `retirePartnerOffer` | `RetirePartnerOffer` | [retire_partner_offer](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L396) |
| Mutation | `createPurchaseRequest` | `CreatePurchaseRequest` | [create_purchase_request](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L397) |
| Mutation | `updatePurchaseRequestDraft` | `UpdatePurchaseRequestDraft` | Belum ada field |
| Mutation | `submitPurchaseRequest` | `SubmitPurchaseRequest` | [submit_purchase_request](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L398) |
| Mutation | `cancelPurchaseRequest` | `CancelPurchaseRequest` | [cancel_purchase_request](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L399) |
| Mutation | `closePurchaseRequest` | `ClosePurchaseRequest` | [close_purchase_request](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L400) |
| Mutation | `decideProcurementApproval` | `DecideProcurementApproval` | [decide_procurement_approval](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L401) |
| Mutation | `createRFQ` | `CreateRFQ` | [create_procurement_rfq](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L402) |
| Mutation | `updateRFQDraft` | `UpdateRFQDraft` | Belum ada field |
| Mutation | `inviteRFQPartner` | `InviteRFQPartner` | [invite_rfq_partner](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L406) |
| Mutation | `openRFQ` | `PublishRFQ` | [publish_procurement_rfq](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L403) |
| Mutation | `beginRFQEvaluation` | `BeginRFQEvaluation` (perlu ditambah) | Belum ada field |
| Mutation | `closeRFQ` | `CloseRFQ` | [close_procurement_rfq](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L404) |
| Mutation | `cancelRFQ` | `CancelRFQ` | [cancel_procurement_rfq](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L405) |
| Mutation | `createSupplierQuotationDraft` | `CreateSupplierQuotation` | [create_supplier_quotation](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L407) |
| Mutation | `submitSupplierQuotation` | `SubmitSupplierQuotation` | [submit_supplier_quotation](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L408) |
| Mutation | `withdrawSupplierQuotation` | `WithdrawSupplierQuotation` | [withdraw_supplier_quotation](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L409) |
| Mutation | `createSourcingAward` | `CreateSourcingAward` | [create_sourcing_award](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L410) |
| Mutation | `submitSourcingAward` | `SubmitSourcingAward` | [submit_sourcing_award](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L411) |
| Mutation | `createPurchaseOrder` | `CreatePurchaseOrder` | [create_purchase_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L368) |
| Mutation | `createPurchaseOrderRevision` | `CreatePurchaseOrderRevision` | [create_purchase_order_revision](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L369) |
| Mutation | `updatePurchaseOrderRevisionDraft` | `UpdatePurchaseOrderRevisionDraft` | Belum ada field |
| Mutation | `submitPurchaseOrderRevision` | `SubmitPurchaseOrderRevision` | [submit_purchase_order_revision](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L370) |
| Mutation | `issuePurchaseOrderRevision` | `IssuePurchaseOrderRevision` | [issue_purchase_order_revision](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L371) |
| Mutation | `recordPurchaseOrderResponse` | `RecordPurchaseOrderResponse` | [record_purchase_order_response](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L372) |
| Mutation | `cancelPurchaseOrder` | `CancelPurchaseOrder` | [cancel_purchase_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L373) |
| Mutation | `closePurchaseOrder` | `ClosePurchaseOrder` | [close_purchase_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L374) |
| Mutation | `createServiceAcceptance` | `CreateServiceAcceptance` | [create_service_acceptance](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L375) |
| Mutation | `submitServiceAcceptance` | `SubmitServiceAcceptance` | [submit_service_acceptance](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L376) |
| Mutation | `reverseServiceAcceptance` | `ReverseServiceAcceptance` | [reverse_service_acceptance](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L377) |
| Mutation | `createPurchaseReturn` | `CreatePurchaseReturn` | [create_purchase_return](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L378) |
| Mutation | `submitPurchaseReturn` | `SubmitPurchaseReturn` | [submit_purchase_return](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L379) |
| Mutation | `cancelPurchaseReturn` | `CancelPurchaseReturn` | [cancel_purchase_return](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L380) |
| Mutation | `closePurchaseReturn` | `ClosePurchaseReturn` | [close_purchase_return](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_orders.graphql#L381) |
| Mutation | `createProductionBatch` | `CreateProductionBatch` | [create_production_batch](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L216) |
| Mutation | `updateProductionBatchDraft` | `UpdateProductionBatchDraft` | Belum ada field |
| Mutation | `addProductionStage` | `AddProductionStage` | [add_production_stage](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L217) |
| Mutation | `addProductionStageDependency` | `AddProductionStageDependency` | [add_production_stage_dependency](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L218) |
| Mutation | `releaseProductionBatch` | `ReleaseProductionBatch` | [release_production_batch](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L219) |
| Mutation | `closeProductionBatch` | `CloseProductionBatch` | [close_production_batch](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L220) |
| Mutation | `createVendorWorkOrder` | `CreateVendorWorkOrder` | [create_vendor_work_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L221) |
| Mutation | `updateVendorWorkOrderDraft` | `UpdateVendorWorkOrderDraft` | Belum ada field |
| Mutation | `releaseVendorWorkOrder` | `ReleaseVendorWorkOrder` | [release_vendor_work_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L222) |
| Mutation | `replaceVendorWorkOrder` | `ReplaceVendorWorkOrder` | Belum ada field |
| Mutation | `closeVendorWorkOrder` | `CloseVendorWorkOrder` | [close_vendor_work_order](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L223) |
| Mutation | `allocateVendorHandoff` | `AllocateVendorHandoff` | [allocate_vendor_handoff](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L224) |
| Mutation | `configureVendorWorkOrderCharges` | `ConfigureVendorWorkOrderCharges` | [configure_vendor_work_order_charge](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L225) |
| Mutation | `submitVendorProgress` | `SubmitVendorProgress` | [submit_vendor_progress](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L226) |
| Mutation | `submitVendorMaterialReconciliation` | `SubmitVendorMaterialReconciliation` | Belum ada field |
| Mutation | `approveVendorMaterialReconciliation` | `ApproveVendorMaterialReconciliation` | [approve_vendor_material_reconciliation](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L227) |
| Mutation | `createVendorQualityAction` | `CreateVendorQualityAction` | Belum ada field |
| Mutation | `approveVendorQualityAction` | `ApproveVendorQualityAction` | [approve_vendor_quality_action](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L228) |
| Mutation | `resolveVendorQualityAction` | `ResolveVendorQualityAction` | Belum ada field |
| Mutation | `requestMaterialSubstitution` | `RequestMaterialSubstitution` | [request_material_substitution](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L229) |
| Mutation | `approveMaterialSubstitution` | `ApproveMaterialSubstitution` | [approve_material_substitution](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L230) |
| Mutation | `applyMaterialSubstitution` | `ApplyMaterialSubstitution` | [apply_material_substitution](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_external.graphql#L231) |
| Mutation | Tambahan di SDL | Lihat resolver owner | [request_partner_account_provisioning](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L394) |

## Finance

Rancangan sumber di [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md).

| Jenis | Calon field | RPC backend | Field SDL dan lokasi |
| --- | --- | --- | --- |
| Mutation | `configureAccounting` | `ConfigureAccounting` | [configureAccounting](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L275) |
| Query | `accountingSettings` | `GetAccountingSettings` | [accountingSettings](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L267) |
| Mutation | `createFiscalYear` | `CreateFiscalYear` | [createFiscalYear](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L276) |
| Mutation | `createAccountingPeriod` | `CreateAccountingPeriod` | [createAccountingPeriod](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L277) |
| Mutation | `openAccountingPeriod` | `OpenAccountingPeriod` | [openAccountingPeriod](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L278) |
| Mutation | `closeAccountingPeriod` | `CloseAccountingPeriod` | [closeAccountingPeriod](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L279) |
| Mutation | `reopenAccountingPeriod` | `ReopenAccountingPeriod` | [reopenAccountingPeriod](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L280) |
| Mutation | `closeFiscalYear` | `CloseFiscalYear` | [closeFiscalYear](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L281) |
| Mutation | `createGLAccount` | `CreateGLAccount` | [createGLAccount](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L282) |
| Query | `glAccount` | `GetGLAccount` | [glAccount](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L268) |
| Query | `glAccounts` | `ListGLAccounts` | [glAccounts](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L269) |
| Mutation | `configurePostingRule` | `ConfigurePostingRule` | [configurePostingRule](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L283) |
| Mutation | `configurePayrollPostingRule` | `ConfigurePostingRule` | [configurePayrollPostingRule](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L284) |
| Mutation | `configureFinancePostingRule` | `ConfigurePostingRule` | [configureFinancePostingRule](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L285) |
| Mutation | `createJournalDraft` | `CreateJournalDraft` | [createJournalDraft](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L286) |
| Mutation | `approveJournal` | `ApproveJournal` | [approveJournal](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L287) |
| Mutation | `postJournal` | `PostJournal` | [postJournal](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L288) |
| Mutation | `reverseJournal` | `ReverseJournal` | [reverseJournal](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L289) |
| Query | `journalEntry` | `GetJournalEntry` | [journalEntry](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L270) |
| Query | `journalEntries` | `ListJournalEntries` | [journalEntries](../../../backend/gateway/graphql/finance_accounting/finance_accounting.graphql#L271) |
| Mutation | `createAPInvoice` | `CreateAPInvoice` | [createAPInvoice](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L191) |
| Query | `apInvoice` | `GetAPInvoice` | [apInvoice](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L133) |
| Query | `apInvoices` | `ListAPInvoices` | [apInvoices](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L134) |
| Mutation | `performThreeWayMatching` | `PerformThreeWayMatching` | [performThreeWayMatching](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L192) |
| Mutation | `approveAPInvoice` | `ApproveAPInvoice` | [approveAPInvoice](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L193) |
| Mutation | `postAPInvoice` | `PostAPInvoice` | [postAPInvoice](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L194) |
| Mutation | `createARInvoice` | `CreateARInvoice` | Belum ada field |
| Mutation | `approveARInvoice` | `ApproveARInvoice` | Belum ada field |
| Mutation | `postARInvoice` | `PostARInvoice` | Belum ada field |
| Query | `tradeOutstanding` | `GetTradeOutstanding` | Belum ada field |
| Mutation | `createBudgetDraft` | `CreateBudgetDraft` | [createBudgetDraft](../../../backend/gateway/graphql/finance_accounting/finance_budget.graphql#L34) |
| Mutation | `approveBudget` | `ApproveBudget` | [approveBudget](../../../backend/gateway/graphql/finance_accounting/finance_budget.graphql#L35) |
| Mutation | `createExpenseClaim` | `CreateExpenseClaim` | Belum ada field |
| Mutation | `approveExpenseClaim` | `ApproveExpenseClaim` | Belum ada field |
| Mutation | `postExpenseClaim` | `PostExpenseClaim` | Belum ada field |
| Mutation | `reverseExpenseClaim` | `ReverseExpenseClaim` | Belum ada field |
| Mutation | `registerFixedAsset` | `RegisterFixedAsset` | Belum ada field |
| Mutation | `postDepreciation` | `PostDepreciation` | Belum ada field |
| Query | `inventoryBookBalance` | `GetInventoryBookBalance` | Belum ada field |
| Mutation | `completePayrollSettlement` | `CompletePayrollSettlement` | [completePayrollSettlement](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql#L19) |
| Mutation | `reviewFinanceClearance` | `ReviewFinanceClearance` | [reviewFinanceClearance](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql#L17) |
| Query | `trialBalance` | `GetTrialBalance` | [trialBalance](../../../backend/gateway/graphql/finance_accounting/finance_reporting.graphql#L42) |
| Query | `profitAndLoss` | `GetProfitAndLoss` | [profitAndLoss](../../../backend/gateway/graphql/finance_accounting/finance_reporting.graphql#L43) |
| Query | `taxSummary` | `GetTaxSummary` | [taxSummary](../../../backend/gateway/graphql/finance_accounting/finance_reporting.graphql#L44) |
| Mutation | `createPaymentDraft` | `CreatePaymentDraft` | [createPaymentDraft](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L255) |
| Mutation | `approvePayment` | `ApprovePayment` | [approvePayment](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L256) |
| Mutation | `confirmPaymentEvidence` | `ConfirmPaymentEvidence` | [confirmPaymentEvidence](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L257) |
| Mutation | `postPayment` | `PostPayment` | [postPayment](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L258) |
| Mutation | `importBankStatement` | `ImportBankStatement` | Belum ada field |
| Mutation | `scheduleSupplierPayout` | `ScheduleSupplierPayout` | Belum ada field |
| Mutation | `approvePayoutSchedule` | `ApprovePayoutSchedule` | Belum ada field |
| Mutation | `issueBudgetAuthorization` | `IssueHRBudgetAuthorization` | [issueBudgetAuthorization](../../../backend/gateway/graphql/finance_accounting/finance_budget.graphql#L59) |
| Mutation | `revokeBudgetAuthorization` | `RevokeHRBudgetAuthorization` | [revokeBudgetAuthorization](../../../backend/gateway/graphql/finance_accounting/finance_budget.graphql#L60) |
| Mutation | Tambahan di SDL | Lihat resolver owner | [postServiceAcceptanceCreditNote](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L188) |
| Mutation | Tambahan di SDL | Lihat resolver owner | [approveServiceAcceptanceCreditNote](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L189) |
| Mutation | Tambahan di SDL | Lihat resolver owner | [createServiceAcceptanceCreditNote](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L190) |
| Query | Tambahan di SDL | Lihat resolver owner | [budget](../../../backend/gateway/graphql/finance_accounting/finance_budget.graphql#L74) |
| Mutation | Tambahan di SDL | Lihat resolver owner | [publishPayrollRun](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql#L18) |

## HRM

Pencocokan nama RPC pada 22 proto bisnis dengan tabel katalog per service menghasilkan 1.018 RPC, terdiri dari 674 kandidat Query, 326 kandidat Mutation dan 18 nonaktif. Finder dapat berbagi satu connection melalui filter. Snapshot budget requisition dan workforce melayani orkestrasi Finance. Tabel total inventaris telah memuat 1.018, sedangkan paragraf pembuka masih menyebut 1.015. Gunakan definisi proto dan tabel per service ketika memperbarui narasi. [Inventaris dan tautan 22 domain](../hrm/1-inventaris-operasi-grpc-hrm.md) · [Konvensi pemetaan](../hrm/36-penyelarasan-sdl-gateway.md)

`postLeaveAdjustment` memanggil `PostLedgerEntry`. `LinkWorkerPrincipal` berjalan privat dalam provisioning IAM melalui `createEmployeeUser`. `requestEmployeeAccount` adalah pengajuan tersendiri. Empat field slip pegawai sudah ada. Perbaikan pada publikasi dan pemulihan anggaran perlu dikerjakan pada jalur internal. [Ledger cuti](../../../backend/gateway/internal/adapter/human_resource/leave_posting.go#L47) · [Pengajuan akun](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L20) · [Provisioning](../hrm/33-alur-pengajuan-akun-employee.md) · [SDL slip](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L31)

| File SDL HRM | Jenis | Field aktual |
| --- | --- | --- |
| [hrm_access_request.graphql](../../../backend/gateway/graphql/human_resource/hrm_access_request.graphql#L22) | Mutation (1) | `requestEmployeeAccount` |
| [hrm_access_request.graphql](../../../backend/gateway/graphql/human_resource/hrm_access_request.graphql#L26) | Query (2) | `employeeAccountRequest`, `workerEmployeeAccountRequest` |
| [hrm_accessrequest_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_accessrequest_entities.graphql#L32) | Query (2) | `accessRequest`, `accessRequests` |
| [hrm_accessrequest_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_accessrequest_entities.graphql#L84) | Mutation (3) | `createAccessRequest`, `updateAccessRequest`, `deleteAccessRequest` |
| [hrm_analytics_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_analytics_entities.graphql#L28) | Query (4) | `humanCapitalMetricDefinition`, `humanCapitalMetricDefinitions`, `humanCapitalMetricSnapshot`, `humanCapitalMetricSnapshots` |
| [hrm_analytics_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_analytics_entities.graphql#L76) | Mutation (6) | `createHumanCapitalMetricDefinition`, `updateHumanCapitalMetricDefinition`, `deleteHumanCapitalMetricDefinition`, `createHumanCapitalMetricSnapshot`, `updateHumanCapitalMetricSnapshot`, `deleteHumanCapitalMetricSnapshot` |
| [hrm_benefits_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_benefits_commands.graphql#L16) | Mutation (2) | `approveBenefitEnrollment`, `endBenefitEnrollment` |
| [hrm_benefits_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_benefits_entities.graphql#L31) | Query (4) | `benefitPlan`, `benefitPlans`, `benefitEnrollment`, `benefitEnrollments` |
| [hrm_benefits_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_benefits_entities.graphql#L85) | Mutation (6) | `createBenefitPlan`, `updateBenefitPlan`, `deleteBenefitPlan`, `createBenefitEnrollment`, `updateBenefitEnrollment`, `deleteBenefitEnrollment` |
| [hrm_candidate.graphql](../../../backend/gateway/graphql/human_resource/hrm_candidate.graphql#L42) | Mutation (3) | `createCandidate`, `updateCandidate`, `deleteCandidate` |
| [hrm_candidate.graphql](../../../backend/gateway/graphql/human_resource/hrm_candidate.graphql#L60) | Query (2) | `candidate`, `candidates` |
| [hrm_career_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_career_commands.graphql#L28) | Mutation (4) | `completeDevelopmentPlan`, `approveDevelopmentPlan`, `completeDevelopmentAction`, `approveCareerPath` |
| [hrm_career_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_career_entities.graphql#L29) | Query (10) | `developmentPlan`, `developmentPlans`, `developmentAction`, `developmentActions`, `careerPath`, `careerPaths`, `careerPathTransition`, `careerPathTransitions`, `workerCareerAspiration`, `workerCareerAspirations` |
| [hrm_career_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_career_entities.graphql#L79) | Mutation (15) | `createDevelopmentPlan`, `updateDevelopmentPlan`, `deleteDevelopmentPlan`, `createDevelopmentAction`, `updateDevelopmentAction`, `deleteDevelopmentAction`, `createCareerPath`, `updateCareerPath`, `deleteCareerPath`, `createCareerPathTransition`, `updateCareerPathTransition`, `deleteCareerPathTransition`, `createWorkerCareerAspiration`, `updateWorkerCareerAspiration`, `deleteWorkerCareerAspiration` |
| [hrm_compensation_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_compensation_commands.graphql#L10) | Mutation (1) | `approveCompensationRecord` |
| [hrm_compensation_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_compensation_entities.graphql#L31) | Query (4) | `compensationBand`, `compensationBands`, `compensationRecord`, `compensationRecords` |
| [hrm_compensation_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_compensation_entities.graphql#L85) | Mutation (6) | `createCompensationBand`, `updateCompensationBand`, `deleteCompensationBand`, `createCompensationRecord`, `updateCompensationRecord`, `deleteCompensationRecord` |
| [hrm_competency_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_competency_entities.graphql#L22) | Query (8) | `workerQualification`, `workerQualifications`, `competency`, `competencies`, `jobCompetency`, `jobCompetencies`, `workerCompetencyAssessment`, `workerCompetencyAssessments` |
| [hrm_competency_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_competency_entities.graphql#L60) | Mutation (12) | `createWorkerQualification`, `updateWorkerQualification`, `deleteWorkerQualification`, `createCompetency`, `updateCompetency`, `deleteCompetency`, `createJobCompetency`, `updateJobCompetency`, `deleteJobCompetency`, `createWorkerCompetencyAssessment`, `updateWorkerCompetencyAssessment`, `deleteWorkerCompetencyAssessment` |
| [hrm_employeerelations_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_employeerelations_entities.graphql#L33) | Query (6) | `employeeRelationsCase`, `employeeRelationsCases`, `employeeRelationsEvent`, `employeeRelationsEvents`, `collectiveAgreement`, `collectiveAgreements` |
| [hrm_employeerelations_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_employeerelations_entities.graphql#L89) | Mutation (9) | `createEmployeeRelationsCase`, `updateEmployeeRelationsCase`, `deleteEmployeeRelationsCase`, `createEmployeeRelationsEvent`, `updateEmployeeRelationsEvent`, `deleteEmployeeRelationsEvent`, `createCollectiveAgreement`, `updateCollectiveAgreement`, `deleteCollectiveAgreement` |
| [hrm_engagement_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_engagement_entities.graphql#L28) | Query (4) | `engagementSurvey`, `engagementSurveys`, `engagementSurveyResult`, `engagementSurveyResults` |
| [hrm_engagement_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_engagement_entities.graphql#L76) | Mutation (3) | `createEngagementSurvey`, `updateEngagementSurvey`, `deleteEngagementSurvey` |
| [hrm_hrservice_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_hrservice_commands.graphql#L28) | Mutation (4) | `resolveServiceRequest`, `closeServiceRequest`, `cancelServiceRequest`, `assignServiceRequest` |
| [hrm_hrservice_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_hrservice_entities.graphql#L31) | Query (4) | `serviceRequest`, `serviceRequests`, `serviceRequestEvent`, `serviceRequestEvents` |
| [hrm_hrservice_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_hrservice_entities.graphql#L81) | Mutation (6) | `createServiceRequest`, `updateServiceRequest`, `deleteServiceRequest`, `createServiceRequestEvent`, `updateServiceRequestEvent`, `deleteServiceRequestEvent` |
| [hrm_job.graphql](../../../backend/gateway/graphql/human_resource/hrm_job.graphql#L27) | Query (5) | `job`, `jobsByIds`, `jobs`, `jobFamily`, `jobFamilies` |
| [hrm_job.graphql](../../../backend/gateway/graphql/human_resource/hrm_job.graphql#L92) | Mutation (2) | `createJobFamily`, `createJob` |
| [hrm_learning_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_learning_commands.graphql#L22) | Mutation (3) | `startLearningEnrollment`, `completeLearningEnrollment`, `withdrawLearningEnrollment` |
| [hrm_learning_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_learning_entities.graphql#L31) | Query (6) | `learningCourse`, `learningCourses`, `learningEnrollment`, `learningEnrollments`, `learningCourseCompetency`, `learningCourseCompetencies` |
| [hrm_learning_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_learning_entities.graphql#L85) | Mutation (9) | `createLearningCourse`, `updateLearningCourse`, `deleteLearningCourse`, `createLearningEnrollment`, `updateLearningEnrollment`, `deleteLearningEnrollment`, `createLearningCourseCompetency`, `updateLearningCourseCompetency`, `deleteLearningCourseCompetency` |
| [hrm_leave_crud_extra.graphql](../../../backend/gateway/graphql/human_resource/hrm_leave_crud_extra.graphql#L128) | Mutation (8) | `createLeavePolicy`, `updateLeavePolicy`, `deleteLeavePolicy`, `createLeaveEntitlement`, `updateLeaveEntitlement`, `deleteLeaveEntitlement`, `updateLeaveAdjustment`, `deleteLeaveAdjustment` |
| [hrm_leave_policy.graphql](../../../backend/gateway/graphql/human_resource/hrm_leave_policy.graphql#L33) | Query (8) | `leavePolicy`, `leavePolicies`, `leaveEntitlement`, `leaveEntitlements`, `leaveLedgerEntry`, `leaveLedgerEntries`, `leaveAdjustment`, `leaveAdjustments` |
| [hrm_leave_policy.graphql](../../../backend/gateway/graphql/human_resource/hrm_leave_policy.graphql#L146) | Mutation (5) | `approveLeaveAdjustment`, `rejectLeaveAdjustment`, `createLeaveAdjustment`, `closeLeaveEntitlement`, `postLeaveAdjustment` |
| [hrm_offboarding.graphql](../../../backend/gateway/graphql/human_resource/hrm_offboarding.graphql#L24) | Query (4) | `offboardingPlan`, `offboardingPlans`, `offboardingTask`, `offboardingTasks` |
| [hrm_offboarding.graphql](../../../backend/gateway/graphql/human_resource/hrm_offboarding.graphql#L81) | Mutation (8) | `createOffboardingPlan`, `createOffboardingTask`, `completeOffboardingTask`, `waiveOffboardingTask`, `approveOffboardingPlan`, `startOffboardingPlan`, `cancelOffboardingPlan`, `completeOffboardingPlan` |
| [hrm_offboarding_crud_extra.graphql](../../../backend/gateway/graphql/human_resource/hrm_offboarding_crud_extra.graphql#L59) | Mutation (4) | `updateOffboardingPlan`, `deleteOffboardingPlan`, `updateOffboardingTask`, `deleteOffboardingTask` |
| [hrm_onboarding_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_onboarding_commands.graphql#L22) | Mutation (3) | `startOnboardingPlan`, `completeOnboardingPlan`, `completeOnboardingTask` |
| [hrm_onboarding_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_onboarding_entities.graphql#L31) | Query (4) | `onboardingPlan`, `onboardingPlans`, `onboardingTask`, `onboardingTasks` |
| [hrm_onboarding_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_onboarding_entities.graphql#L75) | Mutation (6) | `createOnboardingPlan`, `updateOnboardingPlan`, `deleteOnboardingPlan`, `createOnboardingTask`, `updateOnboardingTask`, `deleteOnboardingTask` |
| [hrm_payroll_crud_extra.graphql](../../../backend/gateway/graphql/human_resource/hrm_payroll_crud_extra.graphql#L173) | Mutation (11) | `createPayrollPeriod`, `updatePayrollPeriod`, `deletePayrollPeriod`, `createPayrollGroup`, `updatePayrollGroup`, `deletePayrollGroup`, `createPayrollRuleSet`, `updatePayrollRuleSet`, `deletePayrollRuleSet`, `updatePayrollRun`, `deletePayrollRun` |
| [hrm_payroll_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_payroll_entities.graphql#L32) | Query (6) | `payrollProfile`, `payrollProfiles`, `payComponent`, `payComponents`, `payrollAdjustment`, `payrollAdjustments` |
| [hrm_payroll_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_payroll_entities.graphql#L88) | Mutation (9) | `createPayrollProfile`, `updatePayrollProfile`, `deletePayrollProfile`, `createPayComponent`, `updatePayComponent`, `deletePayComponent`, `createPayrollAdjustment`, `updatePayrollAdjustment`, `deletePayrollAdjustment` |
| [hrm_payroll_posting.graphql](../../../backend/gateway/graphql/human_resource/hrm_payroll_posting.graphql#L15) | Mutation (7) | `requestPayrollPosting`, `calculatePayrollRun`, `submitPayrollRunForReview`, `approvePayrollRun`, `cancelPayrollRun`, `freezePayrollInput`, `createPayrollRun` |
| [hrm_payroll_posting.graphql](../../../backend/gateway/graphql/human_resource/hrm_payroll_posting.graphql#L45) | Query (8) | `payrollRun`, `payrollRuns`, `payrollPeriod`, `payrollPeriods`, `payrollGroup`, `payrollGroups`, `payrollRuleSet`, `payrollRuleSets` |
| [hrm_payroll_result.graphql](../../../backend/gateway/graphql/human_resource/hrm_payroll_result.graphql#L29) | Query (4) | `payrollResult`, `payrollResults`, `payrollResultLine`, `payrollResultLines` |
| [hrm_payslip.graphql](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L32) | Query (3) | `myPayslip`, `myPayslipPdf`, `myPayslips` |
| [hrm_payslip.graphql](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L44) | Mutation (1) | `reissueEmployeePayslip` |
| [hrm_performance_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_performance_commands.graphql#L46) | Mutation (7) | `launchPerformanceCycle`, `closePerformanceCycle`, `submitPerformanceGoal`, `approvePerformanceGoal`, `rejectPerformanceGoal`, `submitPerformanceReview`, `finalizePerformanceReview` |
| [hrm_performance_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_performance_entities.graphql#L31) | Query (10) | `performanceCycle`, `performanceCycles`, `performanceGoal`, `performanceGoals`, `performanceReview`, `performanceReviews`, `performanceGoalUpdate`, `performanceGoalUpdates`, `performanceReviewGoalScore`, `performanceReviewGoalScores` |
| [hrm_performance_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_performance_entities.graphql#L85) | Mutation (15) | `createPerformanceCycle`, `updatePerformanceCycle`, `deletePerformanceCycle`, `createPerformanceGoal`, `updatePerformanceGoal`, `deletePerformanceGoal`, `createPerformanceReview`, `updatePerformanceReview`, `deletePerformanceReview`, `createPerformanceGoalUpdate`, `updatePerformanceGoalUpdate`, `deletePerformanceGoalUpdate`, `createPerformanceReviewGoalScore`, `updatePerformanceReviewGoalScore`, `deletePerformanceReviewGoalScore` |
| [hrm_personnel_crud_extra.graphql](../../../backend/gateway/graphql/human_resource/hrm_personnel_crud_extra.graphql#L36) | Mutation (3) | `deleteWorker`, `updateWorkerAssignment`, `deleteWorkerAssignment` |
| [hrm_personnel_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_personnel_entities.graphql#L29) | Query (6) | `workerLifecycleEvent`, `workerLifecycleEvents`, `workerDocument`, `workerDocuments`, `emergencyContact`, `emergencyContacts` |
| [hrm_personnel_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_personnel_entities.graphql#L125) | Mutation (6) | `createWorkerDocument`, `updateWorkerDocument`, `deleteWorkerDocument`, `createEmergencyContact`, `updateEmergencyContact`, `deleteEmergencyContact` |
| [hrm_policy_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_policy_commands.graphql#L10) | Mutation (1) | `acknowledgePolicy` |
| [hrm_policy_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_policy_entities.graphql#L30) | Query (4) | `policy`, `policies`, `policyAcknowledgement`, `policyAcknowledgements` |
| [hrm_policy_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_policy_entities.graphql#L82) | Mutation (6) | `createPolicy`, `updatePolicy`, `deletePolicy`, `createPolicyAcknowledgement`, `updatePolicyAcknowledgement`, `deletePolicyAcknowledgement` |
| [hrm_position.graphql](../../../backend/gateway/graphql/human_resource/hrm_position.graphql#L30) | Query (2) | `position`, `positions` |
| [hrm_position.graphql](../../../backend/gateway/graphql/human_resource/hrm_position.graphql#L60) | Mutation (4) | `activatePosition`, `createPosition`, `updatePosition`, `deletePosition` |
| [hrm_promotion_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_promotion_commands.graphql#L40) | Mutation (5) | `submitPromotionRecommendation`, `approvePromotionRecommendation`, `rejectPromotionRecommendation`, `returnPromotionRecommendation`, `applyPromotionRecommendation` |
| [hrm_promotion_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_promotion_entities.graphql#L36) | Query (4) | `promotionRecommendation`, `promotionRecommendations`, `promotionDecision`, `promotionDecisions` |
| [hrm_promotion_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_promotion_entities.graphql#L96) | Mutation (6) | `createPromotionRecommendation`, `updatePromotionRecommendation`, `deletePromotionRecommendation`, `createPromotionDecision`, `updatePromotionDecision`, `deletePromotionDecision` |
| [hrm_recruitment_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_recruitment_commands.graphql#L107) | Mutation (15) | `startApplicationScreening`, `shortlistApplication`, `rejectApplication`, `startInterview`, `completeInterview`, `cancelInterview`, `approveOffer`, `issueOffer`, `acceptOffer`, `declineOffer`, `withdrawOffer`, `publishJobPosting`, `closeJobPosting`, `withdrawJobPosting`, `completeApplicationAssessment` |
| [hrm_recruitment_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_recruitment_entities.graphql#L34) | Query (16) | `application`, `applications`, `applicationStageHistory`, `applicationStageHistories`, `interview`, `interviews`, `interviewPanelist`, `interviewPanelists`, `interviewFeedback`, `interviewFeedbacks`, `offer`, `offers`, `jobPosting`, `jobPostings`, `applicationAssessment`, `applicationAssessments` |
| [hrm_recruitment_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_recruitment_entities.graphql#L92) | Mutation (21) | `createApplication`, `updateApplication`, `deleteApplication`, `createInterview`, `updateInterview`, `deleteInterview`, `createInterviewPanelist`, `updateInterviewPanelist`, `deleteInterviewPanelist`, `createInterviewFeedback`, `updateInterviewFeedback`, `deleteInterviewFeedback`, `createOffer`, `updateOffer`, `deleteOffer`, `createJobPosting`, `updateJobPosting`, `deleteJobPosting`, `createApplicationAssessment`, `updateApplicationAssessment`, `deleteApplicationAssessment` |
| [hrm_requisition.graphql](../../../backend/gateway/graphql/human_resource/hrm_requisition.graphql#L33) | Query (2) | `requisition`, `requisitions` |
| [hrm_requisition.graphql](../../../backend/gateway/graphql/human_resource/hrm_requisition.graphql#L115) | Mutation (8) | `closeRequisition`, `approveRequisition`, `openRequisition`, `updateRequisition`, `deleteRequisition`, `rejectRequisition`, `submitRequisition`, `createRequisition` |
| [hrm_succession_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_succession_commands.graphql#L10) | Mutation (1) | `approveSuccessionPlan` |
| [hrm_succession_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_succession_entities.graphql#L30) | Query (4) | `successionPlan`, `successionPlans`, `successionCandidate`, `successionCandidates` |
| [hrm_succession_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_succession_entities.graphql#L76) | Mutation (6) | `createSuccessionPlan`, `updateSuccessionPlan`, `deleteSuccessionPlan`, `createSuccessionCandidate`, `updateSuccessionCandidate`, `deleteSuccessionCandidate` |
| [hrm_talent_commands.graphql](../../../backend/gateway/graphql/human_resource/hrm_talent_commands.graphql#L79) | Mutation (11) | `launchTalentReviewCycle`, `startTalentReviewCycleCalibration`, `closeTalentReviewCycle`, `submitTalentAssessment`, `withdrawTalentAssessment`, `finalizeTalentAssessment`, `openTalentCalibrationSession`, `recordTalentCalibrationDecision`, `approveTalentPoolMembership`, `rejectTalentPoolMembership`, `exitTalentPoolMembership` |
| [hrm_talent_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_talent_entities.graphql#L28) | Query (18) | `talentFramework`, `talentFrameworks`, `talentDimension`, `talentDimensions`, `talentReviewCycle`, `talentReviewCycles`, `talentAssessment`, `talentAssessments`, `talentAssessmentScore`, `talentAssessmentScores`, `talentCalibrationSession`, `talentCalibrationSessions`, `talentCalibrationDecision`, `talentCalibrationDecisions`, `talentPool`, `talentPools`, `talentPoolMembership`, `talentPoolMemberships` |
| [hrm_talent_entities.graphql](../../../backend/gateway/graphql/human_resource/hrm_talent_entities.graphql#L76) | Mutation (27) | `createTalentFramework`, `updateTalentFramework`, `deleteTalentFramework`, `createTalentDimension`, `updateTalentDimension`, `deleteTalentDimension`, `createTalentReviewCycle`, `updateTalentReviewCycle`, `deleteTalentReviewCycle`, `createTalentAssessment`, `updateTalentAssessment`, `deleteTalentAssessment`, `createTalentAssessmentScore`, `updateTalentAssessmentScore`, `deleteTalentAssessmentScore`, `createTalentCalibrationSession`, `updateTalentCalibrationSession`, `deleteTalentCalibrationSession`, `createTalentCalibrationDecision`, `updateTalentCalibrationDecision`, `deleteTalentCalibrationDecision`, `createTalentPool`, `updateTalentPool`, `deleteTalentPool`, `createTalentPoolMembership`, `updateTalentPoolMembership`, `deleteTalentPoolMembership` |
| [hrm_worker.graphql](../../../backend/gateway/graphql/human_resource/hrm_worker.graphql#L27) | Query (8) | `worker`, `workers`, `workerAssignmentTransfers`, `workerAssignmentTransfer`, `workerAssignment`, `workerAssignments`, `employmentContract`, `employmentContracts` |
| [hrm_worker.graphql](../../../backend/gateway/graphql/human_resource/hrm_worker.graphql#L77) | Mutation (13) | `createWorker`, `updateWorker`, `activateWorker`, `terminateWorker`, `applyPrimaryAssignment`, `approveWorkerAssignment`, `createWorkerAssignment`, `submitWorkerAssignmentTransfer`, `decideWorkerAssignmentTransfer`, `createEmploymentContract`, `recordSignedEmploymentContract`, `deleteEmploymentContract`, `updateEmploymentContract` |
| [hrm_workforce_plan.graphql](../../../backend/gateway/graphql/human_resource/hrm_workforce_plan.graphql#L28) | Query (4) | `workforcePlan`, `workforcePlans`, `workforcePlanLine`, `workforcePlanLines` |
| [hrm_workforce_plan.graphql](../../../backend/gateway/graphql/human_resource/hrm_workforce_plan.graphql#L98) | Mutation (10) | `submitWorkforcePlan`, `cancelWorkforcePlan`, `deleteWorkforcePlan`, `approveWorkforcePlan`, `supersedeWorkforcePlan`, `createWorkforcePlan`, `updateWorkforcePlan`, `updateWorkforcePlanLine`, `createWorkforcePlanLine`, `deleteWorkforcePlanLine` |
| [hrm_workforceplanning_crud_extra.graphql](../../../backend/gateway/graphql/human_resource/hrm_workforceplanning_crud_extra.graphql#L49) | Mutation (4) | `updateJobFamily`, `deleteJobFamily`, `updateJob`, `deleteJob` |

## Operasi tambahan yang dibutuhkan alur bisnis

| Owner | Operasi yang diusulkan | Data minimum dan otorisasi |
| --- | --- | --- |
| Finance Treasury | Perluas Mutation `postPayment` dengan alokasi payroll bertipe; pertahankan kontrak alokasi pemasok | Referensi item settlement, nominal decimal eksak, arah, worker dan badan hukum dari Finance. Permission posting, MFA serta bukti terkonfirmasi. |
| Finance Payroll | Query detail batch dan connection item settlement | ID opaque, periode, total, mata uang, nominal teralokasi, sisa dan status. Hanya lingkup payroll Finance berwenang. |
| Finance Expenses | Query detail/connection klaim biaya | Nomor klaim, jenis, tanggal, status, subtotal dan total. Cakupan badan hukum dan pembatasan data penerima. |
| Finance Treasury | Query detail/connection jadwal payout dan daftar kebijakan aktif | ID, kode invoice, waktu lokal/zona, status, versi policy dan jumlah. Data rekening tetap privat. |
| Procurement Sourcing | Evaluasi publikasi Mutation koreksi draft quotation pada fase pengguna yang memiliki draft | RPC `UpdateSupplierQuotationDraft` sudah tersedia. Ikat kepada undangan, pemilik, status draft dan batas waktu. Di luar 11 gap katalog. |
| HRM dan Finance | Konfirmasi privat komitmen anggaran serta receipt slip bila Finance menyimpan bukti per item | Workload mTLS dengan scope, digest, versi, tujuan, dan kunci idempotensi. Kemajuan dapat ditampilkan melalui resource manusia yang sudah ada. |

## RBAC dan kriteria publikasi

Owner memeriksa izin serta scope objek kembali di service. Gateway memakai sesi opaque, CSRF untuk Mutation dan MFA untuk aksi sensitif. Node, edges, nodes, pageInfo dan totalCount tunduk pada scope yang sama. Untuk AR, inventory cost, aset dan bank statement, tutup dependensi dan skenario baca/tulis owner sebelum mengaktifkan field ke pengguna. Tabel selisih mengukur rancangan seluruh fase. [Pedoman Gateway](../../how_to_code_wiring_grpc_managements_to_gateway.md)
