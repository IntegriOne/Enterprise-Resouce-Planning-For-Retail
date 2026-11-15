# 7 — Integrasi Redis Security, Shared Auth Modules, Outbox/Reconciler, CSRF Deterministik, dan Bootstrap IAM

Menyelesaikan integrasi runtime keamanan Redis + PostgreSQL serta pembuktian internal gRPC untuk `backend/services/iam` dan `backend/shared`:

- **Infrastruktur & Shared Packages**: Menyiapkan sertifikat dev TLS + konfigurasi Redis (`noeviction` & ACL `iam_writer`/`gw_runtime`) di `backend/deploy/`, serta mengimplementasikan `backend/shared/{authcontext,authsession,authz}` pada `backend/go.work`.
- **Skema, Outbox & Reconciler IAM**: Menjaga baseline `000001`, menambahkan `RegisterAfterCommit` di `pgstore.Transactor`, repositori `outbox`, proyektor/koordinator/reconciler Redis (`includeScheduled=true` saat `BootstrapGateAndDrain`), serta `HMACCSRFDeriver`.
- **Registrasi, Verifikasi Email & Bootstrap**: Mengimplementasikan `RegisterCustomer`, `ConfirmEmailVerificationByToken` (termasuk aktivasi identitas `username`), `GetSessionCsrf`, `RefreshAccessSnapshot`, `BootstrapSystemPrincipalAndRBAC`, serta `InternalContextUnaryInterceptor` gRPC.
- **Docker & Konfigurasi**: Menyesuaikan `Dockerfile`/`docker-compose.yml` ke build context `backend/` dan memperbarui `.env`/`.env.example`.
- **Verifikasi**: Seluruh `backend/services/iam` (`go test -race -count=1 ./...`), uji live PostgreSQL 18 + TLS Redis 8.10.3 (`TestLivePGAndTLSRedis_OutboxBootstrapAndOutageRecovery`), seluruh `backend/shared/*`, dan pemeriksaan nol `.UTC()` lulus tanpa error.
