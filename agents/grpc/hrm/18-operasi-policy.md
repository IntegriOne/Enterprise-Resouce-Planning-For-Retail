# List operasi

`human_resource.policy.v1.PolicyService` memiliki 20 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/policy/v1/policy.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreatePolicy` | Membuat entri `policies`. | Mutation | `CreatePolicyRequest` → `PolicyResponse` | `human_resource.policies.create` |
| `UpdatePolicy` | Mengubah entri `policies` sesuai field mask. | Mutation | `UpdatePolicyRequest` → `PolicyResponse` | `human_resource.policies.update` |
| `DeletePolicy` | Menghapus entri `policies` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.policies.delete` |
| `GetPolicyByID` | Mengambil entri `policies` berdasarkan ID. | Query | `IDRequest` → `PolicyResponse` | `human_resource.policies.read` |
| `GetAllPolicies` | Mendaftar entri `policies` dengan paginasi. | Query | `PageRequest` → `PoliciesResponse` | `human_resource.policies.read` |
| `FindPolicyByPolicyCode` | Menyaring daftar `policies` berdasarkan `policy_code`, dengan paginasi. | Query | `StringFinderRequest` → `PoliciesResponse` | `human_resource.policies.read` |
| `FindPolicyByCategoryCode` | Menyaring daftar `policies` berdasarkan `category_code`, dengan paginasi. | Query | `StringFinderRequest` → `PoliciesResponse` | `human_resource.policies.read` |
| `FindPolicyByLegalEntityReferenceID` | Menyaring daftar `policies` berdasarkan `legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PoliciesResponse` | `human_resource.policies.read` |
| `FindPolicyByDocumentVersionReferenceID` | Menyaring daftar `policies` berdasarkan `document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PoliciesResponse` | `human_resource.policies.read` |
| `FindPolicyByApprovedByPrincipalID` | Menyaring daftar `policies` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PoliciesResponse` | `human_resource.policies.read` |
| `CreatePolicyAcknowledgement` | Membuat entri `policy_acknowledgements`. | Mutation | `CreatePolicyAcknowledgementRequest` → `PolicyAcknowledgementResponse` | `human_resource.policy_acknowledgements.create` |
| `UpdatePolicyAcknowledgement` | Mengubah entri `policy_acknowledgements` sesuai field mask. | Mutation | `UpdatePolicyAcknowledgementRequest` → `PolicyAcknowledgementResponse` | `human_resource.policy_acknowledgements.update` |
| `DeletePolicyAcknowledgement` | Menghapus entri `policy_acknowledgements` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.policy_acknowledgements.delete` |
| `GetPolicyAcknowledgementByID` | Mengambil entri `policy_acknowledgements` berdasarkan ID. | Query | `IDRequest` → `PolicyAcknowledgementResponse` | `human_resource.policy_acknowledgements.read` |
| `GetAllPolicyAcknowledgements` | Mendaftar entri `policy_acknowledgements` dengan paginasi. | Query | `PageRequest` → `PolicyAcknowledgementsResponse` | `human_resource.policy_acknowledgements.read` |
| `FindPolicyAcknowledgementByPolicyID` | Menyaring daftar `policy_acknowledgements` berdasarkan `policy_id`, dengan paginasi. | Query | `StringFinderRequest` → `PolicyAcknowledgementsResponse` | `human_resource.policy_acknowledgements.read` |
| `FindPolicyAcknowledgementByWorkerID` | Menyaring daftar `policy_acknowledgements` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PolicyAcknowledgementsResponse` | `human_resource.policy_acknowledgements.read` |
| `FindPolicyAcknowledgementByAcknowledgedByPrincipalID` | Menyaring daftar `policy_acknowledgements` berdasarkan `acknowledged_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PolicyAcknowledgementsResponse` | `human_resource.policy_acknowledgements.read` |
| `FindPolicyAcknowledgementByEvidenceDocumentVersionReferenceID` | Menyaring daftar `policy_acknowledgements` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PolicyAcknowledgementsResponse` | `human_resource.policy_acknowledgements.read` |
| `AcknowledgePolicy` | Mencatat pengakuan atas entri `policy_acknowledgements` melalui transisi workflow HRM. | Mutation | `AcknowledgePolicyRequest` → `PolicyAcknowledgementResponse` | `human_resource.policy_acknowledgements.acknowledge` |

## Butuh GraphQL

13 kandidat Query dan 7 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
