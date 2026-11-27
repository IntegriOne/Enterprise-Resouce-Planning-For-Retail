# List operasi

Catatan untuk sinkronisasi saat Finance Accounting dibangun.

| Operasi atau peristiwa | Pemilik | Butuh GraphQL |
| --- | --- | --- |
| `RequestPayrollPosting` | HRM mengirim pengajuan dari snapshot yang sudah disetujui. | Mutation HRM |
| Konfirmasi jurnal dan penerbitan payroll | Finance memvalidasi pengajuan, membukukan jurnal, dan menerbitkan payroll. | Event dan verifikasi privat untuk sinkronisasi HRM |
| Eksekusi dan rekonsiliasi pembayaran | Finance mengirim instruksi pembayaran melalui bank atau provider, lalu mengonfirmasi hasil per pekerja. | Kebutuhan operasi Finance ditetapkan saat desain Finance |
| `GetMyPayslip` dan `ListMyPayslips` | HRM menyediakan akses pribadi setelah penerbitan Finance dan penyimpanan PDF di DRM. | `myPayslip`, `myPayslips`, dan `myPayslipPdf` |

## Penjelasan RequestPayrollPosting

HRM menyimpan satu event outbox `hr_payroll_runs.request_posting` v1 untuk run yang sudah disetujui. Status run tetap `approved` sampai HRM menerima bukti Finance. Envelope outbox memuat ID event untuk deduplikasi dan ID agregat yang mengidentifikasi run. Payload memuat `aggregate_id`, `record_key`, `change_kind`, dan `actor_principal_id`. Objek `details` memuat `payroll_period_id`, `payroll_group_id`, `employing_legal_entity_reference_id`, `currency_code`, `settlement_direction`, `input_snapshot_reference_id`, `input_digest_sha256`, dan `output_digest_sha256`.

Sebelum menulis event, HRM mewajibkan digest input dan output masing-masing 32 byte. Finance perlu mengambil dan memverifikasi nominal snapshot melalui kontrak privat yang disepakati. Event membawa referensi dan digest untuk korelasi. Bukti jurnal harus mengikat pengajuan ke run, badan hukum, dan digest output yang sama. Pengiriman ulang ID event yang sama harus menunjuk jurnal yang sama.

## Penjelasan penerbitan payroll dan slip

Finance mengirim konfirmasi penerbitan yang dapat diverifikasi dengan ID penerbitan, ID run, ID jurnal, badan hukum, ID snapshot input, digest SHA-256 input dan output, serta waktu penerbitan yang stabil. Bentuk awal event HRM adalah `payroll.published` v1. Sepakati nama, versi, autentikasi workload, pengiriman ulang, dan cara Finance menyediakan bukti untuk verifier HRM. Setelah bukti diterima, HRM mencocokkan periode, grup payroll, badan hukum, mata uang, dan yurisdiksi dengan payroll run. HRM kemudian menghasilkan PDF per pekerja dan menyimpan versi privat di DRM melalui adapter. Pekerja dapat mencetak slip setelah dokumen siap, termasuk saat pembayaran masih terjadwal.

## Penjelasan pembayaran

Finance Accounting memiliki instruksi transfer, penanganan respons bank atau provider, retry, dan rekonsiliasi. Sepakati konfirmasi per hasil pekerja yang dapat menyatakan pembayaran berhasil, gagal, atau masih diproses, dengan referensi Finance dan waktu perubahan. Jika pembayaran sebagian berhasil, hanya hasil pekerja yang berhasil ditandai dibayar. Verifier clearance offboarding menerima referensi pembayaran individual jika tersedia. HRM mengirim rincian dan mencatat status berdasarkan fakta Finance.

## Penjelasan clearance offboarding

Verifier HRM menerima ID rencana offboarding, pekerja, run terminasi, hasil payroll individual, badan hukum, mata uang, tanggal efektif terminasi, digest run dan hasil, serta referensi pembayaran ketika tersedia. Finance perlu menyediakan cara memeriksa bahwa keputusan clearance berlaku untuk kombinasi tersebut. Bukti yang merujuk pekerja, rencana atau hasil lain harus ditolak meskipun ID clearance ditemukan. Clearance IAM tetap diperiksa sesuai tanggung jawabnya.

