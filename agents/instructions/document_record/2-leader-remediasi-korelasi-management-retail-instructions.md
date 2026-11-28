# Otorisasi versi, bukti payroll, retensi dan penghapusan tahan crash

- **id** — 2
- **owner** — document_record. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-05, KM-06, KM-07, KM-08, KM-09, KM-12.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-05, KM-06, KM-07, KM-08, KM-09, KM-12 yang dimiliki document_record melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Otorisasi versi, bukti payroll, retensi dan penghapusan tahan crash

- file: [cmd/wiring.go](../../../backend/services/document_record/cmd/wiring.go) · [services/document/service.go](../../../backend/services/document_record/services/document/service.go) · [services/publication/service.go](../../../backend/services/document_record/services/publication/service.go) · [repository/publication/repo.go](../../../backend/services/document_record/repository/publication/repo.go) · [services/records/service.go](../../../backend/services/document_record/services/records/service.go) · [infrastructure/storage/fsstore.go](../../../backend/services/document_record/infrastructure/storage/fsstore.go) · [infrastructure/scanner/engine.go](../../../backend/services/document_record/infrastructure/scanner/engine.go) · [infrastructure/integration/foundation/client.go](../../../backend/services/document_record/infrastructure/integration/foundation/client.go) · [contract/proto/document_record/v1](../../../backend/services/document_record/contract/proto/document_record/v1). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `VerifyDocumentVersion`, `StreamDocument`, `AuthorizePublicationRecipient`, `SelectActiveRecipientMatch`, `ApplyRetentionAnchor`, `ApproveDisposition`, worker scan/purge dan operasi privat payslip.
- logic: Validasi owner bisnis, entity, source, subject, purpose, type dan versi sebelum grant metadata/bytes/publikasi; resolve PIC dengan EF dan status pegawai/buyer pada owner. Pisahkan pencabutan serving + commit fence/job dari delete fisik. Worker claim lease lalu delete generation persis, lookup hasil, commit bukti/tombstone. Bind trigger retention ke owner/type/source yang dibekukan, recheck hold dan owner clearance. Hubungkan port store/read/restore HRM serta relay untuk event owner. Gunakan scanner yang mencakup format/policy resmi; tahan dokumen tidak dapat dipindai.
- error_handling: Audit allow wajib tersimpan sebelum stream. Error retry/unknown setelah purge tidak boleh hilang karena rollback; simpan state durable, blok serving/restore dan reconcile. Hindari rollback yang menghidupkan kembali metadata untuk file terhapus. Duplikat berpayload berbeda ditahan.
- security: MFA pada HR/legal review, SoD, least privilege untuk worker purge, cert/audience per owner. Dokumen HR/Finance tetap membutuhkan hak owner saat staf punya izin global. Signed file, staging, cache dan event tidak boleh membuka bytes ke pihak lain.
- expected_result: Hold pending menghentikan purge; restart sesudah delete mempertahankan fence dan akhirnya menemukan bukti. Slip hanya dapat dibaca pekerja aktif yang sesuai atau HR berwenang untuk reissue. Arsip supplier dapat diakses PIC sah tanpa memberi akses ke party lain.
- testing: Uji storage disposable: interruption setelah fence, sesudah delete sebelum commit, dua hold, event salah source, audit down, revoke recipient di sela chunk, scanner offline, party berubah serta lifecycle PDF HRM–Finance aktual.

## Action

Urutan pengujian untuk document_record adalah pemeriksaan owner → izin baca, lalu hold → fence → purge → rekonsiliasi, selanjutnya scan → seal → publikasi dan konfirmasi penerima. Perluas [pengujian yang sudah ada](../../../backend/services/document_record/services/records/test/acceptance_test.go) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik document_record dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
