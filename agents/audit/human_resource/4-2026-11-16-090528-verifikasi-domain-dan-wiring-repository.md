# Verifikasi domain dan wiring repository HRM

Cakupan: `backend/services/human_resource/domain/`, generator dan binding `repository/domain_writers.go`. Baseline Git sebelum pekerjaan: `f54e3029bfc09d70b1b2f69bd1b854cf1d02c9e8`.

## Hasil

- Entity/enum/port yang telah ada dilengkapi 86 constructor/validator dan 147 predicate CHECK statis, tanpa SQL runtime atau perubahan schema. Assertion generic memeriksa kesamaan entity validator dan port.
- `Registry.Writers()` mengikat seluruh 86 port; invalid/nil model ditolak sebelum Insert/Update. Error port dan context transaksi dipertahankan. Repository mentah tidak menjalankan keputusan bisnis.
- Aturan agregat sembilan domain diuji dengan fixture memori: penempatan/hiring, rekonsiliasi payroll, ledger/reversal/adjustment, talent/calibration, performance, career, learning dan bukti promotion. Case keputusan promosi tidak mengubah jabatan atau grant otomatis.
- Tes mock membuktikan validasi sebelum IO dan error propagation. Race test 16 writer serentak memakai start barrier dan port tersinkronisasi.
- Pada PostgreSQL Docker 18.6, dua database disposable akhiran `_test`: 86 constructor writer dan rejection nil; validasi snapshot dari 27 tabel per fixture; update melalui writer dalam rollback. Matrix repository 86 GetAll/Count, 503 finder dan 81 percobaan write per fixture tetap lulus. Fixture integritas SQL tetap melewati seluruh 70 penolakan. Database dibuang.
- Formatter, vet, race test dan build modul mandiri memakai `GOWORK=off` lulus. Tautan dan regenerasi deterministik diperiksa sebelum commit.

## Batas bukti dan tindak lanjut

Ini verifikasi domain/persistence, bukan E2E HRM melalui Gateway. Belum semua 86 tabel mempunyai fixture CRUD positif. Aturan row tidak membuktikan FK/unique, otorisasi, sumber master, kelengkapan halaman ledger, mandat approval atau readiness runtime. Sembilan domain mempunyai aturan agregat awal; domain lain memiliki validasi entity, belum keseluruhan kebijakan perusahaan. Workflow `services/`, clock/server context, read-lock dan outbox atomik, retry concurrency, integrasi IAM/Foundation/Finance/DRM serta E2E masih harus diimplementasikan. Lihat [kontrak domain](../../../backend/services/human_resource/docs/domain.md).
