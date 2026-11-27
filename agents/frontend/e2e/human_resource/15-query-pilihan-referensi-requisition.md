# Uji browser penyaringan pilihan referensi requisition

Uji memakai Chrome dan HTTPS pada stack terisolasi. Superadmin masuk dengan MFA. Role dan permission dipasang melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Pilih jabatan perekrut | Posisi perekrut tersedia. Posisi aktif milik jabatan lain tidak tampil. |
| Ganti jabatan | Posisi yang terkait dengan jabatan baru tersedia. |
| Ganti kategori pekerja ke kontraktor | Lini untuk pegawai tidak tampil. Lini muncul kembali setelah kategori dikembalikan. |
| Periksa lini nonperekrutan | Lini retensi tidak ditawarkan. |
| Simpan dan ajukan dengan referensi yang cocok | Draf tersimpan, masuk tahap persetujuan, dan referensi tetap terlihat dengan label bisnis. |

TypeScript, ESLint pada direktori requisition, dan build image frontend lulus.
