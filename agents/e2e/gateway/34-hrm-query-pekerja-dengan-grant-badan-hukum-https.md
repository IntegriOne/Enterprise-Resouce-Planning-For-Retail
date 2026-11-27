# Pengujian HTTPS pekerja dengan grant badan hukum

`TestHRMWorkerLegalEntityGrantOverHTTPS` memakai login IAM melalui Caddy dan sesi browser asli. Sediakan akun uji yang dapat menyelesaikan login password, grant `human_resource.workers.read` untuk satu badan hukum, serta dua worker dari badan hukum berbeda.

| Variabel | Nilai fixture |
| --- | --- |
| `E2E_HRM_EMPLOYEE_IDENTIFIER` dan `E2E_HRM_EMPLOYEE_PASSWORD` | Kredensial akun uji yang berizin |
| `E2E_HRM_LEGAL_ENTITY_ID` | Badan hukum dari grant |
| `E2E_HRM_ALLOWED_WORKER_ID` | UUID pekerja dalam badan hukum |
| `E2E_HRM_DENIED_WORKER_ID` | UUID pekerja yang ada di badan hukum lain |
| `E2E_HRM_DENIED_LEGAL_ENTITY_ID` | UUID badan hukum pekerja pembanding di luar grant |

Dari modul Gateway, jalankan `go test -tags=e2e ./internal/e2e -run '^TestHRMWorkerLegalEntityGrantOverHTTPS$' -count=1 -v`. Tes meminta detail yang diizinkan, memastikan detail di luar cakupan memberi NOT_FOUND, dan memeriksa cakupan item pada halaman daftar. Tes filter meminta badan hukum dalam grant dan badan hukum di luar grant. Permintaan di luar grant harus mengembalikan daftar kosong dan total nol.

Kompilasi paket ber-tag dan vet lulus. Skenario berakun masih SKIP karena fixture IAM resmi belum tersedia. Tes anonim payroll dalam paket yang sama lulus.

## Verifikasi ulang pada rollout detail tugas offboarding

Suite HTTPS HRM lulus untuk guard cuti, payroll, filter pekerja, kontrak, assignment, dan posisi. Percobaan yang berdekatan dengan pemeriksaan seluruh schema sempat mencapai limit 60 mutation per 60 detik dan menerima `TOO_MANY_REQUESTS`. Suite lulus setelah jendela limit berakhir. Skenario grant badan hukum masih SKIP karena fixture resmi belum tersedia. Suite HRM `go test ./...` serta suite race untuk server Gateway dan adapter HRM juga lulus.

## Verifikasi scope pada gRPC HRM dan PostgreSQL

`TestWorkerScopedQueryRPCWithPostgreSQL` lulus dengan tag `integration` pada PostgreSQL Docker dan database sementara khusus `*_test`. Tes memeriksa detail dan paginasi pekerja lintas badan hukum, penggantian scope sesi, serta penolakan sesi yang tidak sah. Database sementara telah dihapus. Tes menggunakan proyeksi sesi fixture di sisi HRM. Bukti login IAM melalui Caddy masih memerlukan manifest dan assignment IAM resmi. [Kontrak handoff IAM](../../notes/iam_management/1-human_resource-permission-handoff.md).

## Uji HTTPS dengan assignment resmi

`go test -tags=e2e ./internal/e2e -run '^TestHRMWorkerLegalEntityGrantOverHTTPS$' -count=1 -v` lulus. Dua badan hukum aktif, dua party, dan dua pekerja dibuat melalui API resmi di database Enterprise Foundation serta HRM disposable. Grant HRM global hanya dipakai saat setup, lalu assignment global dan izin membuat pekerja dicabut. Role uji dengan `human_resource.workers.read` diberikan kembali ke akun superadmin berjenis employee dengan scope `LEGAL_ENTITY` pada badan hukum A. Akun tetap memiliki izin baseline IAM, sementara akses baca pekerja HRM hanya berasal dari assignment berscope.

| Kasus | Hasil |
| --- | --- |
| Detail pekerja dalam badan hukum A | Berhasil dan mengembalikan ID serta badan hukum yang sesuai |
| Detail pekerja dari badan hukum B | `NOT_FOUND` |
| Daftar tanpa filter | Hanya memuat pekerja dalam scope |
| Filter badan hukum A | Mengembalikan pekerja dalam scope |
| Filter badan hukum B | Daftar kosong, total nol |

Assignment, role, dan sesi uji dibersihkan. Kedua database disposable dihapus setelah service kembali sehat pada konfigurasi development semula.

## Relay node melalui HTTPS

Dengan fixture dan role berscope yang sama, `node` menerima GID `gid://integrione/hrm-worker/<uuid>` pekerja dalam badan hukum A dan mengembalikan identitas yang sesuai. GID pekerja badan hukum B menghasilkan `NOT_FOUND`. GID dengan resource `worker` menghasilkan `INVALID_ARGUMENT`. Pengulangan sempat terkena rate limit saat setup. Assignment global pengulangan sudah dicabut sebelum limit tercapai, dan role sementara yang tersisa dihapus melalui IAM setelah jendela limit berakhir. Service development kembali sehat.
