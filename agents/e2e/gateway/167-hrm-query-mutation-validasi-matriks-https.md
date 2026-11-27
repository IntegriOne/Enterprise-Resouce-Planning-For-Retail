# Validasi matriks Query dan Mutation HRM

Pemeriksaan memakai Gateway development melalui Caddy dan CA lokal. Skrip `backend/gateway/internal/e2e/scripts/check_hrm_schema_guards.py` membaca field dan argumen dari SDL aktif.

| Kasus | Hasil |
| --- | --- |
| Query tanpa sesi | 52 dari 52 mengembalikan `UNAUTHENTICATED`. |
| Mutation tanpa header CSRF | 55 dari 55 mengembalikan `FORBIDDEN` dengan alasan header CSRF tidak ada. |
| Pencocokan matriks dengan deklarasi root SDL | Semua 107 operasi domain tercantum tepat satu kali. Semua tautan laporan lokal valid. |
| Jalur sukses berdasarkan laporan yang sudah tersedia | 49 Query dan 50 Mutation. Sebagian pengujian menggunakan data bisnis sintetis pada database disposable. |

Pengujian guard ini menggunakan request anonim. Pada laporan jalur sukses, pemberian role dan pemasangan permission dilakukan melalui GraphQL IAM. Tiga Query slip dan lima Mutation pada bagian dependency runtime masih menunggu bukti positif dari stack dengan adapter yang sesuai.

[Matriks Query](../../grpc/hrm/31-bukti-https-query-gateway.md) · [Matriks Mutation](../../grpc/hrm/30-bukti-https-mutation-gateway.md).
