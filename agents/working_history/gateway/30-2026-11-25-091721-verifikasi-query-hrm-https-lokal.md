# Verifikasi Query HRM melalui HTTPS lokal

Bangun ulang image Gateway dan aktifkan HRM pada kedua replika. Kedua replika sehat. Sembilan jalur query HRM menolak sesi anonim di kedua port langsung dan Caddy HTTPS dengan sertifikat yang diverifikasi. Katalog IAM lokal belum berisi permission HRM sehingga akses positif dengan grant resmi belum dapat diverifikasi. [Matriks HTTPS](../../e2e/gateway/18-hrm-query-akses-anonim-https.md).
