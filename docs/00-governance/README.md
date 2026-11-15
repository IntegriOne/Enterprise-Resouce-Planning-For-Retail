# Tata Kelola Dokumentasi

> **Dokumentasi Integrione · 12 November 2026.** Penetapan owner, desain, implementasi, verifikasi dan deployment adalah status berbeda. Kebijakan normatif mengikuti [AGENTS.md](../../AGENTS.md), [arsitektur keamanan](../../agents/arsitektur.md), dan [batas management](../../agents/boundary.md).

Dokumen dalam folder ini memegang **aturan membaca dokumentasi**, bukan business logic. Peta keputusan dan sumber kebenaran: [keputusan](decisions.md), [status/evidence](status-and-evidence.md), [konvensi dokumen](documentation-policy.md), [glosarium](glossary.md).

## Urutan otoritas
1. Instruksi terbaru owner yang eksplisit dan keputusan keamanan D01–D12 di `agents/arsitektur.md`.
2. `AGENTS.md` untuk workflow, git, test dan delivery; `agents/boundary.md` untuk pemilik data.
3. Kontrak dan kode yang benar-benar ada di `backend/contracts/`, `backend/services/` dan `backend/gateway/`; bukti tes/CI untuk status VERIFIED.
4. ADR yang ACCEPTED, instruksi management dan dokumen proyek yang sudah disetujui.
5. Draf usulan, template, arsip dan contoh hanyalah bahan kerja.

Jangan memakai dokumen lama, klaim dari README, atau file proto sebagai bukti endpoint tersedia. Jika dua dokumen bertentangan, koreksi ke sumber yang berwenang dan catat bukti; jangan diam-diam mengganti keputusan owner.
