# Audit proporsionalitas dan integritas skema HRM

- Tanggal: 16 November 2026, WIB.
- Lingkup: seluruh 1.003 baris [DBML HRM](../../../backend/services/human_resource/db/schema/human_resource_management.dbml), termasuk 45 tabel, 9 enum, 56 relasi dan SQL hasil ekspor.
- Baseline Git: `a17eac35a7640cc22b01b1315ff81e6c918753e9`.
- SHA-256 DBML yang diperiksa: `ebc263902dc8b2e686ade10c7bdb4ff544ad46760b62ef12e61c3e926033d315`.
- Berkas sumber masih untracked saat audit; SHA Git di atas tidak mengidentifikasi isi DBML. Hash berkas menjadi identitas bukti.
- Path dari permintaan memakai `human_resouce`; path aktual adalah `human_resource`. Dokumen lama memakai `human_resource_management`; laporan memakai nama direktori aktual tanpa melakukan rename.
- Tidak mengubah DBML, kode, konfigurasi atau database. Ekspor SQL hanya di direktori sementara.
- Skill: `aivana-database-engineer:cognitive-schema-mapper`; pemetaan semantik dan review struktur dilakukan oleh agen. Workflow advisor vendor ERP dan reviewer SQL query tidak diterapkan karena objeknya DBML ERP kustom, tanpa workload vendor atau query aplikasi.
- Kontrak keluaran skill: `usp: cognitive_schema_mapper`, `source: analysis`. Bagian pemetaan merupakan `ontology`; bagian kesenjangan merupakan `semanticGaps`; rekomendasi, bukti dan tingkat keyakinan dijelaskan di bawah.

## Penilaian utama

**Terlalu luas bila semua fitur wajib diimplementasikan sekarang untuk F1 Retail ERP; cukup masuk akal sebagai peta HRM jangka panjang. Namun kelengkapan fitur belum seimbang dengan kontrol integritas pada fungsi inti.** Jumlah 45 tabel sendiri bukan bukti desain buruk atau lambat. Pemisahan worker, assignment, kontrak, payroll run/result/line dan relasi banyak-ke-banyak memang berguna.

Kelebihan utama adalah cakupan suite HR lintas industri: rekrutmen mendetail, penilaian kinerja, pembelajaran, suksesi dan kerangka metrik generik. Kekurangan yang lebih mendesak berada pada saldo cuti, konsistensi payroll, periode efektif, kepemilikan referensi internal, dan mekanisme integrasi yang tahan retry.

Penilaian prioritas didasarkan pada fase internal lebih dahulu di [visi produk](../../../docs/01-product/vision-and-scope.md), [boundary](../../boundary.md), [arsitektur](../../arsitektur.md) dan [konsep principal](../../principals.md). Dokumen tersebut memuat snapshot historis; audit ini tidak menganggap service lama masih aktif. Pada direktori `backend/services/human_resource/` yang diperiksa, inventaris berkas hanya menemukan DBML ini, bukan implementasi service/migrasi untuk membuktikan guard tambahan.

## Pemetaan konsep dan ukuran desain

| Kelompok | Jumlah tabel | Penilaian untuk kebutuhan sekarang |
|---|---:|---|
| Job/posisi, worker, kontrak, assignment, lifecycle, dokumen, kontak darurat | 9 | Fondasi; pertahankan pemisahan fakta dan riwayat |
| Rekrutmen dari requisition sampai offer | 8 | Layak bila rekrutmen internal lewat ERP benar-benar dibutuhkan; panel, feedback versi dan tahapan lengkap dapat ditunda |
| Onboarding plan/task | 2 | Berguna bila ada checklist operasional; jangan menyamakan tugas selesai dengan grant IAM efektif |
| Compensation band/record | 2 | Record inti; band dan struktur grade lebih lanjut mengikuti kebutuhan perusahaan |
| Benefit plan/enrollment | 2 | Berguna untuk kepesertaan dan kontribusi; bukan bukti perhitungan statutory lokal siap |
| Payroll profile/component/period/run/result/line/adjustment | 7 | Pemisahan wajar; implementasi menunggu sumber input, aturan dan Finance yang disepakati |
| Leave policy/entitlement/adjustment | 3 | Diperlukan tetapi baru snapshot, belum alur saldo otoritatif lengkap |
| Performance cycle/goal/review | 3 | Tahap setelah administrasi pekerja, cuti dan penggajian stabil |
| Learning course/enrollment/qualification | 3 | Qualification dapat dibutuhkan awal; katalog kursus dan enrollment menyusul |
| Succession plan/candidate | 2 | Prioritas rendah tanpa kebutuhan suksesi posisi yang nyata |
| Employee relations case/event | 2 | Aktifkan sesuai kebutuhan kasus, pembatasan akses dan penanggung jawab; bukan otomatis dihapus |
| Human-capital metric definition/snapshot | 2 | Tunda mesin metrik generik; laporan terarah dari fakta owner cukup untuk awal |
| **Total** | **45** | Tidak perlu menghapus tabel hanya untuk mencapai target jumlah |

