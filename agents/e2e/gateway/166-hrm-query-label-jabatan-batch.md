# Uji HTTPS label jabatan batch

Superadmin memberi izin melalui GraphQL IAM. Stack terisolasi membaca dua job fixture HRM dan unit Enterprise Foundation dalam satu operasi GraphQL.

| Kasus | Hasil |
| --- | --- |
| GID job, UUID job, duplikat, dan ID yang tidak ada | Hanya dua jabatan unik yang ditemukan, dengan urutan input |
| Unit EF diminta bersamaan | Nama unit dikembalikan |
| Input kosong | Daftar jabatan kosong |
| GID pekerja dikirim sebagai job | Ditolak dengan `INVALID_ARGUMENT` |

Go test dan Go vet pada bagian HRM serta Gateway terkait lulus.
