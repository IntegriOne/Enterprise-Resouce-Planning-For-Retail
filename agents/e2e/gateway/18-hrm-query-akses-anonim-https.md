# Pengujian akses anonim Query HRM pada Gateway lokal

Image Gateway dibangun ulang dan kedua replika sehat dengan HRM aktif. Pengujian memakai port langsung kedua replika dan Caddy HTTPS. Permintaan HTTPS memverifikasi sertifikat dengan CA development.

| Query | Gateway 1 | Gateway 2 | Caddy HTTPS |
| --- | --- | --- | --- |
| `hrmWorker` | Ditolak | Ditolak | Ditolak |
| `hrmWorkers` | Ditolak | Ditolak | Ditolak |
| `node` dengan GID pekerja | Ditolak | Ditolak | Ditolak |
| `hrmWorkforcePlan` | Ditolak | Ditolak | Ditolak |
| `hrmWorkforcePlans` | Ditolak | Ditolak | Ditolak |
| `hrmWorkforcePlanLine` | Ditolak | Ditolak | Ditolak |
| `hrmWorkforcePlanLines` | Ditolak | Ditolak | Ditolak |
| `hrmMyPayslip` | Ditolak | Ditolak | Ditolak |
| `hrmMyPayslips` | Ditolak | Ditolak | Ditolak |

Seluruh 27 respons mencantumkan kode `UNAUTHENTICATED` dan path field yang dipanggil. Pemeriksaan baca-saja menemukan nol permission dengan prefiks `human_resource.` dalam katalog IAM lokal. Uji positif dengan akun dan grant resmi masih terbuka.
