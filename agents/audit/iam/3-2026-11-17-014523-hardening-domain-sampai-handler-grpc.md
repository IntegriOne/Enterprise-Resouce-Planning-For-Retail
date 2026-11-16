# Audit dan perbaikan IAM gRPC: domain sampai handler

Tanggal: 17 November 2026 (WIB). Cakupan: domain entity/error/repository, implementasi repository terkait, service user/auth, role, permission, group, session, bootstrap, serta handler dan interceptor gRPC. Pemeriksaan berupa penelusuran kode, unit/mocks, race test, generator proto, build image, dan smoke live. Matriks E2E autentikasi/RBAC/MFA belum dijalankan penuh.

## Temuan yang diperbaiki

1. **Tinggi — katalog permission dapat bertabrakan lintas management.** Bootstrap membuang namespace dari `resource`; sekarang parser menjaga `iam.users` untuk `iam.users.read`. `NewPermission` dan mapper penyimpanan menolak `code/resource/action` yang tidak kanonik. Bootstrap memperbaiki metadata lama saat startup dan berhenti jika `is_sensitive` berbeda dari katalog yang disetujui; perubahan sensitivitas memerlukan penanganan terkontrol agar cache otorisasi tidak diam-diam menjadi longgar.
2. **Tinggi — pencabutan permission dan outbox sesi dapat meninggalkan akses lama.** Delete permission kini mengunci baris, menolak jika masih terikat pada role, dan mengembalikan konflik domain; penghapusan tidak lagi mengandalkan cascade yang melewati bump revisi izin. Issue/update/revoke sesi kini menggagalkan transaksi bila pencatatan outbox gagal. Refresh snapshot juga gagal tertutup jika recovery epoch tidak dapat dibaca.
3. **Tinggi — jalur gRPC dapat menaikkan status MFA tanpa kode.** `MarkSessionMfaVerified` sekarang selalu mengembalikan `FailedPrecondition` dan didepresiasi di proto. Jalur sah ialah `AuthService.VerifySessionMfa`. `IssueSession` menolak caller anonim, cross-user, dan klaim `mfa_verified=true`; login internal user service tetap dapat menerbitkan sesi setelah verifikasi kredensial/MFA.
4. **Sedang — sesi pemilik nonaktif masih lolos operasi sesi tertentu.** Authenticate, VerifyCSRF, MarkMFAVerified, dan GetSessionCSRF kini memeriksa status pemilik sesi; pemilik yang hilang/nonaktif mendapat `Unauthorized`.
5. **Sedang — actor assignment berasal dari payload.** `GrantedBy` pada user role dan group role kini diambil dari caller internal yang terverifikasi; nilai request lama diabaikan. `group_members.added_by` sekarang terisi dari caller yang sama bila tersedia. Kontrak proto menjelaskan field lama yang diabaikan.
6. **Sedang — proof internal opsional dalam development.** Runtime development dan production menuntut HMAC minimal 32 byte; interceptor menolak business RPC tanpa proof dengan `Unauthenticated`. Mode `test` mempertahankan fixture bufconn tanpa secret.

## Bukti verifikasi

- `go generate ./...` pada kontrak IAM: lulus (`buf lint`, `buf generate`).
- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`: lulus pada modul IAM. Tes baru mencakup kegagalan outbox, pemilik sesi nonaktif, recovery epoch, permission terikat role, actor tepercaya, dan batas RPC MFA/sesi.
- `docker compose build iam` dan `docker compose up -d --wait --no-deps iam`: lulus; container IAM baru `healthy` tanpa restart dependency lain.
- Smoke live: `grpc.health.v1.Health/Check` mengembalikan `SERVING`; `GetUserById` tanpa proof internal mengembalikan `Unauthenticated`; query publik `__typename` melalui HTTPS `/api/graphql` berhasil.
- E2E Gateway/IAM `TestMutationUserLogin` lulus. Matriks `Test(Query|Mutation)` gagal hanya pada `TestQueryPermissions/FilterByResourceSensitiveAndPagination` karena fixture Gateway masih meminta `resource="users"`, sedangkan katalog kini menggunakan `iam.users`. Matriks yang sama dengan `TestQueryPermissions` dilewati lulus. Setelah rebuild terakhir, E2E login, assignment user role, penambahan anggota grup, dan assignment group role lulus. Detail cases dicatat di `agents/e2e/iam/1-permission-query-dan-auth-mutation-gateway.md`.

## Batas audit lanjutan

IAM tetap hanya memvalidasi format dan jenis `scope_id`; eksistensi/kelayakan target `legal_entity`, `facility`, `organization_unit`, `supplier_party`, dan `vendor_party` memerlukan pembuktian owner Enterprise Foundation pada jalur assignment. Perubahan pada Gateway/Enterprise Foundation yang sedang berlangsung di working tree lain tidak disentuh. Tes filter resource Gateway perlu diselaraskan oleh pemilik Gateway dengan namespace katalog IAM. Review ini tidak menjamin ketiadaan semua cacat pada setiap cabang user/auth; audit kontrak scope masih diperlukan sebelum klaim selesai penuh.
