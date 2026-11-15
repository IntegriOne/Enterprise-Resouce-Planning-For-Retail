# Migrasi Dokumentasi — 12 November 2026

**Seluruh isi direktori `docs/` lama telah diganti dengan struktur yang dibuat ulang.** Git menjaga histori: [dokumentasi sebelum rebuild](https://github.com/hackerenesia/erp/tree/a78bb6651c04fe5908d260b05e83238221766602/docs). Arsip itu **bukan sumber kebijakan yang berlaku**, terutama pada klaim JWT 15 menit, tidak ada idle timeout, tabel IAM lama, jumlah module lawas dan path `back-end/`.

## Pemetaan dokumen lama ke struktur baru
| Lokasi lama | Sumber baru |
|---|---|
| `docs/index.md` | [README docs](../README.md) (index.md kini hanya penunjuk) |
| `docs/architecture/{overview,inter-service-boundary}.md` | [System](../02-architecture/system.md) dan [Ownership](../02-architecture/service-boundaries.md) |
| `docs/architecture/{raas-business-model,workspace-and-modules}.md` | [Deployment & Workspace](../02-architecture/deployment-and-workspace.md) |
| `docs/gateway/*` | [Security](../02-architecture/security-and-sessions.md) dan [API](../04-engineering/api-contracts.md) |
| `docs/managements/iam_management/*` | [IAM](../03-managements/iam.md), [Security](../02-architecture/security-and-sessions.md), [Evidence](../00-governance/status-and-evidence.md) |
| `docs/managements/enterprise_foundation/*` | [Enterprise Foundation](../03-managements/enterprise-foundation.md) |
| `docs/managements/planned_managements.md` | [Management map](../03-managements/README.md) |
| `docs/standards/*`, `docs/developer-guide/*` | [Engineering](../04-engineering/README.md) |
| `docs/plans/*` | [Charter](../06-project/project-charter.md) dan [History](README.md) |
| `docs/retail-commerce/01-initiation..03-requirements/*` | [Product](../01-product/README.md) dan [Project](../06-project/README.md) |
| `docs/retail-commerce/04-ux/*` | [Experience](../05-experience/README.md) |
| `docs/retail-commerce/05-architecture/*` | [Architecture](../02-architecture/README.md) |
| `docs/retail-commerce/06-quality/*` | [Testing](../04-engineering/testing.md) dan [QA/UAT](../06-project/qa-results-and-uat.md) |
| `docs/retail-commerce/07-release..09-monitoring/*` | [Operations](../07-operations/README.md) dan [Project status](../06-project/status-and-closure.md) |
| `docs/retail-commerce/10-standards/*` | [Engineering](../04-engineering/README.md) |
| `docs/retail-commerce/11-legal/*` | [Legal](../08-legal/README.md) |

## Prinsip tidak menghilangkan bukti
Audit dan working history tetap ada di `agents/` dan source; riwayat Git mengembalikan halaman Markdown lama jika memang perlu membaca konteks waktu itu. Namun seluruh pekerjaan baru wajib merujuk dokumen **kanonik** `docs/README.md` dan sumber normatif `AGENTS.md`, `agents/arsitektur.md`, `agents/boundary.md`.
