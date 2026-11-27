# Uji HTTPS koneksi Relay requisition

Stack terisolasi menjalankan Gateway dan HRM melalui Caddy. Superadmin memberi izin melalui GraphQL IAM. HRM membuat dua requisition melalui GraphQL dengan unit EF dan hiring manager yang akunnya tertaut di IAM.

| Kasus | Hasil |
| --- | --- |
| `first: 1` | Satu node, edge sesuai node dan cursor akhir, total dua |
| `first: 1` dengan `after` | Halaman kedua berisi node berbeda dan merupakan halaman terakhir |
| `last: 1` dengan `before` | Kembali ke node di halaman pertama |
| `first: 0` | `edges` dan `nodes` kosong, total tetap dua |
| Cursor dengan versi tidak valid | Ditolak dengan `INVALID_ARGUMENT` |

Go test seluruh Gateway serta TypeScript dan ESLint pada pemanggil frontend yang berubah lulus.
