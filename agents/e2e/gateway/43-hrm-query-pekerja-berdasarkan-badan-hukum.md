# Pengujian filter badan hukum pada daftar pekerja

| Kasus Personnel gRPC fixture | Hasil |
| --- | --- |
| Filter badan hukum valid | Pekerja yang cocok dengan badan hukum dan total dari HRM |
| RPC mengembalikan pekerja dari badan hukum berbeda | UNAVAILABLE |
| Nomor pekerja dan badan hukum dikirim bersamaan | INVALID_ARGUMENT |

Kasus filter dan regresi daftar tercakup dalam `TestHRMQueriesThroughGateway`. Scope tetap diperiksa oleh HRM. Verifikasi dengan grant IAM resmi melalui HTTPS masih terbuka.

Kedua replika Gateway lokal sehat setelah rebuild. `TestHRMWorkerFiltersRequireSessionOverHTTPS` lulus dengan race detector. Filter nomor pekerja serta badan hukum mengembalikan UNAUTHENTICATED dan data null melalui Caddy dengan CA development terverifikasi. [Tes berakun](34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md) masih SKIP karena fixture IAM resmi belum tersedia.

Regresi gRPC `TestWorkerScopedQueryRPCWithPostgreSQL` lulus ulang dengan race detector, PostgreSQL uji dan Redis TLS disposable. Policy HRM produksi membatasi query dan total menurut scope serta menolak sesi yang dicabut. Tes ini memakai proyeksi sesi fixture. [Matriks gRPC HRM](../human_resource/8-personnel-query-worker-legal-entity-scope.md).
