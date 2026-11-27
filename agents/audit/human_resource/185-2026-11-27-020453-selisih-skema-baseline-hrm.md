# Selisih skema HRM sebelum rollout

Tujuh migrasi dari commit pemulihan dan baseline `000001` pada checkout dijalankan pada dua database PostgreSQL sementara. Perbandingan memakai definisi tabel, kolom, enum, constraint, dan indeks dari katalog PostgreSQL. Kedua database sementara dihapus setelah hasil diambil.

| Objek | Hasil |
| --- | --- |
| Tabel | Satu tabel baru, `hr_worker_assignment_transfers` |
| Kolom | 39 penambahan, termasuk 13 pada tabel baru. Seluruh 26 kolom baru pada tabel lama menerima null. |
| Constraint | 37 definisi hanya ada di baseline baru. Satu exclusion constraint lama diganti. |
| Indeks | 12 definisi hanya ada di baseline baru. Dua definisi lama diganti. |
| Enum | Definisi sama untuk enum pada skema yang dibandingkan. |

Kolom baru menyentuh assignment pekerja, offboarding, payroll, cuti, workforce, penawaran, dan suksesi. Exclusion constraint assignment primer serta indeks unik periode assignment versi baru hanya berlaku ketika `cancelled_at IS NULL`.

Perubahan tracker saja akan melewatkan DDL yang diperlukan. Evaluasi upgrade perlu memeriksa data lama terhadap constraint baru, menjalankan DDL pada salinan terisolasi, dan membandingkan fungsi, trigger, serta definisi objek lain sebelum memilih jalur migrasi. Audit ini memakai database kosong sehingga belum membuktikan bahwa data development dapat dipindah tanpa koreksi. [Insiden rollout dan pemulihan](184-2026-11-27-020305-kompatibilitas-rollout-grpc.md).
