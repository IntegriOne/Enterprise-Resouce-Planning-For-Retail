# Dokumentasi Integrione — ERP Retail + Website Commerce

**Status indeks:** 12 November 2026 · **Dokumentasi dibangun ulang dari nol di `docs/`** · `57` halaman di folder kategori sebelum README utama. Seluruh dokumen lama tetap dapat diaudit lewat [Git history](09-history/migration-map.md), tetapi bukan sumber kebijakan yang aktif.

## Mulai dari sini

0. **[Perspektif Produk Fase Development](development-product-perspektif.md)** — **wajib dibaca seluruh agen (terutama Frontend & UI/UX)**: menjelaskan bahwa kita membangun codebase *single-instance* untuk klien (diwakili brand **IntegriOne** saat development, bukan SaaS multitenant), serta bahwa `/login` adalah satu pintu masuk umum untuk **calon pembeli Web Commerce, internal (employee), dan vendor/supplier**.
1. **[Keputusan dan status](00-governance/decisions.md)** — sumber keputusan produk; IAM D01–D12 tetap di [agents/arsitektur.md](../agents/arsitektur.md).
2. **[Visi dan scope](01-product/vision-and-scope.md)** — 13 management, Manufacturing PAUSED dan urutan F1→F4.
3. **[Arsitektur sistem](02-architecture/system.md)** dan **[IAM/session](02-architecture/security-and-sessions.md)** — backend, cookie opaque, 30 menit idle, 8 jam absolut, 30 detik freshness. JWT browser 15 menit hanya legacy.
4. **[Peta management](03-managements/README.md)** — siapa menulis data PO, stok, harga, sales order, Finance, customer dan HRM.
5. **[Engineering & tests](04-engineering/README.md)** — kontrak, build, Go, E2E, Definition of Done.
6. **[Project & legal](06-project/README.md)** / **[Legal](08-legal/README.md)** — Charter, requirements, rencana, SOW/NDA/MSA, UAT, privasi, compliance.

## Scope dokumentasi management

Dokumentasi management yang berlaku hanya mencakup **13 management** pada [peta management](03-managements/README.md). Kedelapan halaman rinci di `docs/03-managements/` **mengelompokkan** beberapa management, bukan 8 management terpisah. Nama management di luar daftar 13 yang muncul dalam audit atau riwayat bukan target pengembangan aktif. Seluruh tautan kode menggunakan struktur repository saat ini: `backend/services/`, bukan `services/` di root Git.

## Urutan fase yang diputuskan owner

| Fase | Prioritas | Gerbang selesai |
|---|---|---|
| **F1 — Internal** | Staf, perusahaan, katalog SKU, supplier, PO, receiving/stock, Finance, IAM, HRM | E2E owner data, dua aktor positif/negatif, alur tidak ganda |
| **F2 — Vendor & supplier** | Wakil pihak eksternal, supplier PO/ASN, kontrak/dokumen, scoped IAM | PIC party A tidak dapat melihat B; grant/expiry/revoke auditable |
| **F3 — Marketplace & Analytics** | Integrasi Shopee/TikTok Shop dan lain-lain melalui API resmi; webhook, retry, reconciliation dan metrik produk | Provider approval/signature tervalidasi; OMS/WMS/Finance sumber kebenaran |
| **F4 — Customer/calon pembeli terakhir** | Website publik, varian, guest/cart, checkout/quote, OMS order, tracking, retur | IAM/CSRF, Pricing/OMS, Finance, privacy, UX dan E2E/UAT lulus |

**Manufacturing Operations PAUSED untuk pengembangan baru**; tidak termasuk 13 management aktif. Direktori service Manufacturing tidak ada pada tree `main` yang diperiksa 13 November 2026; keputusan mempertahankan riwayat tidak boleh ditafsirkan sebagai bukti implementasi aktif. Fitur berstatus `DESIGNED` bukan otomatis `IMPLEMENTED` atau `DEPLOYED`. Tersedianya SDL/protobuf tidak berarti API aktif.

## Peta navigasi lengkap

### 00 · Governance, keputusan dan sumber kebenaran (5)

- [Decisions](00-governance/decisions.md)
- [Documentation Policy](00-governance/documentation-policy.md)
- [Glossary](00-governance/glossary.md)
- [Ringkasan folder](00-governance/README.md)
- [Status And Evidence](00-governance/status-and-evidence.md)

### 01 · Produk retail, aktor, requirement dan roadmap (6)

- [Actors And Journeys](01-product/actors-and-journeys.md)
- [BRD PRD](01-product/brd-prd.md)
- [Ringkasan folder](01-product/README.md)
- [Requirements](01-product/requirements.md)
- [Roadmap](01-product/roadmap.md)
- [Vision And Scope](01-product/vision-and-scope.md)

