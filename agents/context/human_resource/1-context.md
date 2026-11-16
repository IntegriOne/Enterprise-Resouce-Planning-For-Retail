# Konteks Human Resource — 16 November 2026

## Apa yang sudah dikerjakan

- Path aktif adalah `backend/services/human_resource/`; path lama `human_resource_management` dalam dokumentasi historis tidak membuktikan runtime aktif.
- Audit awal disimpan di `agents/audit/human_resource/1-2026-11-16-032217-audit-proporsionalitas-dan-integritas-skema-hrm.md`.
- Atas permintaan owner, riset CIPD memperluas DBML 45→69 tabel. Keputusan lanjutan 16 November 2026 menyatukan Talent Management dalam HRM dengan tambahan 17 tabel (total 86), tetap 11 kelompok bisnis + 1 integrasi; recruitment/ATS tetap HRM. Pemetaan rinci, sumber dan alasan: [desain skema](../../../backend/services/human_resource/docs/schema-design.md).
- Ditambahkan baseline tunggal `000001_hrm.up.sql`/`down.sql`, generator CLI DBML 12.0.0 dan invariant PostgreSQL untuk overlap, ledger immutable, rekonsiliasi approval payroll serta pembekuan snapshot.
- Verifikasi nyata PostgreSQL Docker 18.6 pada database sementara: 70 penolakan data tidak valid (38 HRM + 32 Talent), fixture valid, up/down/up lulus. Ini uji SQL, bukan service/E2E.
- Identitas/akses tetap IAM, master tetap Foundation, pembayaran/jurnal tetap Finance, versi dokumen tetap DRM; roster/absensi tidak diambil alih. Tidak ada seed role, endpoint, deployment atau migrasi database bisnis.

- [Desain Talent Management](../../../backend/services/human_resource/docs/talent-management.md): rubrik/asesmen/pool/kalibrasi, jalur/aspirasi karier, rekomendasi dan keputusan promosi, riwayat KPI serta kursus multi-kompetensi. Worker tidak diduplikasi. Approval tidak mengubah jabatan; applied wajib bukti assignment/event sesuai worker/target/tanggal. Rubrik terpakai, asesmen final dan keputusan dilindungi dari overwrite.

- Struktur folder pra-development mengikuti lapisan IAM: `domain/{entity,errors,repository,service}`, repository/services per 22 subdomain, kontrak lokal, gRPC privat, adapter owner dan inbox/outbox. [README HRM](../../../backend/services/human_resource/README.md) memetakan 86 tabel tanpa duplikasi; seluruh folder baru masih placeholder, bukan runtime.

## Apa yang belum dikerjakan

- Workflow Talent nyata, mandat panel/approval, eligibility/bobot rubrik, scope/SoD terhadap pekerja, finalisasi performance review, proyeksi KPI, transaksi promosi+outbox serta concurrency/E2E belum diimplementasikan; detail batas ada di dokumen Talent.
- Runtime HRM/Go, kontrak internal, wiring GraphQL, consumer/outbox relay, otorisasi dan E2E belum dibuat pada folder aktif.
- Posting ledger dan proyeksi saldo atomik, reverse tepat jumlah, approval lintas owner, validitas temporal lintas objek, SoD dan pembulatan/rule payroll lokal masih wajib pada service.
- Tarif BPJS/pajak/THR, workflow HR perusahaan, proses clearance dan prioritas fitur belum dianggap keputusan implementasi otomatis dari riset.
- Temporal/immutable guard di baseline tidak membuktikan keseluruhan business workflow atau performa. Dokumen desain memuat batas klaim dan pekerjaan lanjutan.

## Apa yang harus dikerjakan selanjutanya

- Implementasikan F1 inti worker/assignment/kontrak/onboarding-offboarding dengan IAM, Foundation dan DRM sesuai kontrak owner; gunakan nama management pada kontrak yang disepakati, bukan otomatis mengikuti perbedaan path historis.
- Pastikan baseline selalu digenerate dengan `bash backend/services/human_resource/db/schema/generate.sh`; uji dengan `bash backend/services/human_resource/db/test/verify_schema.sh`.
- Payroll/cuti harus mengikuti transaksi domain, data sumber, aturan efektif dan persetujuan owner Finance/IAM; jalankan concurrency tests dan E2E nyata sebelum aktivasi.
- Kembangkan workforce planning, kompetensi, performance/learning/succession, engagement dan analytics bertahap berdasarkan kebutuhan nyata.
