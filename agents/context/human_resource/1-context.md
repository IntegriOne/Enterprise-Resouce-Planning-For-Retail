# Konteks Human Resource — 16 November 2026

## Apa yang sudah dikerjakan

- Path aktif adalah `backend/services/human_resource/`; path lama `human_resource_management` dalam dokumentasi historis tidak membuktikan runtime aktif.
- Audit awal disimpan di `agents/audit/human_resource/1-2026-11-16-032217-audit-proporsionalitas-dan-integritas-skema-hrm.md`.
- Atas permintaan owner, riset CIPD memperluas DBML 45→69 tabel. Keputusan lanjutan 16 November 2026 menyatukan Talent Management dalam HRM dengan tambahan 17 tabel (total 86), tetap 11 kelompok bisnis + 1 integrasi; recruitment/ATS tetap HRM. Pemetaan rinci, sumber dan alasan: [desain skema](../../../backend/services/human_resource/docs/schema-design.md).
- Ditambahkan baseline tunggal `000001_hrm.up.sql`/`down.sql`, generator CLI DBML 12.0.0 dan invariant PostgreSQL untuk overlap, ledger immutable, rekonsiliasi approval payroll serta pembekuan snapshot.
- Verifikasi nyata PostgreSQL Docker 18.6 pada database sementara: 70 penolakan data tidak valid (38 HRM + 32 Talent), fixture valid, up/down/up lulus. Ini uji SQL, bukan service/E2E.
- Identitas/akses tetap IAM, master tetap Foundation, pembayaran/jurnal tetap Finance, versi dokumen tetap DRM; roster/absensi tidak diambil alih. Tidak ada seed role, endpoint, deployment atau migrasi database bisnis.

- [Desain Talent Management](../../../backend/services/human_resource/docs/talent-management.md): rubrik/asesmen/pool/kalibrasi, jalur/aspirasi karier, rekomendasi dan keputusan promosi, riwayat KPI serta kursus multi-kompetensi. Worker tidak diduplikasi. Approval tidak mengubah jabatan; applied wajib bukti assignment/event sesuai worker/target/tanggal. Rubrik terpakai, asesmen final dan keputusan dilindungi dari overwrite.

- Struktur folder pra-development mengikuti lapisan IAM: `domain/{entity,errors,repository,service}`, repository/services per 22 subdomain, kontrak lokal, gRPC privat, adapter owner dan inbox/outbox. [README HRM](../../../backend/services/human_resource/README.md) memetakan 86 tabel tanpa duplikasi; lapisan repository telah diimplementasikan pada task berikutnya, lapisan runtime lain masih placeholder.

- Repository GORM 86 tabel, entity/enum bertipe dan 503 finder, CRUD + GetAll(limit, offset)/Count, registry, transaksi context/savepoint dan row lock sudah tersedia. Default limit 20, maksimum 100, negatif ditolak; numeric memakai Decimal teks, NULL pointer/JSON eksplisit, Update snapshot lengkap tanpa upsert. [Kontrak dan bukti](../../../backend/services/human_resource/docs/repository.md). Modul mandiri diverifikasi dengan `GOWORK=off`: fmt/vet/test-race/build lulus. PostgreSQL Docker: seluruh model/kolom, GetAll/Count dan 503 finder pada dua fixture lulus; 27 tabel berisi data per fixture dan 81 percobaan write per fixture dibatalkan. Ini belum positive CRUD fixture seluruh tabel atau E2E.

- Domain: 86 constructor/Validate, sembilan enum, 147 CHECK dikompilasi menjadi Go statis, UUID/field/precision/evidence validation dan aritmetika decimal eksak. Sembilan package aturan agregat (personnel, recruitment, payroll, leave, talent, performance, career, learning, promotion), serta writer domain terikat ke 86 port melalui `Registry.Writers()`. Tes unit/mock/race dan wiring PostgreSQL fixture lulus. [Kontrak dan batas](../../../backend/services/human_resource/docs/domain.md). Tidak ada perubahan DBML/SQL atau aktivasi runtime.

## Apa yang belum dikerjakan

- Workflow Talent nyata, mandat panel/approval, eligibility perusahaan, scope/SoD terverifikasi IAM, persistensi finalisasi review/proyeksi KPI, transaksi promosi+outbox serta concurrency/E2E belum diimplementasikan; validasi bobot/bukti dan korelasi agregat sudah ada sebagai fungsi domain murni; detail batas ada di dokumen Talent.
- Proses runtime HRM, konfigurasi/koneksi DB dan bootstrap, kontrak internal, wiring GraphQL, consumer/outbox relay, otorisasi dan E2E belum dibuat pada folder aktif.
- Posting ledger dan proyeksi saldo atomik, approval lintas owner, validitas temporal master, SoD dan pembulatan/rule payroll lokal masih wajib pada service. Replay saldo/reversal tepat jumlah, rekonsiliasi total dan bukti promosi sudah divalidasi domain, belum workflow persisten.
- Tarif BPJS/pajak/THR, workflow HR perusahaan, proses clearance dan prioritas fitur belum dianggap keputusan implementasi otomatis dari riset.
- Temporal/immutable guard di baseline tidak membuktikan keseluruhan business workflow atau performa. Dokumen desain memuat batas klaim dan pekerjaan lanjutan.

## Apa yang harus dikerjakan selanjutanya

- Pakai port/registry repository yang sudah ada untuk mengimplementasikan F1 inti worker/assignment/kontrak/onboarding-offboarding dengan IAM, Foundation dan DRM sesuai kontrak owner; gunakan nama management pada kontrak yang disepakati, bukan otomatis mengikuti perbedaan path historis.
- Pastikan baseline selalu digenerate dengan `bash backend/services/human_resource/db/schema/generate.sh`; uji dengan `bash backend/services/human_resource/db/test/verify_schema.sh`.
- Payroll/cuti harus mengikuti transaksi domain, data sumber, aturan efektif dan persetujuan owner Finance/IAM; jalankan concurrency tests dan E2E nyata sebelum aktivasi.
- Kembangkan workforce planning, kompetensi, performance/learning/succession, engagement dan analytics bertahap berdasarkan kebutuhan nyata.
