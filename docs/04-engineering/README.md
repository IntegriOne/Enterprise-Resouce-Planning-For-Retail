# Panduan Engineering

> **Pedoman engineering.** Ikuti [AGENTS.md](../../AGENTS.md), keputusan [arsitektur](../../agents/arsitektur.md), [ownership](../../agents/boundary.md) dan [scope](../01-product/vision-and-scope.md). Bukti tes terpisah dari rencana tes.

| Area | Dokumen |
|---|---|
| Onboarding dan workspace | [Local setup](local-setup.md) |
| Konvensi code/DDD/commit | [Code & Collaboration](coding-and-collaboration.md) |
| API GraphQL/protobuf/gRPC dan versions | [API Contracts](api-contracts.md) |
| Dua lapis pengujian dan skenario negatif | [Testing](testing.md) |
| Gate readiness dan test evidence | [Definition of Done & Quality](quality-gates.md) |

Semua developer bekerja langsung di `main`, tanpa branch baru, tidak force push; jangan menimpa perubahan developer lain pada code generated atau file konfigurasi bersama.
