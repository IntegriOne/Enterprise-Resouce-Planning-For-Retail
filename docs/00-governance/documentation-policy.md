# Konvensi Dokumentasi dan Sumber Kebenaran

> **Dokumentasi Integrione · 12 November 2026.** Penetapan owner, desain, implementasi, verifikasi dan deployment adalah status berbeda. Kebijakan normatif mengikuti [AGENTS.md](../../AGENTS.md), [arsitektur keamanan](../../agents/arsitektur.md), dan [batas management](../../agents/boundary.md).

## Satu dokumen untuk satu keputusan
| Topik | Sumber utama |
|---|---|
| Aturan agent/Git/test/migrasi | [AGENTS.md](../../AGENTS.md) |
| Keamanan dan D01–D12 | [agents/arsitektur.md](../../agents/arsitektur.md) |
| Principal, grant, scopes | [agents/principals.md](../../agents/principals.md) |
| Ownership bisnis | [agents/boundary.md](../../agents/boundary.md) |
| Cakupan dan fase produk | [product vision](../01-product/vision-and-scope.md) |
| Status kode | Source Git dan [status berbukti](status-and-evidence.md) |
| Kontrak API | Proto/GraphQL aktif, [pedoman API](../04-engineering/api-contracts.md) |
| Rencana proyek | [project management](../06-project/project-plan.md) |
| Legal final | Dokumen ditinjau legal yang disahkan; draft di `docs/08-legal/` bukan perjanjian berlaku |

## Standar setiap dokumen
Heading jelas, penanggung jawab, status (`DRAFT/ACCEPTED/HISTORICAL`), tanggal & source revision jika status kode, fakta vs rekomendasi, scope dan out-of-scope, referensi sumber/tes, dependency & link relatif yang valid. Hindari mengulang aturan normatif dalam bentuk berbeda; tautkan ke sumbernya.

Dokumen rekayasa yang memuat contoh kode harus menyatakan apakah ilustratif atau sesuai source saat ini. Jangan menyimpan data identitas, kata sandi, bearer token, cookie atau kunci enkripsi di Markdown. Rencana, laporan UAT dan hasil tes menggunakan status `NOT RUN` sampai dieksekusi.

## Perubahan
Owner keputusan menyetujui perubahan substantif; agent developer mengubah dokumentasi seiring commit domain dalam `main` tanpa membuat branch. Jika rename path, audit semua backlink pada `AGENTS.md`, `agents/`, `services/`, `gateway/`, dan README root. Berkas lama dapat ditelusuri lewat Git history sebelum rekonstruksi docs; jangan menghidupkan kembali klaim legacy tanpa bukti.
