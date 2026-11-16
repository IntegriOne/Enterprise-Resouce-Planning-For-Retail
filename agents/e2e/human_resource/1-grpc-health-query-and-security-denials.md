# E2E terbatas gRPC HRM — health dan penolakan keamanan

Tanggal: 16 November 2026. Lingkungan: Docker Compose development, service `human-resource` privat pada `127.0.0.1:9003`, PostgreSQL dan Redis TLS container. Ini uji transport nyata, belum E2E GraphQL atau CRUD positif.

| Case | Jalur / input | Hasil aktual |
| --- | --- | --- |
| Health setelah startup | `docker compose up -d --no-deps --wait human-resource` dan probe `grpc.health.v1.Health/Check` dari container | `healthy`, RPC `OK` tercatat di log. |
| Sertifikat klien hilang | `grpcurl` ke RPC `PersonnelService/GetAllWorkers` tanpa client cert | Handshake ditolak: `tls: certificate required`. |
| Pemanggil tanpa bukti Gateway | RPC yang sama dengan sertifikat klien sah tetapi tanpa HMAC/context | `Unauthenticated`, tanpa data pekerja. |
| Migrasi startup | Query tracker `hrm_schema_migrations` | Versi `1`, checksum panjang 64. |
| Restart idempoten | `docker compose restart human-resource`, lalu `up -d --no-deps --wait human-resource` | Kembali `healthy`; tracker tetap satu baris versi `1`. |
| Dependency bersama | `docker compose ps` sesudah startup | Redis, PostgreSQL, IAM, Gateway, frontend dan Enterprise Foundation tetap sehat. |

Belum diuji: caller dengan sesi IAM dan grant sah, scope objek, mutasi bisnis, outbox, login GraphQL, timeout Redis nyata dan seluruh 84 resource. Katalog permission masih draf dan Integrity runtime menolak write tanpa validator, sehingga penolakan di sini tidak membuktikan jalur bisnis sukses.
