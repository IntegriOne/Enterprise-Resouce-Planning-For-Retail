# Release, Maintenance dan Support

> **Operasional Retail ERP — draft SOP.** Tidak ada perintah deployment produksi atau tes restore yang diklaim berhasil melalui penulisan dokumen. Kebijakan owner dan gate ada di [decisions](../00-governance/decisions.md), [quality gates](../04-engineering/quality-gates.md).

- [Release dan change](release-and-deployment.md) — persetujuan, checklist, release notes/changelog.
- [Runbooks dan incident](runbooks-and-incidents.md) — pemulihan webhook, OMS, IAM dan prosedur postmortem.
- [Backup dan disaster recovery](backup-and-dr.md) — PostgreSQL, IAM/Redis, dokumen dan checkpoints.
- [Support & manuals](support-and-user-guides.md) — quick-start internal, supplier, vendor, pembeli, system admin dan FAQ.

Jangan anggap semua endpoint tersedia hanya karena ada panduan. Alur customer baru untuk fase terakhir. Fitur marketplace memerlukan kontrak API resmi yang diverifikasi. Ketika incident terjadi, operator harus membedakan ack transport, state owner, status `PENDING`, dan klaim hasil akhir.
