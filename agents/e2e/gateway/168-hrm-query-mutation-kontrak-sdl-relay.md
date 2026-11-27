# Verifikasi kontrak SDL dan guard HTTPS HRM

Skrip `backend/gateway/internal/e2e/scripts/check_hrm_schema_guards.py` memvalidasi SDL domain, lalu memakai CA lokal untuk menguji Gateway melalui HTTPS.

| Kasus | Hasil |
| --- | --- |
| SDL aktif | 24 tipe Node dan 24 Connection sesuai kontrak. Semua Mutation memiliki payload dengan `MutationResult`. |
| Variasi SDL sintetis dengan nama operasi mengandung `HrM` | Pemeriksaan menolak nama. |
| Variasi SDL dengan `Node.id` nullable atau antarmuka Node dihapus | Pemeriksaan menolak perubahan. |
| Variasi SDL dengan edge cursor nullable atau argumen `last` dihapus | Pemeriksaan menolak perubahan. |
| Variasi SDL dengan payload tanpa `MutationResult` | Pemeriksaan menolak perubahan. |
| HTTPS anonim | Seluruh 52 Query ditolak saat sesi tidak ada. Seluruh 55 Mutation ditolak saat header CSRF tidak ada. |

Variasi negatif dieksekusi dalam memori. Tautan ke bukti jalur bisnis positif ada di [matriks Query](../../grpc/hrm/31-bukti-https-query-gateway.md) dan [matriks Mutation](../../grpc/hrm/30-bukti-https-mutation-gateway.md).
