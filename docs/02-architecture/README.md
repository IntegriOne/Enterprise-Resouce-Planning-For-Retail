# Arsitektur Sistem — Peta

> **Baseline 12 November 2026.** Kebijakan keamanan/izin harus cocok dengan [agents/arsitektur.md](../../agents/arsitektur.md); ownership ditentukan oleh [boundary management](../../agents/boundary.md). Status runtime perlu dibuktikan dari kode/tes terbaru.

| Topik | Dokumen kanonik |
|---|---|
| Topologi keseluruhan dan trust boundaries | [System](system.md) |
| Sesi opaque, IAM, CSRF, 30m/8h, ABAC dan perbedaan legacy | [Security and Session](security-and-sessions.md) |
| Bounded context dan kontrak satu sumber kebenaran | [Ownership & Integration](service-boundaries.md) |
| Webhook Shopee/TikTok, sync, retry, reconciliation | [Marketplace & Events](marketplace-and-events.md) |
| Product metrics dan lineage | [Product Analytics](product-analytics.md) |
| Go workspace, deployment pelanggan, database dan RaaS | [Deployment & Workspace](deployment-and-workspace.md) |
| API GraphQL/gRPC | [Engineering contract guide](../04-engineering/api-contracts.md) |

Target arsitektur bukan klaim semua service aktif. **D01–D12 pada `agents/arsitektur.md` tidak boleh diedit lewat dokumen ringkasan ini.**
