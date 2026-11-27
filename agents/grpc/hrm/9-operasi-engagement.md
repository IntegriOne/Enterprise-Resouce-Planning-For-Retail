# List operasi

`human_resource.engagement.v1.EngagementService` memiliki 17 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/engagement/v1/engagement.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateEngagementSurvey` | Membuat entri `engagement_surveys`. | Mutation | `CreateEngagementSurveyRequest` → `EngagementSurveyResponse` | `human_resource.engagement_surveys.create` |
| `UpdateEngagementSurvey` | Mengubah entri `engagement_surveys` sesuai field mask. | Mutation | `UpdateEngagementSurveyRequest` → `EngagementSurveyResponse` | `human_resource.engagement_surveys.update` |
| `DeleteEngagementSurvey` | Menghapus entri `engagement_surveys` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.engagement_surveys.delete` |
| `GetEngagementSurveyByID` | Mengambil entri `engagement_surveys` berdasarkan ID. | Query | `IDRequest` → `EngagementSurveyResponse` | `human_resource.engagement_surveys.read` |
| `GetAllEngagementSurveys` | Mendaftar entri `engagement_surveys` dengan paginasi. | Query | `PageRequest` → `EngagementSurveysResponse` | `human_resource.engagement_surveys.read` |
| `FindEngagementSurveyBySurveyCode` | Menyaring daftar `engagement_surveys` berdasarkan `survey_code`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveysResponse` | `human_resource.engagement_surveys.read` |
| `FindEngagementSurveyByLegalEntityReferenceID` | Menyaring daftar `engagement_surveys` berdasarkan `legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveysResponse` | `human_resource.engagement_surveys.read` |
| `FindEngagementSurveyByTopicCode` | Menyaring daftar `engagement_surveys` berdasarkan `topic_code`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveysResponse` | `human_resource.engagement_surveys.read` |
| `FindEngagementSurveyByQuestionnaireDocumentVersionReferenceID` | Menyaring daftar `engagement_surveys` berdasarkan `questionnaire_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveysResponse` | `human_resource.engagement_surveys.read` |
| `CreateEngagementSurveyResult` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `CreateEngagementSurveyResultRequest` → `EngagementSurveyResultResponse` | — |
| `UpdateEngagementSurveyResult` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `UpdateEngagementSurveyResultRequest` → `EngagementSurveyResultResponse` | — |
| `DeleteEngagementSurveyResult` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `IDRequest` → `EmptyResponse` | — |
| `GetEngagementSurveyResultByID` | Mengambil entri `engagement_survey_results` berdasarkan ID. | Query | `IDRequest` → `EngagementSurveyResultResponse` | `human_resource.engagement_survey_results.read` |
| `GetAllEngagementSurveyResults` | Mendaftar entri `engagement_survey_results` dengan paginasi. | Query | `PageRequest` → `EngagementSurveyResultsResponse` | `human_resource.engagement_survey_results.read` |
| `FindEngagementSurveyResultBySurveyID` | Menyaring daftar `engagement_survey_results` berdasarkan `survey_id`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveyResultsResponse` | `human_resource.engagement_survey_results.read` |
| `FindEngagementSurveyResultByOrganizationUnitReferenceID` | Menyaring daftar `engagement_survey_results` berdasarkan `organization_unit_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveyResultsResponse` | `human_resource.engagement_survey_results.read` |
| `FindEngagementSurveyResultByQuestionCode` | Menyaring daftar `engagement_survey_results` berdasarkan `question_code`, dengan paginasi. | Query | `StringFinderRequest` → `EngagementSurveyResultsResponse` | `human_resource.engagement_survey_results.read` |

## Butuh GraphQL

11 kandidat Query dan 3 kandidat Mutation dan 3 RPC nonaktif. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
