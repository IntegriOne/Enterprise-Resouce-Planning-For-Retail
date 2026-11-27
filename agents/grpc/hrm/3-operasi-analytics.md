# List operasi

`human_resource.analytics.v1.AnalyticsService` memiliki 16 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/analytics/v1/analytics.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateHumanCapitalMetricDefinition` | Membuat entri `human_capital_metric_definitions`. | Mutation | `CreateHumanCapitalMetricDefinitionRequest` → `HumanCapitalMetricDefinitionResponse` | `human_resource.human_capital_metric_definitions.create` |
| `UpdateHumanCapitalMetricDefinition` | Mengubah entri `human_capital_metric_definitions` sesuai field mask. | Mutation | `UpdateHumanCapitalMetricDefinitionRequest` → `HumanCapitalMetricDefinitionResponse` | `human_resource.human_capital_metric_definitions.update` |
| `DeleteHumanCapitalMetricDefinition` | Menghapus entri `human_capital_metric_definitions` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.human_capital_metric_definitions.delete` |
| `GetHumanCapitalMetricDefinitionByID` | Mengambil entri `human_capital_metric_definitions` berdasarkan ID. | Query | `IDRequest` → `HumanCapitalMetricDefinitionResponse` | `human_resource.human_capital_metric_definitions.read` |
| `GetAllHumanCapitalMetricDefinitions` | Mendaftar entri `human_capital_metric_definitions` dengan paginasi. | Query | `PageRequest` → `HumanCapitalMetricDefinitionsResponse` | `human_resource.human_capital_metric_definitions.read` |
| `FindHumanCapitalMetricDefinitionByMetricCode` | Menyaring daftar `human_capital_metric_definitions` berdasarkan `metric_code`, dengan paginasi. | Query | `StringFinderRequest` → `HumanCapitalMetricDefinitionsResponse` | `human_resource.human_capital_metric_definitions.read` |
| `FindHumanCapitalMetricDefinitionByUnitCode` | Menyaring daftar `human_capital_metric_definitions` berdasarkan `unit_code`, dengan paginasi. | Query | `StringFinderRequest` → `HumanCapitalMetricDefinitionsResponse` | `human_resource.human_capital_metric_definitions.read` |
| `CreateHumanCapitalMetricSnapshot` | Membuat entri `human_capital_metric_snapshots`. | Mutation | `CreateHumanCapitalMetricSnapshotRequest` → `HumanCapitalMetricSnapshotResponse` | `human_resource.human_capital_metric_snapshots.create` |
| `UpdateHumanCapitalMetricSnapshot` | Mengubah entri `human_capital_metric_snapshots` sesuai field mask. | Mutation | `UpdateHumanCapitalMetricSnapshotRequest` → `HumanCapitalMetricSnapshotResponse` | `human_resource.human_capital_metric_snapshots.update` |
| `DeleteHumanCapitalMetricSnapshot` | Menghapus entri `human_capital_metric_snapshots` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.human_capital_metric_snapshots.delete` |
| `GetHumanCapitalMetricSnapshotByID` | Mengambil entri `human_capital_metric_snapshots` berdasarkan ID. | Query | `IDRequest` → `HumanCapitalMetricSnapshotResponse` | `human_resource.human_capital_metric_snapshots.read` |
| `GetAllHumanCapitalMetricSnapshots` | Mendaftar entri `human_capital_metric_snapshots` dengan paginasi. | Query | `PageRequest` → `HumanCapitalMetricSnapshotsResponse` | `human_resource.human_capital_metric_snapshots.read` |
| `FindHumanCapitalMetricSnapshotByMetricDefinitionID` | Menyaring daftar `human_capital_metric_snapshots` berdasarkan `metric_definition_id`, dengan paginasi. | Query | `StringFinderRequest` → `HumanCapitalMetricSnapshotsResponse` | `human_resource.human_capital_metric_snapshots.read` |
| `FindHumanCapitalMetricSnapshotByOrganizationUnitReferenceID` | Menyaring daftar `human_capital_metric_snapshots` berdasarkan `organization_unit_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `HumanCapitalMetricSnapshotsResponse` | `human_resource.human_capital_metric_snapshots.read` |
| `FindHumanCapitalMetricSnapshotByJurisdictionCode` | Menyaring daftar `human_capital_metric_snapshots` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `HumanCapitalMetricSnapshotsResponse` | `human_resource.human_capital_metric_snapshots.read` |
| `FindHumanCapitalMetricSnapshotByDataQualityStatus` | Menyaring daftar `human_capital_metric_snapshots` berdasarkan `data_quality_status`, dengan paginasi. | Query | `StringFinderRequest` → `HumanCapitalMetricSnapshotsResponse` | `human_resource.human_capital_metric_snapshots.read` |

## Butuh GraphQL

10 kandidat Query dan 6 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
