# Uji sesi MFA HRM melalui HTTPS

IAM, Enterprise Foundation dan HRM memakai tiga database PostgreSQL disposable berakhiran `_test`. IAM, Gateway, dan HRM memakai Redis TLS sementara dengan penyimpanan tmpfs dan ACL yang sama dengan development. Prefix `development` tetap dipakai di instance Redis terpisah. Konfigurasi Compose sementara mengarahkan layanan ke dependency uji selama eksekusi.

| Kasus | Hasil |
| --- | --- |
| Login superadmin fixture, pembuatan role, assignment empat grant dan login ulang | Berhasil |
| Buat badan hukum aktif dan draf workforce | Berhasil |
| `enrollMyTotp`, `confirmMyTotp` dan login ulang | Login menantang MFA |
| `validateUserMfa` dengan TOTP valid | Berhasil |
| `hrmPayrollResults` dengan izin baca dan MFA terkini | Berhasil, total nol pada database uji |
| `hrmDeleteWorkforcePlan` untuk draf yang dibuat | Berhasil |
| Pemeriksaan PostgreSQL setelah penghapusan | Nol draf |
| Logout dan pemulihan runtime development | Berhasil, IAM, kedua Gateway, HRM dan Foundation sehat |

Semua operasi GraphQL memakai HTTPS dan mutasi memakai CSRF sah. Faktor, role, dan sesi fixture berada di database atau Redis disposable yang sudah dihapus. Faktor superadmin development tidak diubah. [Uji penolakan MFA sebelumnya](96-hrm-mutation-seluruh-schema-superadmin-https.md).

## Enam Query sensitif dengan MFA aktif

Pada pengulangan yang sama-sama memakai runtime disposable, role uji juga mendapat `employment_contracts.read` dan `payroll_result_lines.read`. Semua Query berikut memakai sesi TOTP dan ID fixture berbentuk UUID yang tidak ada.

| Query | Hasil |
| --- | --- |
| `hrmEmploymentContract`, `hrmPayrollResult`, `hrmPayrollResultLine` | `NOT_FOUND` dari jalur detail |
| `hrmEmploymentContracts`, `hrmPayrollResults`, `hrmPayrollResultLines` | Daftar kosong, total nol |
| Regresi `hrmDeleteWorkforcePlan` | Berhasil, PostgreSQL kembali mencatat nol draf |

Ketiga database dan Redis uji dihapus lagi setelah layanan development kembali sehat. Hasil ini mencakup keberhasilan melewati MFA serta respons untuk data kosong. Validasi proyeksi data sensitif dengan baris nyata tetap perlu uji terpisah.
