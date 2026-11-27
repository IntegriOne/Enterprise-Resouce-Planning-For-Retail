# Uji HTTPS referensi Enterprise Foundation di HRM

Superadmin menyiapkan grant melalui GraphQL IAM. Badan hukum, party, unit, dan lokasi dibuat lewat GraphQL Enterprise Foundation. Tes memakai ID yang dikembalikan payload EF.

| Kasus | Hasil |
| --- | --- |
| Create worker dengan GID party dan badan hukum | Pekerja tercatat dan dapat diprovisioning di IAM |
| Filter worker memakai GID badan hukum | Pekerja pada badan hukum tersebut ditemukan |
| Create requisition dengan GID job, unit, dan lokasi | Draf berhasil dicatat dan dibaca ulang |
| GID party dipakai untuk referensi unit organisasi | Mutation menolak input |
| Revisi dan submit requisition yang dibuat dengan GID EF | Perubahan draf dan submit berhasil sesuai workflow |

Go test dan Go vet seluruh Gateway lulus. Bukti HTTPS ini mencakup worker dan requisition. Jalur posisi, workforce plan, kontrak, dan payroll turut memakai normalisasi referensi EF pada input yang relevan.
