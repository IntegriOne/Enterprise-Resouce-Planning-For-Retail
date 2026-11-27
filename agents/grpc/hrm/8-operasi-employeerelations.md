# List operasi

`human_resource.employeerelations.v1.EmployeeRelationsService` memiliki 34 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/employeerelations/v1/employeerelations.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateEmployeeRelationsCase` | Membuat entri `employee_relations_cases`. | Mutation | `CreateEmployeeRelationsCaseRequest` → `EmployeeRelationsCaseResponse` | `human_resource.employee_relations_cases.create` |
| `UpdateEmployeeRelationsCase` | Mengubah entri `employee_relations_cases` sesuai field mask. | Mutation | `UpdateEmployeeRelationsCaseRequest` → `EmployeeRelationsCaseResponse` | `human_resource.employee_relations_cases.update` |
| `DeleteEmployeeRelationsCase` | Menghapus entri `employee_relations_cases` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.employee_relations_cases.delete` |
| `GetEmployeeRelationsCaseByID` | Mengambil entri `employee_relations_cases` berdasarkan ID. | Query | `IDRequest` → `EmployeeRelationsCaseResponse` | `human_resource.employee_relations_cases.read` |
| `GetAllEmployeeRelationsCases` | Mendaftar entri `employee_relations_cases` dengan paginasi. | Query | `PageRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByCaseNumber` | Menyaring daftar `employee_relations_cases` berdasarkan `case_number`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByCaseType` | Menyaring daftar `employee_relations_cases` berdasarkan `case_type`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByStatus` | Menyaring daftar `employee_relations_cases` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseBySubjectWorkerID` | Menyaring daftar `employee_relations_cases` berdasarkan `subject_worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByReportingPartyReferenceID` | Menyaring daftar `employee_relations_cases` berdasarkan `reporting_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByAssignedInvestigatorPrincipalID` | Menyaring daftar `employee_relations_cases` berdasarkan `assigned_investigator_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByOutcomeCode` | Menyaring daftar `employee_relations_cases` berdasarkan `outcome_code`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByRestrictedSummaryReferenceID` | Menyaring daftar `employee_relations_cases` berdasarkan `restricted_summary_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByLegalHoldReferenceID` | Menyaring daftar `employee_relations_cases` berdasarkan `legal_hold_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `FindEmployeeRelationsCaseByIdempotencyKey` | Menyaring daftar `employee_relations_cases` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsCasesResponse` | `human_resource.employee_relations_cases.read` |
| `CreateEmployeeRelationsEvent` | Membuat entri `employee_relations_events`. | Mutation | `CreateEmployeeRelationsEventRequest` → `EmployeeRelationsEventResponse` | `human_resource.employee_relations_events.create` |
| `UpdateEmployeeRelationsEvent` | Mengubah entri `employee_relations_events` sesuai field mask. | Mutation | `UpdateEmployeeRelationsEventRequest` → `EmployeeRelationsEventResponse` | `human_resource.employee_relations_events.update` |
| `DeleteEmployeeRelationsEvent` | Menghapus entri `employee_relations_events` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IdentityIDRequest` → `EmptyResponse` | `human_resource.employee_relations_events.delete` |
| `GetEmployeeRelationsEventByID` | Mengambil entri `employee_relations_events` berdasarkan ID. | Query | `IdentityIDRequest` → `EmployeeRelationsEventResponse` | `human_resource.employee_relations_events.read` |
| `GetAllEmployeeRelationsEvents` | Mendaftar entri `employee_relations_events` dengan paginasi. | Query | `PageRequest` → `EmployeeRelationsEventsResponse` | `human_resource.employee_relations_events.read` |
| `FindEmployeeRelationsEventByCaseID` | Menyaring daftar `employee_relations_events` berdasarkan `case_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsEventsResponse` | `human_resource.employee_relations_events.read` |
| `FindEmployeeRelationsEventByEventType` | Menyaring daftar `employee_relations_events` berdasarkan `event_type`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsEventsResponse` | `human_resource.employee_relations_events.read` |
| `FindEmployeeRelationsEventByActorPrincipalID` | Menyaring daftar `employee_relations_events` berdasarkan `actor_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsEventsResponse` | `human_resource.employee_relations_events.read` |
| `FindEmployeeRelationsEventByEventReferenceID` | Menyaring daftar `employee_relations_events` berdasarkan `event_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmployeeRelationsEventsResponse` | `human_resource.employee_relations_events.read` |
| `CreateCollectiveAgreement` | Membuat entri `collective_agreements`. | Mutation | `CreateCollectiveAgreementRequest` → `CollectiveAgreementResponse` | `human_resource.collective_agreements.create` |
| `UpdateCollectiveAgreement` | Mengubah entri `collective_agreements` sesuai field mask. | Mutation | `UpdateCollectiveAgreementRequest` → `CollectiveAgreementResponse` | `human_resource.collective_agreements.update` |
| `DeleteCollectiveAgreement` | Menghapus entri `collective_agreements` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.collective_agreements.delete` |
| `GetCollectiveAgreementByID` | Mengambil entri `collective_agreements` berdasarkan ID. | Query | `IDRequest` → `CollectiveAgreementResponse` | `human_resource.collective_agreements.read` |
| `GetAllCollectiveAgreements` | Mendaftar entri `collective_agreements` dengan paginasi. | Query | `PageRequest` → `CollectiveAgreementsResponse` | `human_resource.collective_agreements.read` |
| `FindCollectiveAgreementByAgreementCode` | Menyaring daftar `collective_agreements` berdasarkan `agreement_code`, dengan paginasi. | Query | `StringFinderRequest` → `CollectiveAgreementsResponse` | `human_resource.collective_agreements.read` |
| `FindCollectiveAgreementByLegalEntityReferenceID` | Menyaring daftar `collective_agreements` berdasarkan `legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CollectiveAgreementsResponse` | `human_resource.collective_agreements.read` |
| `FindCollectiveAgreementByRepresentativePartyReferenceID` | Menyaring daftar `collective_agreements` berdasarkan `representative_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CollectiveAgreementsResponse` | `human_resource.collective_agreements.read` |
| `FindCollectiveAgreementByJurisdictionCode` | Menyaring daftar `collective_agreements` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `CollectiveAgreementsResponse` | `human_resource.collective_agreements.read` |
| `FindCollectiveAgreementByDocumentVersionReferenceID` | Menyaring daftar `collective_agreements` berdasarkan `document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CollectiveAgreementsResponse` | `human_resource.collective_agreements.read` |

## Butuh GraphQL

25 kandidat Query dan 9 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
