# Kalkulasi payroll

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy dan kalkulator tiruan.

| Kasus | Hasil |
| --- | --- |
| Principal lain menjalankan kalkulasi | PermissionDenied |
| Grup dinonaktifkan setelah run dibuat | FailedPrecondition, kalkulator tidak dipanggil |
| Digest input tidak cocok | FailedPrecondition, hasil belum tersimpan |
| Hasil bersih melebihi presisi grup payroll meskipun rekonsiliasi cocok | FailedPrecondition |
| Digest artefak aturan yang dipakai kalkulator tidak cocok | FailedPrecondition |
| Nama komponen pada hasil kalkulator berbeda dari master payroll | FailedPrecondition |
| Profil payroll pekerja dinonaktifkan sebelum kalkulasi | FailedPrecondition |
| Output cocok dan konsisten | Status calculated, digest hasil dan komponen tersimpan |
| Kalkulasi diulang | FailedPrecondition |

`TestPayrollRPCAndApprovalWithPostgreSQL` lulus dengan race detector.

Regresi rule set dengan waktu persetujuan di masa depan lulus melalui gRPC dan PostgreSQL sementara dengan race detector. RPC mengembalikan `FailedPrecondition` dan penghitung panggilan adapter tetap sama. Kalkulasi normal tetap lulus setelah fixture dikembalikan ke rule set yang sudah disetujui.

Regresi gRPC juga membuktikan rule set yang berakhir sehari sebelum akhir periode ditolak dengan `FailedPrecondition` sebelum adapter dipanggil. Fixture mewakili run draf lama yang masih menunjuk rule set tersebut.