Bagian yang baik: ID pekerja terpisah dari IAM principal/party, referensi lintas management tidak diberi FK lintas database, nilai memakai decimal, beberapa periode dan versi eksplisit, nomor/idempotency key unik, dokumen sensitif diarahkan ke penyimpanan terlindungi. FK komposit primary assignment dan payroll profile sudah mengikat pekerja yang sesuai; pola ini patut diterapkan konsisten.

## Kesenjangan dan temuan terprioritas

P1 berarti harus ditangani sebelum fungsi terkait diaktifkan; P2 berarti penajaman desain berikutnya. Ini bukan klaim kerentanan produksi yang telah dieksploitasi.

### P1-01 — Saldo cuti belum mempunyai sumber transaksi lengkap

**Bukti:** DBML baris 678–735. `hr_leave_entitlements` menyimpan opening/accrued/used/adjusted/closing snapshot; note menyerahkan approved absence dan pemakaian ke Workforce. Tidak terdapat ledger accrual, consumption, reversal atau rekaman pemakaian ber-ID sumber unik dalam DBML ini. CHECK juga tidak mengikat saldo akhir ke komponennya.

**Dampak:** histori penyebab saldo berubah, deduplikasi pemakaian, pembatalan dan carryover belum jelas. Ini penting karena boundary memberi HRM kepemilikan hak/saldo cuti, sementara Workforce bukan management aktif dalam daftar 13.

**Rekomendasi:** tetapkan ledger hak/saldo di HRM, dengan debit/kredit, sumber unik, waktu efektif dan reversal. Saldo snapshot tetap boleh dipakai sebagai proyeksi. Kontrak permohonan absen/presensi, kalender dan pemakaian dari owner operasional harus disepakati; jangan otomatis membuat management Workforce baru atau mengambil alih roster. Tanpa kontrak itu, nyatakan alur terkait BLOCKED. Uji saldo awal, accrual, penggunaan, pembatalan, expiry/carryover dan replay.

### P1-02 — Total payroll dapat tidak konsisten dengan komponennya

**Bukti:** baris 622–658. Constraint hasil hanya nonnegatif dan `net_pay_amount <= gross_amount`; tidak ada persamaan nominal, hubungan ke agregat result lines, atau aturan rekonsiliasi yang dinyatakan. Contoh gross 10.000.000, deductions 2.000.000, net 9.500.000 lolos kedua ekspresi tersebut.

**Dampak:** schema menerima kombinasi yang tidak cocok dengan arti umum gross/deduction/net. Ini bukti keterbatasan constraint, bukan bukti aplikasi menghasilkan gaji salah.

**Rekomendasi:** definisikan arti gross, reimbursements, deductions dan employer cost. Bila definisi gross memang seluruh basis pembayaran, ikat net ke gross dikurangi deductions; bila reimbursement terpisah, gunakan persamaan sesuai komponen. Rekonsiliasi total result lines pada finalisasi transaksi; CHECK PostgreSQL bukan mekanisme untuk menjumlah baris lain. Bedakan pembulatan amount `numeric(20,2)` dari rate/quantity presisi tinggi. Untuk rancangan multi-currency, definisikan skala mata uang, jangan menganggap dua desimal universal.

### P1-03 — Status approved/paid belum terikat bukti keputusan lengkap

