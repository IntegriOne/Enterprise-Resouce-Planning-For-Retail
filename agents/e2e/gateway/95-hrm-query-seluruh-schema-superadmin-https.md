# Uji 42 Query HRM melalui HTTPS dengan superadmin

IAM development memuat 21 kode permission Query dari manifest yang hanya ada pada konteks build lokal. Role uji menerima 20 permission dengan scope global melalui API IAM. Izin baca slip memakai scope self sehingga tidak dimasukkan ke role global. Login ulang dilakukan sebelum pengujian. Database HRM development kosong.

| Hasil aktual | Jumlah | Query |
| --- | ---: | --- |
| NOT_FOUND | 17 | `hrmJob`, `hrmJobFamily`, `hrmLeavePolicy`, `hrmLeaveEntitlement`, `hrmLeaveLedgerEntry`, `hrmLeaveAdjustment`, `hrmOffboardingPlan`, `hrmOffboardingTask`, `hrmPayrollRun`, `hrmPayrollPeriod`, `hrmPayrollGroup`, `hrmPayrollRuleSet`, `hrmPosition`, `hrmWorker`, `hrmWorkerAssignment`, `hrmWorkforcePlan`, `hrmWorkforcePlanLine` |
| OK, total nol | 17 | `hrmJobs`, `hrmJobFamilies`, `hrmLeavePolicies`, `hrmLeaveEntitlements`, `hrmLeaveLedgerEntries`, `hrmLeaveAdjustments`, `hrmOffboardingPlans`, `hrmOffboardingTasks`, `hrmPayrollRuns`, `hrmPayrollPeriods`, `hrmPayrollGroups`, `hrmPayrollRuleSets`, `hrmPositions`, `hrmWorkers`, `hrmWorkerAssignments`, `hrmWorkforcePlans`, `hrmWorkforcePlanLines` |
| FORBIDDEN, MFA diperlukan | 6 | `hrmPayrollResult`, `hrmPayrollResults`, `hrmPayrollResultLine`, `hrmPayrollResultLines`, `hrmEmploymentContract`, `hrmEmploymentContracts` |
| FORBIDDEN, grant self diperlukan | 2 | `hrmMyPayslip`, `hrmMyPayslips` |


Permintaan memakai HTTPS `/api/graphql`, CA lokal, cookie sesi, dan token CSRF saat mengatur role. Pada percobaan awal, tiga Query daftar tanpa filter ditolak dengan `INVALID_ARGUMENT`. Ketiganya berhasil dengan masing-masing filter `entitlement_id`, `entitlement_id`, dan `worker_id`. ID pada Query detail berupa UUID sintetik tanpa baris yang cocok.

Assignment dicabut, role uji dihapus, dan sesi ditutup setelah kedua percobaan. Image IAM uji yang memuat 21 kode masih berjalan di development. Pengujian ini mencakup Query pada database kosong, guard MFA dan scope self. Uji positif MFA, data nonkosong, dan akses slip tetap terpisah. [Uji awal superadmin](94-hrm-query-pekerja-superadmin-https.md).
