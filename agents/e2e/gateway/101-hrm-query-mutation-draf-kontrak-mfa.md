# Uji HTTPS draf kontrak kerja dengan MFA

Perangkat uji memakai database disposable untuk IAM, Foundation dan HRM, serta Redis TLS sementara. Superadmin fixture mendapat grant pembuatan pekerja dan empat izin kontrak melalui IAM. Setelah fixture badan hukum, party dan pekerja dibuat melalui API, superadmin login ulang dengan TOTP yang telah didaftarkan pada akun fixture.

| Kasus | Hasil |
| --- | --- |
| `hrmCreateEmploymentContract` dengan pekerja preboarding dan badan hukum aktif | Berhasil, status `draft` |
| Pembuatan ulang kontrak yang sama | `CONFLICT` |
| `hrmUpdateEmploymentContract` dengan tanggal akhir baru | Berhasil |
| `hrmEmploymentContract` dan `hrmEmploymentContracts` | Detail memuat tanggal baru, daftar pekerja memuat satu kontrak |
| `hrmDeleteEmploymentContract` | Berhasil |
| Daftar kontrak sesudah penghapusan | Total nol |
| Pemeriksaan PostgreSQL setelah penghapusan | Nol kontrak |

Layanan development kembali sehat setelah database dan Redis uji dibersihkan. Verifikasi tanda tangan melalui DRM berada di luar skenario draf ini. [Pemakaian MFA pada stack uji](100-hrm-query-mutation-mfa-payroll-workforce.md).
