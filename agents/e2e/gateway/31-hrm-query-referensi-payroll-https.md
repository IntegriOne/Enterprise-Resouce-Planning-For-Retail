# Verifikasi HTTPS Query referensi payroll

Kedua replika Gateway dibangun ulang dan sehat. Tes Caddy memverifikasi CA development serta mengirim permintaan tanpa sesi.

| Query | Gateway 1 | Gateway 2 | Caddy HTTPS |
| --- | --- | --- | --- |
| `hrmPayrollPeriods` | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |
| `hrmPayrollGroups` | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |
| `hrmPayrollRuleSets` | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |
| `hrmPayrollRuns` dengan filter kunci idempotensi | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |

Kasus HTTPS dapat diulang melalui `go test -tags=e2e ./internal/e2e -run '^TestHRMPayrollQueriesRequireSessionOverHTTPS$' -count=1` dari modul Gateway. Seluruh 12 pemeriksaan lulus. Uji positif tetap memerlukan grant IAM resmi.
