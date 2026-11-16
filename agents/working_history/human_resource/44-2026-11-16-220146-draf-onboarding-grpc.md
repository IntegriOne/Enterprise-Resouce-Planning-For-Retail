# Draf onboarding gRPC HRM

Validator plan/tugas onboarding kini aktif pada composition root. Worker, parent, kode unik, status draf, PIC worker, tanggal, dan larangan penyelesaian lewat CRUD diperiksa dalam transaksi. Unit mock, seluruh tes race HRM, vet, build, lima pengulangan gRPC–PostgreSQL, serta Docker healthy lulus. Tidak ada perubahan DBML/proto atau grant IAM. Lifecycle start/complete masih memerlukan validasi bukti DRM, relasi PIC, command dan E2E nyata; konteks HRM serta audit `28` memuat urutan lanjutan. Perubahan hanya pada service HRM dan catatan management.
