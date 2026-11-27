# Uji HTTPS replay provisioning IAM dan refetch permintaan HRM

Runner repositori membangun ulang EF, IAM, HRM, dan Gateway dari worktree, lalu menjalankannya terhadap PostgreSQL dan Redis terisolasi di belakang Caddy. Superadmin memberikan role dan permission melalui GraphQL.

| Kasus | Hasil |
| --- | --- |
| Worker belum memiliki event aktivasi | Pengajuan akun ditolak |
| Aktivasi dan pengajuan ulang | Satu pengajuan dikembalikan |
| Worker berbeda dari receipt IAM | Provisioning ditolak |
| Pembuatan akun dan pengulangan request yang sama | Respons berisi akun IAM yang sama |
| Status setelah aktivasi akun | Receipt IAM berstatus `linked`, permintaan HRM berstatus `applied` |
| Relay refetch permintaan HRM | Tipe `EmployeeAccountRequest` dan isi sama dengan Query detail |
| Pemeriksaan persistence | Tepat satu akun dan satu receipt tersimpan, serta principal HRM cocok dengan akun IAM |

Kontrak bertanda tangan pada setup memakai fixture di database HRM tes. Runner menghapus kontainer dan database disposable setelah selesai.

Uji tambahan membuktikan lookup receipt menerima UUID maupun GID `employee-provisioning-request`. GID akun IAM dengan UUID receipt yang valid ditolak pada Query detail dan Mutation provisioning. Pembuatan memakai UUID dan replay memakai GID receipt yang sesuai mengembalikan akun yang sama.

Uji setelah perubahan korelasi receipt membuktikan alur provisioning tetap selesai melalui IAM–HRM. Tes Gateway dengan race detector menguji receipt dengan ID pengajuan, pekerja, atau badan hukum berbeda serta field yang hilang. Seluruh kasus penolakan mempertahankan status awal.
