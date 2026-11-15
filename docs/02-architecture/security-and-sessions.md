# Authentication, Opaque Sessions, RBAC–ABAC dan CSRF

> **Baseline 12 November 2026.** Kebijakan keamanan/izin harus cocok dengan [agents/arsitektur.md](../../agents/arsitektur.md); ownership ditentukan oleh [boundary management](../../agents/boundary.md). Status runtime perlu dibuktikan dari kode/tes terbaru.

## Kebijakan browser target
**Baca keputusan owner langsung di [D01–D12](../../agents/arsitektur.md#21-keputusan-owner-yang-ditetapkan).** Browser memakai cookie `__Host-integrione_session` (Secure, HttpOnly, Path=/, host-only, SameSite=Lax) yang memuat handle opaque, **bukan JWT role/permission**. IAM pemilik lifecycle sesi dan state runtime keamanan Redis, dengan sumber otoritatif PostgreSQL.

| Pengaturan | Kebijakan |
|---|---|
| Idle | **30 menit** sejak activity sah yang diizinkan |
| Absolut | **8 jam**, tidak diperpanjang oleh touch |
| Kesegaran snapshot/context | **maksimum 30 detik** sejak pemeriksaan sumber, bukan TTL cookie |
| Revoke | `pending` sampai Redis fence dipublikasikan/terkonfirmasi; `effective` mewajibkan admission baru melihat revisi baru atau deny |
| Write sensitif | Recheck sesi, grant dan scope sebelum commit, tanpa klaim atomicity lintas Redis+DB |
| Dependency unavailable | **fail closed**, bukan fallback ALLOW |
| Redis failover | Admission gate + recovery epoch + reconcile sebelum trafik dibuka |

## Opaque vs legacy JWT (jangan disamakan)
**Legacy:** Gateway sebagian memakai cookie `erp_access` / `erp_refresh`, bearer JWT dan refresh-token family; umur access JWT **15 menit** dalam desain lama, bukan kebijakan sesi browser baru. Jalur CLI/API mungkin tetap perlu JWT sesuai kontrak sendiri.

**Status pilot (hasil audit 12 Nov):** IAM menyediakan BrowserSessionService, Gateway mempunyai HTTP lifecycle `/api/auth/browser/{begin,login,session,rotate,logout}` secara opt-in via `GATEWAY_AUTH_OPAQUE_BROWSER_ENABLED` (default false saat audit). Penanganan CSRF terikat sesi lewat IAM/Gateway; **tidak ada cookie CSRF kedua pada jalur opaque**. CI terarah lulus pada subset yang dicatat audit; **cutover production end-to-end belum terbukti**. Lihat [audit IAM–Gateway](../../agents/audit/iam_management/3-2026-11-12-164500-integrasi-opaque-csrf-reference-gateway.md), bukan halaman JWT 2026-11-01.

## Single endpoint untuk autentikasi — keputusan 13 November 2026

**Target normatif:** seluruh browser authentication berjalan sebagai operasi
GraphQL di **`/api/graphql`**, tidak boleh ada HTTP API auth tambahan.
Bootstrap pre-auth memberikan CSRF token + cookie opaque melalui GraphQL
dan header `Set-Cookie` (cookie tidak boleh bocor pada JSON/resolver).
Login adalah Mutation GraphQL yang memakai pre-auth handle dan CSRF yang
divalidasi IAM. Pemeriksaan sesi dan penerbitan CSRF berikutnya menggunakan
GraphQL Query; rotasi/logout memakai Mutation dengan Origin dan CSRF ketat.
Cookie sesi tetap `__Host-integrione_session`, bukan JWT di response.

Jalur bootstrap harus bisa diakses sebelum login tanpa bypass keamanan:
batasi origin/host, `Sec-Fetch-*` yang relevan, cache `no-store`, credential
ambiguity, dan replay. Operasi CSRF/bootstrap tidak boleh digabung dengan
field protected agar gagal authorization tidak membocorkan data. Mutasi
membutuhkan token yang valid dan session-bound; saat Redis/IAM tidak tersedia
wajib fail-closed. Perubahan grant dan revoke tetap mengikuti D01–D12.

**Implementasi belum lengkap:** Source historis masih mendaftarkan
`GET /api/auth/csrf` dan `/api/auth/browser/*` (opt-in), sedangkan
GraphQL `login/logout` masih menggunakan kontrak browser JWT legacy.
Ingress Caddy kini hanya memublikasikan `/api/graphql`; migrasi resolver,
gqlgen, middleware, Set-Cookie dan E2E harus dituntaskan sebelum handler REST
internal dihapus. Jangan mengklaim browser login opaque sudah dapat dipakai
hanya berdasarkan Caddy.

## Otorisasi per objek
IAM role/group/assignment/override menyuplai grant yang dibatasi action, effect, source, scope, validity, session activation, DENY dan conditions. Service bisnis memperoleh atribut objek dari database sendiri; customer account bukan principal ID dan vendor party tidak otomatis grant. Operasi `node(id:)`, list/count, nested, alias, bulk/export dan field PII memakai scope yang sama. IAM `superadmin` bukan wildcard semua management.

## Klasifikasi actor
PUBLIC (listing LIVE), GUEST_OWNED (cart own+CSRF), IAM_SELF, IAM_PROTECTED (staff+ABAC), WORKLOAD (authenticated identity/audience). Gateway dapat menolak awal, tetapi service pemilik menegakkan invariant. Bila katalog permission belum aktif, operasi protected ditolak; jangan seed speculative roles.

## Test gate
Browser HTTPS, Origin/CSRF, MFA/login/logout/rotate, 30m/8h/30s boundary with test clock, dual credential rejection, stale/revoke effective, Redis down, two principal/scope, signed reference gRPC, wrong-party Node, no state/outbox on DENY. Perlu test sungguhan untuk status VERIFIED.
