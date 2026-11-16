# Verifikasi application service HRM

Scope: `backend/services/human_resource/internal/service/`, penambahan port/query scoped repository, dokumentasi dan tes HRM. Tidak mengubah DBML, baseline SQL, Gateway, IAM, Foundation atau Finance.

## Bukti dan hasil

| Area | Bukti | Hasil/batas |
|---|---|---|
| Wiring | Registry eksplisit 22 constructor, 84 port resource; performance reader talent | Build komposisi lulus; belum bootstrap runtime |
| DTO | Create tanpa server UUID/audit actor, patch tri-state, mapper field/default eksplisit | CRUD representatif satu resource/domain lulus; bukan positive CRUD seluruh tabel |
| Query scoped | Policy scope AND finder sebelum `SelectWhere`/`CountWhere`; kolom allowlist GORM | Unit mock list/count/search lulus; SQL baru belum diuji pada PostgreSQL |
| Mutasi | Row lock, integrity, mutasi+outbox satu transaksi, recheck actor/revisi | Tes rollback dependency/outbox/commit dan revoke lulus |
| Lifecycle | Guard menolak raw perubahan status/approval/field terlindungi serta riwayat final | Workflow yang belum memiliki command sengaja tetap ditolak |
| Cuti | Command posting ledger, replay dan proyeksi saldo | Duplicate/digest konflik, rollback dan 16 posting paralel lulus pada transaksi mock terserialisasi |
| Payroll | Rekonsiliasi result/line, larangan self approval, evidence server | Success, self approval, total salah, child tersembunyi dan rollback lulus |
| Talent | Rubrik/skor lengkap, lock assessment/cycle, pembaca review pemilik | Success, skor hilang, cycle tertutup, dimensi tersembunyi, reader hilang dan rollback lulus |
| JSON nullable | Clone mempertahankan nil slice entity.JSON | Create/patch tidak dikirim/NULL/literal null lulus |
| Benchmark | 26 operasi × enam pengulangan, `-benchmem`, dataset tetap | [Metrik dan output mentah](../../benchmark/human_resource/1-service-domain-pagination-dan-payroll-approval.md); bukan kapasitas produksi |

Perintah dari modul HRM: `GOWORK=off go fmt ./...`, `GOWORK=off go vet ./...`, `GOWORK=off go test -race ./...`, `GOWORK=off go build ./...`; seluruhnya lulus. Graph dipakai untuk penelusuran registry/command serta coverage scope service. Full reindex selesai; pemeriksaan sembilan file inti dan scope service tidak mencatat gap, metadata cocok. Sinyal coverage bersifat best-effort, bukan bukti audit seluruh kemungkinan perilaku.

## Batas yang harus ditutup sebelum aktivasi

1. `Policy` dan `Integrity` masih port wajib tanpa adapter produksi. Implementasi wajib sesuai D01–D12, katalog permission disetujui, validitas temporal dan referensi owner. Mock allow-all tidak boleh dipakai runtime.
2. Semua mutasi child harus mengunci parent dan menolak perubahan setelah final; jaminan transaksi mock tidak menggantikan uji concurrency PostgreSQL.
3. Pengambilan dataset command dibatasi 10.000 record; outbox scan dibatasi 10.000 event. Versi saat ini bergantung pada riwayat outbox yang dipertahankan; konflik menyebabkan rollback tanpa retry otomatis.
4. Count/list READ COMMITTED tidak menjamin satu snapshot. Filter AND/IN belum mengekspresikan semua policy ABAC; adapter harus menolak representasi yang memperluas izin.
5. Belum ada tarif payroll, posting Finance, promosi resmi, kalibrasi persisten, workflow lifecycle lengkap, bootstrap, gRPC, Gateway atau E2E. Status service layer tidak menyatakan management selesai.

Kerangka awal dibantu generator sebelum arahan penulisan langsung owner. Perubahan lanjutan memakai patch per berkas dan generator service tidak disimpan; provenance source dijelaskan dalam [panduan review](../../../backend/services/human_resource/docs/service.md).
