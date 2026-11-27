# Uji HTTPS penghapusan draf requisition

Uji berjalan di stack terisolasi dengan Caddy. Superadmin menyiapkan grant dan assignment melalui GraphQL IAM. Fixture memakai unit dari Enterprise Foundation dan pekerja aktif dengan akun IAM tertaut.

| Kasus | Hasil |
| --- | --- |
| Penghapusan sebelum MFA | Ditolak dengan `AUTHORIZATION` |
| Penghapusan requisition yang telah diajukan | Ditolak dengan `CONFLICT`, detail tetap menunjukkan `submitted` |
| Penghapusan draf dengan MFA | Payload memberi `deleted: true` |
| Detail draf setelah dihapus | Draf tidak ditemukan |
