# 17. Decoupling Gateway terhadap IAM, Redis, dan Antar Replika

**Tanggal:** 16 November 2026 (06:53 WIB)

Memisahkan ketergantungan startup serta mengisolasi kegagalan runtime antara `gateway-1`, `gateway-2`, `iam`, `redis`, dan `caddy`:

- **Compose & Caddy ([backend/docker-compose.yml](file:///home/muriarfad/Documents/projects/erp-retail/backend/docker-compose.yml), [proxy/docker-compose.yml](file:///home/muriarfad/Documents/projects/erp-retail/proxy/docker-compose.yml), [proxy/Caddyfile](file:///home/muriarfad/Documents/projects/erp-retail/proxy/Caddyfile)):** Menghapus `depends_on` Gateway (`iam`, `redis`, antar replika) dan Caddy; menambahkan active/passive health check serta `lb_retries 1` failover.
- **Ketahanan Gateway ([resilience.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/source_connections/resilience.go), [iam_client.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/source_connections/iam_client.go), [authorizer.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/authorizer.go)):** Menyelaraskan keepalive klien (`Time: 60s`), migrasi wrapper `callIAM*` penuh pada adapter `user_admin` dan `rbac`, serta menembuskan error `UNAVAILABLE` secara fail-closed tanpa memutus sesi aktif.
- **Verifikasi:** Seluruh `8` paket unit Gateway lulus `-race`, `caddy validate` valid, dan E2E penuh [gateway_decoupling_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/gateway_decoupling_e2e_test.go) mencatat `82` test / `156` subtest PASS (`0` fail, `0` skip, `134.331s`).
