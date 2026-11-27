# Uji HTTPS pembuatan draf requisition

Uji berjalan lewat Caddy dan `/api/graphql` pada stack terisolasi. Superadmin login, lalu membuat role uji dan menambahkan permission serta assignment melalui GraphQL IAM. EF membuat badan hukum dan unit. HRM mengaktifkan pekerja, dan IAM menautkan akun melalui `createEmployeeUser`. Job aktif disiapkan sebagai fixture di database HRM tes.

| Kasus | Hasil |
| --- | --- |
| ID pekerja tidak ditemukan | Mutation menolak dan tidak mengembalikan draf |
| Headcount nol | Mutation menolak dan tidak mengembalikan draf |
| Pekerja aktif dengan akun tertaut, job aktif dan unit EF valid | Mutation menghasilkan draf dengan headcount default satu |
| Baca ulang draf lewat GraphQL | ID, nomor, status dan headcount sama |
| Verifikasi persistence di database tes | Principal hiring manager sama dengan akun yang ditautkan IAM |
| Pengulangan provisioning IAM | Akun dan receipt tetap masing-masing satu |
| Label jabatan dan unit di HTTPS | `hrmJob` menghasilkan `REC / Recruiter`; `organizationUnitsByIds` menghasilkan `e2e-rec / Recruiting` sesuai normalisasi kode EF |
| Label hiring manager di HTTPS | `hrmWorker` menghasilkan nomor pekerja; `partiesByIds` dari referensi pekerja menghasilkan nama tampil |
