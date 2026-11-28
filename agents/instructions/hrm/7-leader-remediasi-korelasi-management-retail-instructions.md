# Penyambungan payroll Finance dan akses DRM per worker

- **id** — 7
- **owner** — hrm. Libatkan pemilik lawan pada matriks audit untuk perubahan di kedua ujung.
- **priority** — Ikuti prioritas temuan KM-06, KM-07.
- **Status:** Siap dikerjakan.
- **baseline** — Audit 28 November 2026 pada checkout lokal `e120e679` dan perubahan working tree yang tercantum dalam laporan.

## Task

Tutup bagian KM-06, KM-07 yang dimiliki hrm melalui alur bisnis aktual. Baca [laporan dan sumber temuan](../../audit/gateway/3-2026-11-28-172601-audit-korelasi-management-dan-logika-retail.md) serta [bukti pengujian audit](../../e2e/gateway/192-audit-korelasi-query-legal-entities-dan-mutation-user-login.md) sebelum mengubah kode.

## Request

### Penyambungan payroll Finance dan akses DRM per worker

- file: [cmd/main.go](../../../backend/services/human_resource/cmd/main.go) · [cmd/payslip_wiring.go](../../../backend/services/human_resource/cmd/payslip_wiring.go) · [infrastructure/integration/document_records/bridge.go](../../../backend/services/human_resource/infrastructure/integration/document_records/bridge.go) · [infrastructure/integration/document_records/payslip.go](../../../backend/services/human_resource/infrastructure/integration/document_records/payslip.go) · [infrastructure/inbox/payroll_publication.go](../../../backend/services/human_resource/infrastructure/inbox/payroll_publication.go) · [internal/service/payroll](../../../backend/services/human_resource/internal/service/payroll). Tambah client/consumer dan konfigurasi startup pada management pasangannya sesuai kontrak yang dijelaskan di laporan.
- function: `buildPayslipRegistryOptions`, `WithPayslipPublication`, `WithPayslipVerifier`, `WithPayslipReissue` dan client DRM store/read/restore; pertahankan operasi GraphQL HRM yang ada.
- logic: Pasang Finance verifier, DRMStorage, DRMReader dan font renderer terkendali di startup. Setelah Finance menerbitkan run, HRM render dari snapshot hasil, simpan key publication/result + SHA-256 dan terima exact version. Beri Finance binding item yang sesuai. Pegawai aktif mencetak sendiri. Mantan pegawai menggunakan `reissueEmployeePayslip` melalui HR; regenerasi hanya pada kondisi absensi yang dikonfirmasi owner.
- error_handling: Pisahkan belum pernah dibuat, blob hilang dengan metadata sah, forbidden, timeout dan tombstone. Restore byte identik pada versi logis yang sama bila diizinkan; koreksi gaji membutuhkan versi/publikasi baru. Replay event sebelum dan sesudah commit HRM idempoten.
- security: SELF terikat principal–worker; petugas HR butuh scope entity dan MFA, tanpa data gaji dalam event/log. Penerbitan Finance tetap berbeda dari pembayaran riil.
- expected_result: Satu run dengan dua worker menghasilkan dua PDF yang hanya dapat dibaca sesuai haknya. Retry sesudah DRM menyimpan hasil tidak menciptakan versi ganda.
- testing: Uji HTTPS `myPayslips`, `myPayslipPdf`, `reissueEmployeePayslip` serta callback Finance dengan HRM+DRM hidup. Tutup kasus terminasi, scope salah, pembayaran dijadwalkan, koreksi, missing blob dan DRM unavailable.

## Action

Urutan pengujian untuk hrm adalah run terbit di Finance → PDF per hasil → versi resmi DRM, lalu baca mandiri → cetak ulang oleh HR. Perluas [pengujian yang sudah ada](../../../backend/services/human_resource/internal/service/payroll/payslip_publication_test.go) dengan regresi untuk kasus di atas. Pada setiap tahap, catat ID transaksi uji, status awal dan akhir yang terbaca dari API owner, jumlah efek, serta hasil setelah retry. Gunakan data uji dan samarkan identitas di laporan.

1. Cocokkan identitas transaksi dan arti setiap status dengan owner lawan. Catat perubahan kontrak, event lama, perilaku untuk data historis, dan strategi pemulihan. Untuk RPC/SDL yang berubah perbarui inventaris di `agents/grpc/`.
2. Implementasikan bagian milik hrm dan koordinasikan kontrak pasangan dengan agent pemiliknya. Hubungkan startup, konfigurasi, permission, handler dan pemrosesan respons. Jalankan generator bila kontrak berubah, `go vet ./...`, `go test -race -count=1` pada paket terdampak dan E2E HTTPS. Tempatkan unit test gRPC di lapisan services. Verifikasi retry dan backfill terhadap efek yang sudah tercatat.
3. Selesaikan hingga kriteria hasil terpenuhi. Rekam case positif, penolakan, recovery, ACK dan versi owner, hasil aktual dan SHA di `agents/e2e/` serta working history sesuai AGENTS.md. Cantumkan operasi yang masih bergantung pada owner yang belum tersedia.

## Context

Setelah bukti penutupan dan commit tersedia, tambahkan SHA dan hasil uji, ubah status menjadi DONE dan beri akhiran `-instructions-done.md` sesuai aturan repositori. Pemilik masing-masing service memelihara skema, data dan keputusan bisnisnya.

[Aturan repo](../../../AGENTS.md) · [Boundary](../../boundary.md) · [Arsitektur](../../arsitektur.md) · [IAM dan akses objek](../../users.md) · [Logika domain](../../logic/) · [how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md)
