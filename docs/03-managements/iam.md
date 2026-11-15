# IAM Management — Internal, Supplier, Vendor dan Customer

> **Dokumen management retail · 12 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## Ownership
IAM menyimpan `principals` human/workload, identifier, credential/MFA, autentikasi, sesi opaque, groups, roles, permissions, assignments, overrides, delegations, scopes, revocation, audit. Management lain mengajukan permintaan akses; keputusan independen dan grant efektif tetap IAM. Role IAM `superadmin` tidak otomatis memiliki hak Finance, Commerce atau WMS.

## Alur manusia
- **Internal staff:** IAM login, role sesuai pekerjaan; unit/cabang from Foundation, worker from HRM. Status HRM/kontrak tidak langsung membuat grant.
- **Supplier PIC:** personal principal IAM berbeda dari party supplier Foundation; Procurement membuktikan PO milik party yang aktif. Vendor PIC jasa memiliki kontrak/scope sendiri.
- **Customer/calon pembeli:** public allowlist vs guest session/ownership vs IAM login customer. OIDC provider harus memverifikasi issuer/audience/state/nonce/PKCE sebelum mengikat `provider_subject`; email tidak cukup untuk merge principal.
- **Workload:** mTLS/identity dan audience; tidak meminjam cookie/browser principal.

## Access evaluation
Sesi aktif/generation, auth epoch/revision fresh, permission `resource.action`, effect ALLOW/DENY, scope UUID, conditions/validity, MFA assurance, atribut object dari owner. No grant → deny. Dependency unavailable → unavailable/fail closed. List/query/Node/exports juga terikat scope.

## Implementasi dan blocker
Sesi opaque dan snapshot IAM mempunyai source/test; Gateway pilot HTTP opt-in; cutover umum belum disebut VERIFIED (cek bukti baru). IAM audience Commerce dan grant bisnis memerlukan review khusus, bukan seed spekulatif. Keputusan detail idle 30m, absolute 8h, freshness 30s tetap pada [policy](../02-architecture/security-and-sessions.md).

## Kontrak GraphQL dan matriks izin

Pemetaan **38 Query dan 82 Mutation IAM** ke RPC serta permission katalog diperinci pada [matriks akses seluruh GraphQL IAM](../../agents/roles-permissions/iam_management/3-graphql-access-matrix.md). Kategori publik, self, admin dan requester dibatasi oleh enforcement service. Kelulusan mapping statis tidak sama dengan kelulusan E2E semua jalur.

## Sumber aktual
[Service IAM](../../backend/services/iam_management/README.md), [model principal](../../agents/principals.md), [audit Gateway](../../agents/audit/iam_management/3-2026-11-12-164500-integrasi-opaque-csrf-reference-gateway.md).
