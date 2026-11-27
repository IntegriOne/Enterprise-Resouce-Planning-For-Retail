# Verifikasi HTTPS hasil payroll

Kedua replika Gateway dibangun ulang dan sehat. Suite HTTPS melalui Caddy dengan CA development terverifikasi mencakup 12 skenario anonim.

| Skenario | Hasil |
| --- | --- |
| Daftar periode, grup, dan aturan | UNAUTHENTICATED |
| Detail periode, grup, dan aturan | UNAUTHENTICATED dan field detail null |
| Daftar periode dengan filter grup | UNAUTHENTICATED |
| Pencarian run dengan kunci idempotensi | UNAUTHENTICATED |
| Detail hasil dan komponen payroll | UNAUTHENTICATED dan field detail null |
| Daftar hasil per run dan komponen per hasil | UNAUTHENTICATED |

Ulangi dengan `go test -race -tags=e2e ./internal/e2e -run '^TestHRMPayrollQueriesRequireSessionOverHTTPS$' -count=1`. Jalur positif terkini dengan MFA dan izin resmi masih terbuka.
