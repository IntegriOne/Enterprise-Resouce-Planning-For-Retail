# Uji CSRF mutation HRM dengan sesi superadmin

Uji memakai binary kandidat pada stack HTTPS disposable. Superadmin masuk dengan MFA dan memasang role serta permission melalui GraphQL IAM. Sesi dipastikan aktif melalui `browserSession` sebelum dan sesudah pemeriksaan.

| Kasus | Hasil |
| --- | --- |
| Seluruh mutation domain dengan cookie sesi superadmin dan tanpa header CSRF | 55 dari 55 mengembalikan `FORBIDDEN` karena header CSRF tidak ada. |
| Provisioning sah dengan CSRF | `createEmployeeUser` berhasil dan pengulangan mempertahankan satu akun dan satu receipt. |
| Role uji dan logout | Assignment, permission role, dan role dicabut melalui GraphQL. Logout berhasil. |

Jeda sebelum dan sesudah rangkaian guard menjaga suite di bawah batas rate limit. Percobaan sebelumnya terkena limit saat menjalankan rangkaian tanpa jeda yang cukup. Semua database dan kontainer uji dibersihkan.
