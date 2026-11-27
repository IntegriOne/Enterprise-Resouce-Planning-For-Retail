# Uji HTTPS lifecycle posisi melalui MFA

Superadmin fixture memperoleh lima izin posisi melalui IAM dan login dengan TOTP. Badan hukum aktif dan unit organisasi dibuat melalui API Foundation. Job family dan job sintetis disiapkan di database HRM disposable.

| Kasus | Hasil |
| --- | --- |
| `hrmCreatePosition` dengan referensi job dan organisasi | Status `planned` |
| `hrmUpdatePosition` | Judul berubah menjadi `E2E Revised` |
| `hrmActivatePosition` | Status `active` |
| `hrmDeletePosition` pada posisi planned lain yang belum dipakai | Berhasil |
| Hapus posisi aktif | `CONFLICT` |
| `hrmPositions` dengan filter badan hukum dan pemeriksaan PostgreSQL | Tepat satu posisi tersisa, aktif dengan judul baru |

Database dan Redis uji dibersihkan. Kedua replika Gateway, IAM, HRM, serta Foundation kembali sehat pada konfigurasi development.

Regresi HTTPS terbaru memakai izin resmi dari manifest IAM. Role dan assignment uji dipasang melalui GraphQL. Badan hukum dan unit organisasi dibuat di EF. Job memakai fixture HRM. Pembuatan posisi planned, koreksi judul, aktivasi, penghapusan posisi planned yang belum dipakai, penolakan penghapusan posisi aktif dan pembacaan daftar lulus.

IAM development memakai image terbaru dan lulus healthcheck. Query HTTPS superadmin mengonfirmasi 13 izin Personnel dan flag sensitivitasnya. Pemeriksaan kesesuaian manifest dengan pemetaan serta policy gRPC HRM juga lulus pada seluruh izin resmi HRM.
