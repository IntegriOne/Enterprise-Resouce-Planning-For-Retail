# Uji HTTPS keputusan dan pembukuan koreksi cuti

Dua akun fixture dengan sesi TOTP mendapat role lewat IAM. Kebijakan annual grant, entitlement dengan saldo awal 10, dan entri ledger pembukaan disiapkan di PostgreSQL HRM disposable. Verifikasi akun kedua menggunakan SQL fixture pada IAM disposable.

| Kasus | Hasil |
| --- | --- |
| Buat dua koreksi `-0.5` | Berhasil |
| Pengusul mencoba menyetujui sendiri | `FORBIDDEN` |
| Employee kedua menyetujui satu koreksi dan menolak yang lain | Berhasil, kode penolakan tersimpan pada hasil mutation |
| Posting koreksi yang disetujui | Saldo 9,5 |
| Pengulangan posting | `CONFLICT` |
| Query ledger dengan filter koreksi yang disetujui | Tepat satu entri |
| PostgreSQL | Saldo 9,5000, revisi 3, dua entri total termasuk pembukaan |

Saat fixture belum memiliki pembukaan atau kebijakan tidak memuat annual entitlement, HRM menolak operasi terkait dengan `CONFLICT`. Percobaan posting tanpa `leave_adjustments.read` ditolak dengan `FORBIDDEN`. Posting lulus setelah role memuat izin baca koreksi di samping izin baca ledger dan izin posting. Semua data serta runtime uji telah dibersihkan. [Handoff izin](../../grpc/hrm/28-izin-iam-untuk-wiring-gateway.md).

## Regresi dengan katalog izin resmi dan akun hasil provisioning

Superadmin memakai katalog HRM dari checkout dan menetapkan role sementara melalui GraphQL. Penyetuju kedua dibuat melalui pengajuan HRM dan `createEmployeeUser`, diaktifkan setelah penautan, lalu login dengan MFA. Verifikasi identitas login penyetuju dan saldo awal cuti disiapkan sebagai fixture pada database disposable.

Pengusul ditolak saat mencoba menyetujui koreksi sendiri. Employee kedua menyetujui satu koreksi dan menolak yang lain. Posting mengubah saldo 10 menjadi 9,5. Pengulangan posting mengembalikan `CONFLICT`, dan query ledger menemukan tepat satu entri untuk koreksi yang disetujui. Runner berakhir dengan status 0 dan stack disposable dibersihkan.
