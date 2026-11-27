# Persetujuan payroll dengan dua hasil pekerja melalui gRPC

Tes menggunakan PostgreSQL sementara, race detector, dan policy tiruan. Hasil payroll kedua dan baris komponennya disiapkan langsung di database untuk menguji aturan persetujuan.

| Kasus | Hasil |
| --- | --- |
| Pembuat run mencoba menyetujui run sendiri | PermissionDenied |
| Penyetuju berbeda menyetujui run ketika hasil pekerja kedua belum memiliki komponen | InvalidArgument |
| Penyetuju mengulangi persetujuan setelah komponen pekerja kedua tersedia | Status approved dan ID penyetuju tersimpan |
| Persetujuan diulang setelah run berstatus approved | FailedPrecondition |
| Jumlah event outbox untuk run setelah rangkaian percobaan persetujuan | Dua event, tanpa duplikasi |
