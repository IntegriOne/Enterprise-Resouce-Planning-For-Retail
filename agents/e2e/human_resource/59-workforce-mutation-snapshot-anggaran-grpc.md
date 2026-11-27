# Uji gRPC persetujuan workforce plan

Suite workforce planning lulus dengan race detector dan PostgreSQL disposable. Verifier anggaran dan Enterprise Foundation memakai fixture.

| Kasus | Hasil |
| --- | --- |
| Pembuat mencoba menyetujui rencana sendiri | Ditolak |
| Verifier anggaran menolak snapshot | Ditolak |
| Verifier menerima kategori employee, tindakan recruit, dan lokasi kerja yang dipilih | Rencana disetujui dan referensi anggaran tersimpan |
| Event rencana setelah persetujuan | Jumlah sesuai dengan transisi yang berhasil |

[Kontrak snapshot untuk Finance](../../notes/finance_accounting/2-human_resource-workforce-budget.md).
