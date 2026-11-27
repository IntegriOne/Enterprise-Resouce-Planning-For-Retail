# Uji HTTPS persetujuan, penggantian, dan penghapusan workforce plan

Superadmin fixture memakai grant lewat mutasi GraphQL IAM dan sesi TOTP pada stack disposable. Rencana dan rincian dibuat lewat Gateway. Untuk menguji penggantian, dua versi rencana disiapkan dengan status approved dan referensi anggaran sintetis melalui SQL pada database HRM `*_test`.

| Kasus | Hasil |
| --- | --- |
| `hrmApproveWorkforcePlan` tanpa verifier anggaran Finance | `UNAVAILABLE`, status tetap `review` |
| `hrmDeleteWorkforcePlan` pada draf tanpa rincian | Berhasil |
| `hrmSupersedeWorkforcePlan` pada versi yang belum disetujui | `CONFLICT` |
| Penggantian versi 1 dengan versi 2 setelah fixture berstatus approved | Versi 1 `superseded` dan menunjuk versi 2 |
| PostgreSQL | Rencana review tetap review, versi 1 superseded, versi 2 approved, draf versi 3 sudah dihapus |

Database dan Redis uji dibersihkan setelah layanan development kembali sehat. Persetujuan positif tetap memerlukan verifier Finance. [Uji rincian workforce](115-hrm-query-mutation-rincian-dan-pengajuan-workforce-mfa.md).
