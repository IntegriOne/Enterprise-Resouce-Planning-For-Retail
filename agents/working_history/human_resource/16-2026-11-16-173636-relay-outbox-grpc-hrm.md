# Relay outbox gRPC HRM

Relay outbox dibangun untuk skema HRM: klaim `SKIP LOCKED`, retry eksponensial, batas percobaan, pemeriksaan digest, dan envelope webhook HMAC dengan kunci deduplikasi event ID. Bootstrap menjalankannya bersama gRPC; mode bawaan standby menjaga pending. Mode log hanya development. Published/dead letter tidak dipurge karena outbox masih sumber versi agregat. Unit/race/vet, tes PostgreSQL disposable dua claimant, build Docker dan health lulus. Consumer nyata, scope parsial, validator 76 resource mutabel, dan E2E positif masih belum selesai.
