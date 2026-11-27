# Uji gRPC perubahan pemilik slip PDF

`TestPayrollRPCAndApprovalWithPostgreSQL` memakai database HRM terisolasi dan fixture reader dokumen. Tes dijalankan dengan race detector dan tag integration.

| Kasus | Hasil |
| --- | --- |
| Tautan pekerja berlaku selama pembacaan | PDF yang cocok dengan versi dan digest dikembalikan |
| Tautan dilepas setelah verifikasi awal, sebelum reader mengembalikan PDF | RPC menghasilkan `PermissionDenied` tanpa isi PDF |
| Reader mengembalikan versi salah, digest salah atau PDF terlalu besar | RPC menolak konten |
| Akun tertaut ke dua riwayat kerja dan satu tautan dilepas ketika halaman terakhir yang kosong diproses | RPC menghasilkan `PermissionDenied` tanpa halaman atau total lama |
