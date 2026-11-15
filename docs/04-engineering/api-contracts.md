# API GraphQL, gRPC dan Kontrak Lintas Domain

> **Pedoman engineering.** Ikuti [AGENTS.md](../../AGENTS.md), keputusan [arsitektur](../../agents/arsitektur.md), [ownership](../../agents/boundary.md) dan [scope](../01-product/vision-and-scope.md). Bukti tes terpisah dari rencana tes.

## Kontrak HTTP API tunggal — keputusan owner 13 November 2026

Seluruh HTTP API harus menggunakan tepat satu path: **`/api/graphql`**.
Gunakan Query untuk pembacaan/CSRF bootstrap dan Mutation untuk login, logout,
rotasi, recovery serta perubahan domain. Transport metadata (cookie HttpOnly,
Origin, header CSRF dan `Set-Cookie`) tetap dikelola pada HTTP boundary, bukan
GraphQL resolver bisnis. Tidak boleh ada REST path untuk auth, CSRF, health
atau management lain. Liveness melalui `query { health }`.

Penghapusan endpoint legacy hanya setelah kontrak GraphQL, generated gqlgen,
resolver, IAM RPC, dan negative/E2E test lolos. Caddy menolak `/api/*` selain
`/api/graphql` sejak keputusan ini. Pengecualian terbatas hanya disetujui untuk
jalur **inbound webhook server-to-server Payment Gateway resmi** (seperti
`/api/webhooks/xendit` dan kelak `/api/webhooks/midtrans`) saat fase F3/F4
diaktifkan; integrasi marketplace yang membutuhkan jalur callback sendiri tetap
harus melalui keputusan terpisah owner dan tidak boleh menambah REST auth/domain
API. gRPC privat tidak mengubah aturan endpoint HTTP publik.

## Registrasi adalah bukti exposure
GraphQL SDL yang tersimpan belum berarti aktif. **Daftar schema runtime** berasal dari `backend/gateway/gqlgen.yml`; generated gqlgen, resolver adapter, proto, registered gRPC service, DB dan E2E harus cocok. Contoh: Commerce channel Stage A punya SDL terpisah; seluruh Commerce B–H tidak otomatis aktif.

## Elemen kontrak
| Item | Isi minimum |
|---|---|
| Ownership | Owner service, actor/workload audience dan resource |
| Input | Field required/optional, ID type/source, currency/scale, expected revision |
| Response | State `PENDING/ACCEPTED/REJECTED/UNKNOWN` sesuai domain, source version, expiry |
| Auth | Permission `resource.action`, realm/legal entity/party scope, object attributes/SoD |
| Failure | INVALID, UNAUTHENTICATED, FORBIDDEN, NOT_FOUND, CONFLICT, EXPIRED, UNAVAILABLE dan business rejection |
| Distributed | Deadline, retry policy, key idempotency stabil, outbox/inbox, reconciliation status |
| Change | Protobuf field reservation/compat, schema version, regenerasi, consumer tests |
| Evidence | Source file, RPC, GraphQL operation, service/E2E test, commit dan env |

## GraphQL keamanan
PUBLIC/GUEST_OWNED/IAM_SELF/IAM_PROTECTED/WORKLOAD memerlukan klasifikasi endpoint; list, count, nested field, alias, fragments, multiple roots, global `node(id:)` dan pagination tetap terotorisasi. Response HTTP 200 dengan `errors[].extensions.code` mungkin berlaku setelah GraphQL execution; untuk error pre-execution, status HTTP dapat 401/403/503 sesuai boundary. Jangan menggunakan satu status untuk semua penyebab.

## Webhook
- **Customer Payment Gateway (Xendit / abstraksi siap Midtrans — Blueprint F3/F4):** Endpoint inbound (mis. `/api/webhooks/xendit`) diteruskan oleh Gateway ke `finance_accounting` via gRPC internal, **tanpa** memakai session cookie/CSRF browser. Karena `github.com/xendit/xendit-go/v7` hanya menyediakan struct callback (`InvoiceCallback`, `PaymentCallback`, `RefundCallback`) untuk `json.Unmarshal` tanpa helper verifikasi, handler **wajib** memvalidasi header `x-callback-token` dengan *constant-time comparison* terhadap secret vault sebelum menulis ke **durable webhook inbox** (`pg_webhook_inbox`), membalas HTTP 200 setelah persist durable, melakukan dedupe idempoten, memindah saldo clearing, lalu memancarkan outbox ke OMS/WMS/Commerce.
- **Marketplace Shopee/TikTok Shop:** Kontrak provider Shopee/TikTok Shop belum diverifikasi; tidak menulis fungsi signature atau skema event provider fiktif. Sumber dapat mengirim duplicate/out-of-order dan hasil command unknown setelah timeout. Inbound durable persistence sebelum ACK, dedupe per provider/shop/event key, replay safe, channel version mapping, dan reconcile ke OMS/WMS/Finance. Rujuk [integrasi](../02-architecture/marketplace-and-events.md).

## Consumer contract changes
Breaking change disampaikan ke owner/consumer dan dibuktikan pada tes; tidak mengubah management kedua tanpa mandat. Gunakan `agents/notes/{owner}/` dan instruksi resmi di `agents/instructions/{management}/{id}-{from}-{title-instructions}.md`.