## RBAC

Pengajuan HRM memakai izin `human_resource.payroll_runs.request_posting`. Otorisasi jurnal, penerbitan, dan pembayaran berada di Finance Accounting. Integrasi event dan verifier menggunakan identitas workload privat. Query slip pribadi memakai `human_resource.payslips.read_self` dengan scope `self`, lalu memverifikasi kepemilikan hasil dan akses DRM.

[Operasi payroll dan slip gaji](15-operasi-payroll.md) · [Handoff DRM](24-kebutuhan-integrasi-drm.md)

## Status runtime dan kriteria penerimaan

HRM sudah memiliki handler internal `HandlePayrollPublication`. Handler menerima `publication_id` dan `payroll_run_id` dari transport Finance tepercaya, memeriksa nama dan versi event, lalu memproses hasil payroll untuk run itu. HRM mengambil bukti publikasi lengkap melalui verifier. Startup standar masih memerlukan transport event dan adapter produksi untuk publikasi serta akses slip. [Handler publikasi](../../../backend/services/human_resource/infrastructure/inbox/payroll_publication.go) · [Registry adapter](../../../backend/services/human_resource/internal/service/registry.go).

Dalam uji integrasi lintas management, kirim publikasi yang sah sebelum pembayaran dijadwalkan atau dieksekusi sesuai data Finance. Pastikan HRM menyimpan satu versi PDF privat per hasil, termasuk saat event dikirim ulang setelah kegagalan sebagian. Bukti penerbitan dan referensi pembayaran divalidasi menurut tahapan masing-masing. [Regresi HRM dengan fixture Finance dan DRM](../../e2e/human_resource/54-query-mutation-regresi-grpc-postgresql.md).

## Pengiriman ulang outbox

Regresi relay dengan PostgreSQL membuktikan digest tetap valid setelah payload dibaca kembali dari JSONB. Jika publisher mengembalikan kegagalan sementara, relay menyimpan waktu percobaan berikutnya. Percobaan ulang memakai ID event yang sama, lalu mencatat `published_at` setelah berhasil. Penerima Finance perlu menolak pemrosesan ulang berdasarkan ID tersebut karena pengiriman dan commit status relay berada pada batas transaksi berbeda.

Uji integrasi menjalankan publisher webhook HRM dengan penerima lokal yang memberi HTTP 503 pada percobaan pertama dan HTTP 204 pada percobaan kedua. Kedua request memiliki isi dan kunci idempotensi yang identik serta tanda tangan HMAC-SHA256 yang valid. Tes ini memakai PostgreSQL khusus pengujian.

Verifikasi penerbitan HRM menerima ID snapshot dan kedua digest dari run yang sudah disetujui. Fixture Finance menolak ID snapshot atau digest input yang berbeda sebelum renderer dan penyimpanan PDF dipanggil. Stack Finance produksi masih diperlukan untuk membuktikan verifikasi terhadap bukti resmi.

Regresi inbox dengan PostgreSQL dan race detector memastikan sumber salah, event pembayaran, versi baru yang belum didukung, dan ID tidak valid ditolak sebelum penyimpanan PDF. Pada penerbitan yang sah, transaksi berjalan per hasil. Jika satu hasil gagal, HRM melanjutkan hasil lain selama konteks pengiriman aktif. Event tetap mengembalikan kegagalan, sehingga transport dapat mengulang pengiriman. PDF yang sudah tercatat dipertahankan pada pengulangan. Hasil yang terus gagal memerlukan koreksi agar seluruh run dapat diselesaikan.

HRM membawa ID run dari event sampai transaksi tiap hasil. Jika hasil tidak lagi cocok dengan run yang dikirim Finance, pemrosesan hasil tersebut ditolak sebelum verifikasi penerbitan dan penyimpanan PDF.

Kegagalan pemrosesan dilaporkan per hasil payroll pada jalur internal. HRM tetap mencoba hasil lain dalam event yang sama selama konteks aktif. Pengiriman ulang memproses hasil yang tertunda dan mempertahankan versi PDF yang sudah tercatat.