**Bukti:** baris 592–620. `approved_at` diwajibkan untuk approved/posted/paid, tetapi `approved_by_principal_id` tetap nullable. Status paid mensyaratkan journal reference, tidak mensyaratkan payment batch/reference atau bukti status Finance. `input_digest`/note immutable bukan pencegah UPDATE hasil yang sudah disetujui. Run type correction/reversal tersedia, tetapi tidak ada pengaitan eksplisit ke run yang dikoreksi.

**Dampak:** jejak persetujuan, pemisahan tugas, pembatalan dan sumber status paid belum tertutup pada rancangan ini.

**Rekomendasi:** ikat keputusan ke pelaku, revisi dan digest yang disetujui; tentukan transisi dan penguncian hasil final, serta referensi koreksi/reversal. SoD dan IAM recheck tetap di service. Status posting/pembayaran hanya proyeksi fakta Finance yang tervalidasi dengan versi/idempotensi, bukan flag bebas HRM. Jangan membuat kemampuan pembayaran bank di HRM.

### P1-04 — Ada referensi internal yang belum menjaga kepemilikan pekerja

**Bukti:** baris 474–494 dan 972–974. `hr_compensation_records.worker_id` dan `assignment_id` masing-masing valid sendiri tetapi tidak diikat ke pekerja yang sama. Record gaji pekerja A dapat menunjuk assignment pekerja B pada tingkat FK yang tertulis. Bandingkan FK komposit yang benar pada baris 951 dan 982.

**Bukti tambahan:** `hr_employee_relations_cases.subject_worker_id` baris 885 tidak mempunyai Ref ke `hr_workers`; ini berbeda dari referensi lintas management yang memang tidak boleh diberi FK.

**Rekomendasi:** gunakan FK komposit `(assignment_id, worker_id)` ke `(id, worker_id)` yang sudah unik pada assignment, serta FK nullable subject_worker untuk kasus HR. Referensi kompensasi pada kontrak/offer juga perlu ditegaskan target dan lifecycle-nya; jangan menambah FK dengan menebak owner.

### P1-05 — Keunikan tanggal mulai tidak mencegah periode tumpang tindih

**Bukti:** assignment baris 208–231, compensation baris 474–494, payroll profile baris 535–553. Unique key menggunakan tanggal awal, bukan rentang periode. Dua gaji pokok aktif dengan tanggal awal berbeda tetapi masa berlaku bertumpuk masih bisa memenuhi constraint.

**Rekomendasi:** definisikan kombinasi yang boleh paralel dan yang wajib tunggal. Guard periode aktif memerlukan constraint rentang/EXCLUDE pada baseline SQL jika sesuai, atau protokol locking/transaksi dan tes konkurensi yang setara. Jangan melarang semua assignment paralel karena schema sengaja mendukung part-time/concurrent. Primary assignment harus milik worker, bertipe utama dan berlaku pada tanggal keputusan. Tentukan konvensi batas periode agar pergantian tanggal tidak ambigu.

### P1-06 — Konteks badan hukum payroll tidak dibekukan dengan jelas

**Bukti:** worker/contract/position memuat badan hukum, tetapi payroll profile/period/run/result tidak mempunyai employing legal entity snapshot. `payroll_group_code` hanya string, tanpa definisi grup dan pemetaan employer dalam berkas ini.

**Dampak:** pekerja pindah badan hukum setelah payroll dihitung; pembaca dapat keliru memakai employer terkini untuk hasil historis. Yurisdiksi saja tidak membedakan dua perusahaan dalam negara yang sama. Single-tenant tetap dapat mempunyai beberapa legal entity.

**Rekomendasi:** kontrak input snapshot harus mengunci employing entity, assignment/contract, grup/periode dan revisi sumber; pertimbangkan kolom employer pada aggregate run/result sesuai model satu run per employer. Tidak perlu menambah `tenant_id` atau master perusahaan lokal. Verifikasi scope melalui referensi Foundation.

### P1-07 — Integrasi IAM/Finance belum mempunyai model pengiriman yang tahan gagal

**Bukti:** seluruh 45 tabel dan 56 Ref telah diperiksa. Tidak terdapat outbox/inbox atau delivery receipt pada DBML. Lifecycle event baris 234–248 mempunyai idempotency key tetapi belum mencakup pengiriman, acknowledgement dan recovery. Permintaan IAM tidak sama dengan perubahan `iam_principal_reference_id`.

