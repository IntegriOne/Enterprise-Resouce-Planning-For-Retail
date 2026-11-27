# Melanjutkan penautan akun employee

Stack Caddy HTTPS terisolasi menggunakan simulasi respons penautan seperti [uji pemulihan](../../../e2e/gateway/137-iam-mutation-pemulihan-penautan-akun.md). Operator masuk sebagai superadmin dengan MFA. Grant sementara dipasang dan dicabut lewat GraphQL.

| Skenario | Hasil |
| --- | --- |
| Operator memilih pengajuan yang masih `processing` | Opsi bertuliskan Sedang diproses dan tombol bertuliskan Lanjutkan penautan. |
| Detail pengajuan | Nama, nomor pekerja, dan badan hukum terlihat. Teks form tidak memuat `gid://`. |
| Operator mengirim ulang dari browser | Mutation membawa ID worker dan pengajuan yang dipilih. Modal ditutup sesudah berhasil. |
| Verifikasi akhir | IAM menyimpan satu akun. Penautan cocok dengan HRM dan akun dapat diaktifkan. |

Runner browser menunggu modal benar-benar terbuka setelah hidrasi. Percobaan awal mengklik tombol sebelum modal dapat dibuka. Pengujian ulang lulus.

Uji browser berikutnya memastikan field Nama lengkap terisi dari profil pekerja EF setelah pengajuan dipilih. Runner mengisi field akun lainnya dan memeriksa bahwa mutation membawa nama dari EF. Pengulangan penautan tetap memakai akun yang sudah tersimpan. Operator masih dapat mengubah nama pada form. Pemilihan pengajuan lain atau pemuatan ulang daftar menghapus edit nama yang terkait dengan pengajuan sebelumnya.

Build final lulus uji browser terisolasi, lalu dipasang pada kedua replika frontend development. Kedua replika lulus healthcheck.
