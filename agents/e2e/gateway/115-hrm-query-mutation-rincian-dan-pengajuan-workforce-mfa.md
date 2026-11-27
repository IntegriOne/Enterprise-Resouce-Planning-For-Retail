# Uji HTTPS rincian workforce plan dan pengajuan review

Superadmin fixture menggunakan role IAM dan sesi TOTP pada stack disposable. Badan hukum dan unit organisasi dibuat lewat Foundation, sedangkan job memakai fixture SQL pada PostgreSQL HRM uji.

| Kasus | Hasil |
| --- | --- |
| `hrmCreateWorkforcePlanLine` | Rincian nomor 1 dibuat |
| `hrmUpdateWorkforcePlanLine` dan `hrmWorkforcePlanLine` | Kebutuhan FTE berubah menjadi 3 dan terbaca kembali |
| `hrmDeleteWorkforcePlanLine` untuk rincian kedua yang belum dipakai | Berhasil |
| `hrmSubmitWorkforcePlan` | Status `review` |
| Koreksi rincian saat review | `CONFLICT` |
| `hrmCancelWorkforcePlan` | Status `cancelled` |
| Daftar rincian dan PostgreSQL | Tepat satu rincian tersisa pada rencana yang ditarik, dengan kebutuhan 3,000 FTE |

Assertion awal membandingkan string desimal. Pengulangan lulus setelah nilai mutation dan Query dibandingkan sebagai angka desimal sehingga perbedaan skala tampilan tidak menggagalkan uji. Database dan Redis uji dihapus setelah layanan development kembali sehat.

Regresi dengan katalog IAM resmi lulus pada stack disposable. Role dan assignment dipasang lewat GraphQL. Rincian draf dibuat untuk unit organisasi EF, dikoreksi, dan rincian yang belum dipakai dihapus. Pengajuan menjadi review, koreksi setelah review ditolak, dan penarikan berhasil. Runner diselaraskan dengan payload mutation GraphQL terkini setelah percobaan awal gagal validasi schema. Job tetap memakai fixture HRM.

Uji tambahan membuat dua badan hukum dan dua unit organisasi aktif di EF. Mutation rincian dengan unit dari badan hukum lain menghasilkan `CONFLICT` dan tidak mengembalikan rincian. Mutation dengan unit milik badan hukum rencana berhasil. Daftar akhir tetap berisi satu rincian.

IAM development dibangun ulang dan lulus healthcheck. Superadmin membaca keenam definisi izin melalui HTTPS dan memverifikasi flag sensitivitasnya. Uji mutasi rincian memakai stack disposable.
