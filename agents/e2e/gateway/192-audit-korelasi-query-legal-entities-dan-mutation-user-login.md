# Verifikasi saat audit korelasi 13 management

28 November 2026 WIB, checkout lokal `e120e679` beserta perubahan working tree. Endpoint yang diperiksa adalah `https://erp.localhost:8443/api/graphql` melalui Caddy dengan CA lokal tepercaya. Identitas, kredensial dan isi data perusahaan tidak disalin.

## Pemeriksaan HTTPS aktual

Sesi superadmin dibuka melalui `browserSession` dan `userLogin`, menggunakan cookie dari Gateway, Origin yang sesuai serta token CSRF terbaru. Pengujian baca memakai grant yang sudah ada. Tidak ada assignment role, perubahan izin, SQL IAM atau transaksi stok/uang yang dijalankan oleh probe ini.

| Operasi | Hasil yang diamati | Batas kesimpulan |
|---|---|---|
| `health` | `ok` | Menjangkau Gateway. |
| `userLogin`, `permissions(first:1)`, `userLogout` | AUTHENTICATED, data, logout sukses. Login/logout pada probe kedua juga sukses. | Sesi dan pembacaan katalog IAM. |
| `legalEntities(first:1)` | Pertama UNAVAILABLE; log Gateway pada jendela yang sama mencatat deadline exceeded. Dua pembacaan ulang dengan sesi baru menghasilkan data dalam sekitar 0,01 dan 0,02 detik. | Gangguan teramati sekali. Akar timeout dan replika yang menerima setiap retry belum dipastikan. |
| `myPublishedDocuments(limit:1)` | Respons data pada sesi terautentikasi. | Jalur daftar DRM; isi PDF dan otorisasi transaksi owner tidak diuji oleh pembacaan ini. |
| `workers(first:1)`, `productCategories(limit:1)`, `stockReservations` | Masing-masing FORBIDDEN pada grant sesi ini. | Bukti penolakan akses. Bukan bukti transaksi bisnis positif atau kegagalan service. |

`docker compose ps` melaporkan kesembilan service management dan kedua Gateway sehat saat pemeriksaan. Pemeriksaan allowlist konfigurasi container mendapati relay WMS, PLM dan Commerce aktif; validasi EF pada WMS/PLM/Commerce/DRM serta validasi PLM pada Commerce berstatus false di stack development ini. `FINANCE_OUTBOX_MODE=standby` teramati; periksa konfigurasi route khusus sebelum menyimpulkan semua jalur Finance berhenti. Peer target kesembilan service terdaftar pada kedua Gateway. Kesehatan proses tidak menguji isi event atau ACK owner.

## Pemeriksaan Go aktual

Dari setiap modul pada `backend/services/`, jalankan `GOWORK=off go test -json -race -count=1 <paket>`. Hitungan berasal dari event JSON test bernama, termasuk subtest. Pada lintasan awal, satu test PLM, satu WMS dan satu Procurement skip karena DSN belum diset. Ketiganya diulang memakai konfigurasi role dari env lokal dan PostgreSQL Compose khusus DB bernama `product_lifecycle_test`, `inventory_warehouse_test`, `procurement_sourcing_test`. Tidak ada DSN/kredensial di laporan.

| Modul | Paket | PASS pada lintasan terakhir | Cakupan |
|---|---|---:|---|
| iam | `./infrastructure/scopeverifier` | 31 | Tes verifier scope EF yang sudah ada, termasuk subtest. |
| enterprise_foundation | `./services/party/test` | 13 | Aturan service party dan kegagalan dependency. |
| human_resource | `./internal/service/payroll` | 13 | Publikasi, snapshot dan konsistensi hasil HR. |
| finance_accounting | `./internal/service/inventory_cost/test` | 1 | Moving average dan guard nilai negatif pada Finance. |
| procurement_sourcing | `./internal/service/test` | 1 | Use case procurement pada DB uji dengan dependency pilihan fixture. |
| inventory_warehouse | `./internal/service` | 7 | Acceptance WMS dan subtest pada DB uji. |
| product_lifecycle | `./services/masterdata/test` | 1 | Matriks lifecycle pada DB uji; acceptance memasang fake EF, Inventory, Procurement dan Publisher. |
| web_commerce | `./services/storefront/test` | 1 | Matriks domain Commerce pada harness service. |
| document_record | `./services/records/test` | 1 | Matriks DRM dengan storage/DB pengujian. |
| **Total** | **9 paket** | **69** | **0 fail, 0 skip pada lintasan terakhir.** |

## Bukti yang masih harus dibuat agen

Butuhkan correlation yang sama sampai ke penerima, effect ID dan perubahan keadaan nyata untuk receipt → valuasi/jurnal yang tepat, PO → accepted/reversal → outstanding, rilis SKU → stale listing/quote, publication payroll → PDF worker per hasil, serta hold → purge dengan restart. Periksa fixture dan produsen data. Sebagian suite Gateway lama membaca role lewat SQL dan mengubah `users.user_kind` saat membuat aktor uji. Gunakan provisioning resmi dan role/permission melalui GraphQL saat mengulang perjalanan lintas owner. [Audit dan daftar instruksi](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md).
