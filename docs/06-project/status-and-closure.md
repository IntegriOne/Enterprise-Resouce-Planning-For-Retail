# Status Report, Sprint Review, Retrospective dan Project Closure

> **Template pengelolaan proyek, status DRAFT/NOT RUN** sampai persetujuan atau pengujian aktual. Rujukan: [Charter](project-charter.md), [Requirements](../01-product/requirements.md), [DoD](../04-engineering/quality-gates.md).

## Laporan progres per periode
`period | reviewed HEAD SHA | story/domain | DECIDED/DESIGNED/IMPLEMENTED/VERIFIED/BLOCKED/DEPLOYED | dependency owner | actual tests | risks | next action | responsible`. Data status dari audit dan test terbaru; tidak memakai persentase berdasarkan jumlah tabel/commit.

## Sprint review
Demo alur bisnis sesuai fase aktif, catat story yang diterima berdasarkan DoD, dua aktor scope, data PostgreSQL dan event replay. Jumlah story accepted = velocity; burndown mengikuti perubahan scope melalui CR. No branch: semua commit ada di `main`, jangan memerlukan PR sebagai syarat owner review.

## Retrospektif dan lessons learned
Keep / Improve / Try dengan bukti konkret. Bahas sumber regressions (shared gqlgen/proto), scope cross-owner yang belum disepakati, blocker provider, test fixture serta komunikasi. Setiap pembelajaran punya owner, tindakan, tenggat dan test pembuktian; tidak mengarang jam tim dan biaya tanpa log.

## Project Closure Checklist
- [ ] SOW/PRD/RTM seluruh deliverable yang **disetujui** memiliki UAT sign-off.
- [ ] Immutable release SHA/tag, changelog, runbook dan data migration record.
- [ ] Backup restore, keamanan IAM/CSRF dan cross-party cases diuji.
- [ ] Data partner/merchant dan tokens dipindah/dicabut sesuai kontrak.
- [ ] Known issues, SLA, support kontak dan owner eskalasi diserahkan.
- [ ] Legal Privacy/Terms/DPA berlaku, bukan draft.
- [ ] Finance/OMS/WMS/data provider direkonsiliasi; analytics freshness diketahui.
- [ ] Owner dan pelanggan mengonfirmasi acceptance terakhir.

Jika belum memenuhi, project status PARTIAL/BLOCKED, bukan CLOSED. Handoff support menggunakan [operations](../07-operations/support-and-user-guides.md) dan [runbook](../07-operations/runbooks-and-incidents.md).
