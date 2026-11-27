# Verifikasi anggaran requisition

HRM membutuhkan endpoint privat Finance untuk memeriksa otorisasi anggaran pada dua tahap.

- `approveRequisition` mengirim snapshot kebutuhan dan menerima UUID otorisasi Finance.
- `openRequisition` mengirim snapshot serta UUID yang sudah disimpan. Finance memeriksa ulang bahwa otorisasi yang sama masih aktif dan cocok.

Snapshot mencakup requisition, unit, jabatan, headcount, kategori pekerja, target tanggal, dan referensi opsional seperti posisi, lokasi, atau lini workforce plan. Finance perlu menentukan cara menemukan otorisasi awal dan membedakan data hilang, dicabut, kedaluwarsa, atau berubah dari gangguan layanan. Jika pemeriksaan gagal, HRM mempertahankan status dan outbox.

Mandat peninjau pada unit yang tepat sudah diverifikasi melalui IAM privat. Finance memverifikasi anggaran melalui workload HRM terautentikasi. Kedua keputusan dibutuhkan sebelum requisition bisa dibuka. Uji HTTPS dengan verifier anggaran fixture sudah lulus. Uji dengan runtime Finance masih terbuka.

[Kontrak lengkap dan kasus uji](../../grpc/hrm/37-kebutuhan-finance-anggaran-requisition.md) · [Bukti fixture](../../e2e/gateway/175-hrm-mutation-approve-open-requisition-fixture-anggaran.md).
