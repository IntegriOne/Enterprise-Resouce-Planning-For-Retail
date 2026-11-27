# Laporan E2E: Decoupling Gateway dari IAM, Redis, dan Antar Replika

- **Tanggal:** 16 November 2026 (06:53 WIB)
- **Lingkungan:** `https://erp.localhost:8443/api/graphql` (`9` container Compose aktif: `caddy`, `frontend-1`, `frontend-2`, `gateway-1`, `gateway-2`, `iam`, `enterprise-foundation`, `postgres`, `redis`)
- **Suite Dijalankan:** `GOWORK=off /opt/go/bin/go test -tags=e2e -v -count=1 -timeout 12m ./internal/e2e/...`
- **Ringkasan Hasil Penuh:** **`82`** top-level test PASS, **`156`** subtest PASS (`149` level-1 + `7` level-2), **`0` fail**, **`0` skip**, durasi **`134.331s`**.

## Matriks Kasus Decoupling Gateway (`gateway_decoupling_e2e_test.go`)

1. **`TestGateway_BootsWithoutIAMAndRedisThenAutoRecovers` (`22.87s`):**
   - **Startup Tanpa IAM & Redis:** Container `iam` dan `redis` dihentikan (`docker compose stop iam redis`), lalu `gateway-1` dan `gateway-2` di-boot ulang via `docker compose up -d --wait --no-build --force-recreate` tanpa `--no-deps`. Kedua replika tetap mencapai status `running:healthy`.
   - **Health & Fail-Closed Tanpa Kebocoran:** Port langsung `gateway-1` (`127.0.0.1:5000`) dan `gateway-2` (`127.0.0.1:5001`) merespons `query { health }` -> `"ok"`. Query `browserSession` dan mutasi `userLogin` (4 kali beruntun via Caddy HTTPS) gagal tertutup (`HTTP 200` dengan `extensions.code = "UNAVAILABLE"`) tanpa `INTERNAL_ERROR` maupun kebocoran detail internal (`iam:9001`, `redis:6379`, `127.0.0.1`, `dial tcp`, `connection refused`, `postgres`, `sqlstate`, `pq:`).
   - **Pemulihan Bertahap Redis lalu IAM:** Saat `redis` dinyalakan (`iam` masih mati), `browserSession` anonim pulih otomatis tanpa restart Gateway sementara `userLogin` mengembalikan `MutationResult { was_successful: false, errors: [{ type: UNAVAILABLE }] }`. Setelah `iam` dinyalakan, login SuperAdmin dan operasi GraphQL pulih otomatis.
2. **`TestGateway_ReplicaPeerAndCaddyFailoverIndependence` (`19.60s`):**
   - **Independensi `gateway-2` dari `gateway-1`:** Saat `gateway-1` dihentikan, `gateway-2` dan `caddy` dapat dibuat ulang dan berjalan tanpa `--no-deps`; Caddy HTTPS melayani `query { health me { id } users(first: 2) { total_count } }` (`6` kali beruntun OK).
   - **Failover Dua Arah:** `gateway-1` dinyalakan kembali dan `gateway-2` dihentikan; Caddy otomatis memindah trafik ke `gateway-1` (`6` kali beruntun OK), lalu kedua replika pulih bersama.
3. **`TestGateway_RuntimeIAMOutageFaultIsolationAndAutoRecovery` (`11.75s`):**
   - **Isolasi Gangguan IAM saat Runtime:** Dengan snapshot Redis segar untuk pembaca EF dan sesi *stale* (>10 menit) bersnapshot hangat untuk penguji sensitif, `iam` dihentikan saat runtime (`gateway-1`, `gateway-2`, `redis` tetap `healthy`).
   - **Blast Radius Terbatas & Pemulihan Otomatis (`6` iterasi berturut-turut):** `health` & query `allAddresses(first: 1)` tetap berhasil; `users` dan `me` gagal tertutup `UNAVAILABLE`; `userLogin` dan mutasi sensitif `createRole` mengembalikan `MutationResult` error `UNAVAILABLE` tanpa dimasking menjadi `AUTHORIZATION` (`FORBIDDEN`); cookie sesi admin tidak tercabut; dan seluruh query IAM pulih otomatis setelah `iam` dinyalakan tanpa restart Gateway.

## Verifikasi Regresi, Kebersihan DB, & Uji Unit

- Seluruh suite **Enterprise Foundation** (`3` test: lifecycle nested + RBAC scope + fault isolation) dan **IAM** (`76` top-level test / `156` total subtest) lulus penuh (`0` fail, `0` skip).
- Pemeriksaan PostgreSQL pasca-suite: `15` izin baseline pada `iam.permissions` dan `0|0|0|0|0` baris tersisa pada `addresses`, `legal_entities`, `locations`, `organization_units`, serta `parties`.
- Validasi Caddy (`caddy validate --config /etc/caddy/Caddyfile`) -> `Valid configuration`; verifikasi Go (`go vet -tags=e2e ./...`, `go test -race -count=1 ./...`) lulus di seluruh `8` paket teruji Gateway.
