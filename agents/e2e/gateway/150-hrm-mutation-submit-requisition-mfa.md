# Uji HTTPS pengajuan requisition

Stack terisolasi dengan Caddy dan `/api/graphql`. Superadmin membuat role uji serta memberikan permission dan assignment melalui GraphQL IAM. Fixture mengikuti alur badan hukum dan unit EF, pekerja HRM dan tautan akun IAM.

| Kasus | Hasil |
| --- | --- |
| Mencoba submit dengan sesi superadmin sebelum MFA | Ditolak dengan `AUTHORIZATION` |
| Pembuat draf mengajukan setelah MFA | Mutation menghasilkan status `submitted` |
| Baca ulang requisition | Status dan identitas sama dengan payload |
| Submit ulang | Ditolak dengan `CONFLICT` |
| Audit pada database HRM tes | Timestamp pengajuan terisi dan principal pengaju sama dengan pembuat |
