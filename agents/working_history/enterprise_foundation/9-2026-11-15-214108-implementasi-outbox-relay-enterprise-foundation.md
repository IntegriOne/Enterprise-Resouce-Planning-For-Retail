# Implementasi Outbox Relay Asinkron Enterprise Foundation

Mengimplementasikan worker relay asinkron pada [infrastructure/outbox](../../../backend/services/enterprise_foundation/infrastructure/outbox/relay.go) untuk memproses `foundation_outbox_events` secara aman di PostgreSQL dan mengintegrasikannya ke proses gRPC server.

## Perubahan Utama

- **Penerbit & Klasifikasi Error:** Menambah `Publisher`, `Envelope`, `WebhookPublisher` (HMAC-SHA256 + anti-redirect, klasifikasi 4xx permanen vs 408/429/5xx retryable), `RouterPublisher`, `MultiPublisher`, `LogPublisher` non-produksi, `BackoffPolicy`, dan `PermanentError`.
- **Relay & Wakeup Pasca-Commit:** Mengimplementasikan `Relay` (`Run`, `DrainPending`, `DrainBatch` dalam transaksi `SKIP LOCKED`, `CleanupEffective`, `Stats`, shutdown aman tanpa membakar `attempt_count`) serta `WrapNotifyingRepository` berbasis `pgstore.RegisterAfterCommit`.
- **Konfigurasi & Runtime Docker:** Menambah pengaturan `EF_OUTBOX_*` di [config.go](../../../backend/services/enterprise_foundation/config/config.go) (default `standby` aman, melarang `log` di `production`), serta wiring di [wiring.go](../../../backend/services/enterprise_foundation/cmd/wiring.go) dan [main.go](../../../backend/services/enterprise_foundation/cmd/main.go).

## Verifikasi

- `go vet ./...` dan `go test -p 1 -race -count=1 ./...` pada `enterprise_foundation` (termasuk uji PostgreSQL langsung & `TestBuildDependencies_WithLiveDatabase`) lulus penuh.
- Container `enterprise-foundation` dibangun ulang (`Healthy`) dan smoke test gRPC live `127.0.0.1:9002` (Health, auth fail-closed, 5 service) berjalan bersih tanpa error.
