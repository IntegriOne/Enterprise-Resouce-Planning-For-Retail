# Uji HTTPS edit unit dan penetapan kepala

Superadmin menyiapkan pekerja tertaut IAM–HRM dan unit Enterprise Foundation pada stack terisolasi. Role dan permission uji dikelola melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Tetapkan kepala memakai GID pekerja | UUID pekerja tersimpan |
| Edit nama dengan `preserve_head_person: true` | Nama berubah dan kepala yang sudah ditetapkan tetap tersimpan |
| Kirim preserve bersama ID kepala | Ditolak dengan `ARGUMENT` |
| Lepas kepala melalui aksi penetapan | Referensi kepala dikosongkan |

Service memakai row lock PostgreSQL sebelum mutasi detail, penetapan kepala, dan perubahan status unit. Race test EF dan Gateway, Go vet, serta pemeriksaan frontend lulus.
