# Uji HTTPS Query cuti dengan data

Kebijakan annual grant, entitlement, dan ledger pembukaan disiapkan pada database HRM disposable. Pengajuan, persetujuan, penolakan, dan posting koreksi dijalankan melalui HTTPS menggunakan dua akun fixture dengan role IAM dan TOTP.

| Query | Hasil |
| --- | --- |
| `hrmLeavePolicy` dan `hrmLeavePolicies` | Detail sesuai ID dan satu kebijakan pada daftar |
| `hrmLeaveEntitlement` dan `hrmLeaveEntitlements` | Detail sesuai ID dan satu entitlement pada filter pekerja |
| `hrmLeaveLedgerEntry` dan `hrmLeaveLedgerEntries` | Detail pembukaan sesuai ID, dua entri pada filter entitlement, satu entri pada filter koreksi |
| `hrmLeaveAdjustment` dan `hrmLeaveAdjustments` | Detail sesuai ID, dua koreksi pada filter entitlement |
| Pemeriksaan PostgreSQL | Saldo 9,5000, revisi 3, dua entri ledger |

Database dan Redis uji dibersihkan setelah development sehat. [Uji posting koreksi](110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md).