**Rekomendasi:** pastikan baseline implementasi memiliki transactional outbox untuk perubahan HRM yang perlu diteruskan dan inbox/deduplikasi untuk fakta Finance atau pemakaian waktu. Tabel teknis ini boleh berasal dari mekanisme bersama yang disepakati, tidak harus menciptakan framework event generik baru. Hubungkan onboarding/termination dengan request IAM dan pantau hasilnya; tugas HR selesai bukan bukti akses sudah diberikan/dicabut. Tidak ada grant/permission bisnis dalam JSON pekerja.

### P2-01 — Batas FTE terlalu permisif untuk makna persentase biasa

**Bukti:** baris 215 dan 229: default 100, tetapi CHECK mengizinkan hingga 1000. Nilai 1000 memenuhi ekspresi.

**Rekomendasi:** jika 100 berarti satu pekerja penuh, batas per assignment lazimnya maksimum 100; tetapkan aturan bisnis sebelum mengubah. Agregat FTE lintas assignment dan kapasitas posisi memerlukan validasi periode/konkurensi. Jika sengaja bermakna rasio lain, perjelas unit dan nama. Temuan ini berkeyakinan tinggi pada nilai literal, menengah pada batas penggantinya karena kebijakan staffing belum diberikan.

### P2-02 — Unique metrik belum menjamin satu snapshot pada scope NULL

**Bukti:** baris 925–943. Scope unit organisasi dan jurisdiction nullable, tetapi unique index biasa mencakup kedua kolom. SQL ekspor memakai `CREATE UNIQUE INDEX` tanpa `NULLS NOT DISTINCT`. Snapshot total perusahaan tanpa unit/jurisdiction dapat berulang dengan key lain yang sama.

**Rekomendasi:** tentukan apakah satu baris current atau histori banyak versi yang diinginkan. Untuk satu baris per scope gunakan semantik NULL yang tepat; untuk histori tambahkan identitas versi/snapshot secara eksplisit. PostgreSQL menjelaskan bahwa NULL pada unique default tidak dianggap sama. Ini masalah determinisme metrik, bukan kebutuhan mengaktifkan analytics sekarang.

### P2-03 — Jejak audit, metadata dokumen dan referensi aturan belum konsisten

**Bukti:** created/updated/actor tersedia tidak merata; worker lifecycle mempunyai history tetapi compensation, benefit dan adjustment belum mempunyai model koreksi lengkap. Dokumen banyak memakai `document_reference_id` tanpa definisi bahwa ID selalu menunjuk versi immutable. `formula_reference_id`, competency profile, contribution rules, template onboarding dan input snapshot menunjuk kemampuan yang owner/format/version kontraknya belum dijelaskan dalam berkas ini.

**Rekomendasi:** untuk aggregate yang berubah, pilih revision/CAS atau locking yang jelas, rekam pelaku/alasan dan status approval. Definisikan reference ID ke versi DRM tetap atau sertakan version/digest. Buat inventaris dependency untuk formula/snapshot sebelum membuat engine generik. Label `restricted` dan referensi vault membantu klasifikasi tetapi tidak membuktikan RBAC/ABAC, audit baca, enkripsi atau retensi telah ditegakkan. Jangan menambah kolom audit yang sama membabi buta pada tabel immutable; desain sesuai lifecycle.

### P2-04 — Kompleksitas dan constraint yang kurang selaras

- `position_code` sudah unik global; unique tambahan `(organization_unit_reference_id, position_code, valid_from)` tidak menambah keunikan bisnis. Mungkin masih berguna sebagai indeks akses organisasi, tetapi tidak perlu otomatis UNIQUE. Jika hendak menyimpan versi posisi berkode sama, unique global justru menghalangi; tentukan modelnya.
- Unique kompensasi `(worker_id, compensation_type, effective_from)` tidak membedakan assignment atau jenis allowance. Dua allowance berbeda pada tanggal sama dapat tertolak. Tentukan apakah satu total per tipe atau komponen terpisah sebelum mengganti key.
- Enum `hr_approval_decision_enum` didefinisikan tetapi tidak dipakai kolom mana pun dalam DBML. Jangan membuat approval engine hanya untuk memakai enum.
- Rujukan ISO/human-capital standard pada komentar bukan alasan mewajibkan seluruh modul atau bukti compliance. Tidak dilakukan audit kepatuhan standar atau hukum pada tugas ini.

