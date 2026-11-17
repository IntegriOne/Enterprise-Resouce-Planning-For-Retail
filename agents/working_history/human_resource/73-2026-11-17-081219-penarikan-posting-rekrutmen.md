# Penarikan posting rekrutmen

Menambahkan command/RPC `WithdrawJobPosting` dengan izin sensitif, alasan wajib, actor dan waktu server, serta outbox atomik. Migrasi `000003` dan DBML menegakkan metadata status withdrawn. Finder alasan/actor diberi paginasi. Tes unit/mock, gRPC–PostgreSQL disposable, migrasi up/down/up, fmt/vet/race/build, dan Docker healthy lulus. Konfirmasi kanal eksternal, grant IAM nyata, dan verifier DRM produksi masih perlu dikerjakan.