### 02 · Arsitektur, IAM, integrasi dan data (7)

- [Deployment And Workspace](02-architecture/deployment-and-workspace.md)
- [Marketplace And Events](02-architecture/marketplace-and-events.md)
- [Product Analytics](02-architecture/product-analytics.md)
- [Ringkasan folder](02-architecture/README.md)
- [Security And Sessions](02-architecture/security-and-sessions.md)
- [Service Boundaries](02-architecture/service-boundaries.md)
- [System](02-architecture/system.md)

### 03 · 13 management dan batas domain (9)

- [Crm And Customer Service](03-managements/crm-and-customer-service.md)
- [Documents And Hrm](03-managements/documents-and-hrm.md)
- [Enterprise Foundation](03-managements/enterprise-foundation.md)
- [Finance](03-managements/finance.md)
- [IAM](03-managements/iam.md)
- [OMS And Commerce](03-managements/oms-and-commerce.md)
- [Procurement And Inventory](03-managements/procurement-and-inventory.md)
- [Product And Pricing](03-managements/product-and-pricing.md)
- [Ringkasan folder](03-managements/README.md)

### 04 · Developer, API dan quality (6)

- [API Contracts](04-engineering/api-contracts.md)
- [Coding And Collaboration](04-engineering/coding-and-collaboration.md)
- [Local Setup](04-engineering/local-setup.md)
- [Quality Gates](04-engineering/quality-gates.md)
- [Ringkasan folder](04-engineering/README.md)
- [Testing](04-engineering/testing.md)

### 05 · UX, UI dan riset (4)

- [Design System And Usability](05-experience/design-system-and-usability.md)
- [Personas And Scenarios](05-experience/personas-and-scenarios.md)
- [Ringkasan folder](05-experience/README.md)
- [Sitemap And Flows](05-experience/sitemap-and-flows.md)

### 06 · Delivery, Charter, BRD/PRD/SRS, kontrak dan QA/UAT (9)

- [Market Research And Poc](06-project/market-research-and-poc.md)
- [Project Charter](06-project/project-charter.md)
- [Project Plan](06-project/project-plan.md)
- [QA Results And UAT](06-project/qa-results-and-uat.md)
- [Ringkasan folder](06-project/README.md)
- [Requirements Traceability](06-project/requirements-traceability.md)
- [Risk And Change](06-project/risk-and-change.md)
- [Status And Closure](06-project/status-and-closure.md)
- [Templates And Contracts](06-project/templates-and-contracts.md)

### 07 · Release, runbook, DR dan user/support (5)

- [Backup And DR](07-operations/backup-and-dr.md)
- [Ringkasan folder](07-operations/README.md)
- [Release And Deployment](07-operations/release-and-deployment.md)
- [Runbooks And Incidents](07-operations/runbooks-and-incidents.md)
- [Support And User Guides](07-operations/support-and-user-guides.md)

### 08 · Legal/privacy/compliance (draft) (4)

- [Compliance Register](08-legal/compliance-register.md)
- [Contracts And Service Terms](08-legal/contracts-and-service-terms.md)
- [Privacy Terms And Dpa](08-legal/privacy-terms-and-dpa.md)
- [Ringkasan folder](08-legal/README.md)

### 09 · Peta migrasi dan arsip Git (2)

- [Migration Map](09-history/migration-map.md)
- [Ringkasan folder](09-history/README.md)

## Sumber normatif di luar docs

- [AGENTS.md](../AGENTS.md): aturan developer, commit ke `main`, instruksi per management dan pengujian.
- [agents/arsitektur.md](../agents/arsitektur.md): keputusan keamanan D01–D12, sesi browser opaque, RBAC–ABAC, Redis/revocation.
- [agents/boundary.md](../agents/boundary.md): ownership database/aggregate domain.
- [agents/users.md](../agents/users.md): objek `users` (`customer`, `employee`, `vendor`, `supplier`, `system`), identitas, sesi opaque, role/group/permission dan scope.
- [agents/instructions/README.md](../agents/instructions/README.md): handoff tugas developer.
- [Riwayat dokumen pra-rebuild](09-history/migration-map.md): hanya untuk audit historis, bukan implementasi.

## Aturan satu pernyataan

Jangan membuat kebijakan baru yang berkontradiksi dengan sumber normatif. Jika fakta runtime berubah, perbarui status berbasis source commit dan test pada [status/evidence](00-governance/status-and-evidence.md); jangan mengganti angka atau janji di 50 halaman sekaligus. API provider/kontrak, harga, SLO, UAT, dan dokumen hukum yang belum disetujui tetap **OPEN/DRAFT/BLOCKED**.