## Kebutuhan lokal yang perlu diperjelas, bukan otomatis menambah tabel

Payroll Indonesia masih dinyatakan lewat jurisdiction, formula/provider reference, benefit dan komponen generik. Belum ada spesifikasi kontrak yang membuktikan perhitungan THR, BPJS, PPh 21, prorata, lembur, koreksi retroaktif dan final pay siap. Tentukan fitur yang benar-benar diminta, sumber waktu/kalender, versi aturan, tanggal berlaku, pembulatan, serta bukti perhitungan. Gunakan konfigurasi/komponen berversi bila cukup; jangan membuat satu tabel per istilah pajak. Tarif dan ketentuan hukum tidak dievaluasi di audit skema ini.

Offboarding perlu alur keputusan HR, pengembalian aset melalui owner, final payroll dan pemantauan pencabutan IAM. Lifecycle termination saja belum membuktikan seluruh proses selesai; fungsi tersebut dapat memakai checklist yang terfokus tanpa menambah engine workflow lintas perusahaan.

## Urutan rekomendasi

1. Kunci worker, job/position seperlunya, assignment, kontrak dan dokumen, termasuk temporal integrity dan Foundation/IAM references.
2. Lengkapi transaksi hak/saldo cuti serta sumber penggunaan yang disepakati owner; pertahankan scope 13 management.
3. Matangkan compensation, benefit yang diperlukan dan payroll; tutup nominal, employer snapshot, approval, koreksi dan kontrak Finance sebelum aktivasi.
4. Tambahkan outbox/inbox dan audit yang memang dibutuhkan oleh alur tersebut; validasi retry, concurrency, deny tanpa write dan seluruh Query/Mutation saat implementasi.
5. Rekrutmen lengkap, appraisal, learning suite, succession dan mesin metrik generik menunggu kebutuhan operasional yang terkonfirmasi. Simpan sebagai rancangan lanjutan; audit ini tidak memerintahkan penghapusan.

## Verifikasi dan batas bukti

| Pemeriksaan | Hasil |
|---|---|
| Membaca seluruh DBML 1.003 baris | PASS |
| Inventaris deklarasi | 45 tabel, 9 enum, 56 Ref, 12 kolom JSONB |
| Graph search HRM | Tidak menemukan simbol; bukan bukti tidak ada source |
| Coverage path DBML | `not_tracked`; pembacaan langsung seluruh source dilakukan |
| `npm exec --yes --package=@dbml/cli -- dbml2sql <path-dbml> --postgres -o /tmp/hrm-schema-audit-20261116.sql` | PASS, CLI 12.0.0 |
| Inspeksi hasil ekspor | 45 CREATE TABLE, 9 CREATE TYPE, 56 FK, 43 CHECK |
| Evaluasi contoh ekspresi FTE dan nominal melalui Python | Nilai bermasalah memenuhi ekspresi yang tertulis; bukan INSERT PostgreSQL |
| Menjalankan SQL pada PostgreSQL | NOT RUN |
| Unit/race/E2E/security runtime/performance | NOT RUN; bukan scope audit DBML |

Sintaks relasi optional `>?` diterima CLI 12.0.0 dan didokumentasikan DBML; tidak dikategorikan sebagai kesalahan. Ekspor SQL berhasil tidak membuktikan migrasi database maupun aturan aplikasi lulus.

**Keyakinan:** tinggi untuk struktur, relasi dan constraint yang terlihat; menengah untuk prioritas pengurangan fitur karena jumlah pegawai, frekuensi rekrutmen dan kebutuhan payroll perusahaan belum tersedia; tidak ada kesimpulan performa atau compliance runtime.

Referensi teknis primer: [sintaks DBML](https://dbml.dbdiagram.io/docs/), [constraint PostgreSQL 18](https://www.postgresql.org/docs/18/ddl-constraints.html). Dokumentasi PostgreSQL digunakan untuk semantik FK, CHECK lintas baris, exclusion dan unique NULL; tidak ada benchmark yang dijadikan dasar usulan indeks.
