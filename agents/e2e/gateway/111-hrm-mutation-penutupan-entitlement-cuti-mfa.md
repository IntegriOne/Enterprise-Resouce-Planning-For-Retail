# Uji HTTPS penutupan entitlement cuti dengan MFA

Uji memakai dua principal fixture dengan role IAM dan TOTP. Entitlement pada database disposable berakhir 31 Oktober 2026 dengan saldo 10 dari ledger pembukaan. Setelah employee kedua menyetujui koreksi, percobaan penutupan dilakukan sebelum dan sesudah posting.

| Kasus | Hasil |
| --- | --- |
| Tutup periode yang sudah berakhir dengan saldo 10 | `CONFLICT` |
| Posting koreksi independen sebesar -10 | Berhasil, saldo nol |
| Tutup entitlement bersaldo nol | Berhasil, `closed_at` terisi |
| Ulangi penutupan | `CONFLICT` |
| Pemeriksaan PostgreSQL | Saldo nol, revisi 3, waktu dan principal penutup tercatat. Ledger tetap berisi dua entri, pembukaan dan koreksi. |

Database dan Redis uji dibersihkan. Layanan development kembali sehat. [Uji posting koreksi](110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md).
