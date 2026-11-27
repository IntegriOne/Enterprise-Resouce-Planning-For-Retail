# Uji HTTPS Query konfigurasi payroll dengan data

Grup, periode, dan aturan payroll disiapkan sebagai fixture SQL pada database HRM disposable. Sesi superadmin fixture diberi role melalui mutasi GraphQL IAM dan membuat satu run draf lewat Gateway.

| Kasus | Hasil |
| --- | --- |
| `hrmPayrollGroup`, `hrmPayrollPeriod`, `hrmPayrollRuleSet`, `hrmPayrollRun` | ID yang diminta cocok pada setiap detail |
| `hrmPayrollGroups`, `hrmPayrollRuleSets` | Satu item dan total satu pada tiap daftar |
| `hrmPayrollPeriods` dengan filter grup | Periode yang sesuai, total satu |
| `hrmPayrollRuns` dengan filter periode dan idempotency key | Run draf yang sama, total satu pada masing-masing filter |
| Filter dengan nilai yang tidak ada | Daftar kosong dan total nol |
| Pemeriksaan PostgreSQL | Run tetap draft, waktu kalkulasi dan pengajuan jurnal kosong |

Database dan Redis uji dihapus setelah development kembali sehat. [Uji pembuatan draf](107-hrm-query-mutation-draf-payroll-mfa.md).
