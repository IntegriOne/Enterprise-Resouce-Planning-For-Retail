# Penerbitan payroll Finance dan slip gaji mandiri

HRM mengajukan snapshot payroll yang sudah disetujui dan dibekukan. Finance Accounting mencatat kewajiban, menerbitkan payroll, mengendalikan otorisasi transfer, dan merekonsiliasi hasil pembayaran per pekerja. HRM memproyeksikan status yang dibuktikan Finance.

| Tahap | Kontrak yang diperlukan dari Finance |
| --- | --- |
| Penerimaan pengajuan | Konsumsi `hr_payroll_runs.request_posting` secara idempoten berdasarkan ID event. Korelasikan run, badan hukum, grup, periode, mata uang, arah settlement dan digest snapshot. |
| Posting jurnal | Kembalikan bukti jurnal yang terikat pada pengajuan dan snapshot. HRM memerlukan status `posted` atau `paid` sebelum dapat menautkan slip. |
| Penerbitan | Kirim `payroll.published` versi 1 dengan ID penerbitan dan ID run. Identitas Finance harus berasal dari transport privat terautentikasi. Simpan bukti penerbitan agar dapat dibaca ulang. |
| Verifikasi penerbitan | Buktikan kecocokan ID penerbitan, run, jurnal, badan hukum, ID snapshot input, dan digest input serta output. Berikan waktu penerbitan otoritatif. |
| Pembayaran dan rekonsiliasi | Berikan status dan referensi tepercaya per hasil pekerja. Settlement agregat memerlukan konfirmasi final Finance. |

## Urutan dan pengulangan

- ID penerbitan adalah identitas bisnis yang tetap saat event dikirim ulang. ID pengiriman dan ID event pengajuan punya fungsi korelasi tersendiri.
- Bila konfirmasi penerbitan tiba sebelum posting terproyeksi di HRM, integrasi perlu menunda atau mengirim ulang setelah bukti posting tersedia.
- HRM memproses satu hasil per transaksi. Jika sebagian gagal, retry dengan ID penerbitan yang sama melanjutkan hasil yang belum tertaut. DRM perlu mengembalikan versi yang sama untuk permintaan simpan yang sama.
- Setelah versi PDF privat tersedia, pekerja bisa membuka dan mencetaknya dengan izin SELF. Konfirmasi pembayaran individual menentukan label pembayaran dan dapat menyusul setelah slip terbit.

## Kebutuhan sinkronisasi

Sepakati transport privat, kontrak pembacaan bukti, data agregat jurnal, dan mekanisme pemulihan event yang datang di luar urutan. Handler internal HRM sudah mengenali event penerbitan versi 1. Transport Finance dan penyimpanan DRM runtime masih memerlukan integrasi.

[Operasi payroll HRM](../../grpc/hrm/15-operasi-payroll.md) · [Kontrak PDF DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md) · [Handler penerbitan](../../../backend/services/human_resource/infrastructure/inbox/payroll_publication.go).

Regresi pada PostgreSQL disposable mengirim event yang valid saat run masih menunggu posting. Handler mengembalikan `ErrWorkflowRequired` dengan nol panggilan penyimpanan PDF. Setelah posting dan verifikasi tersedia melalui fixture, pengiriman ulang dengan identitas penerbitan yang sama menyelesaikan slip yang tertunda. Transport perlu mempertahankan event pada kondisi ini. Uji handler belum mencakup kebijakan retry pada transport Finance.

Untuk cetak ulang oleh HR saat PDF belum pernah tertaut, HRM juga memerlukan penemuan ID penerbitan Finance berdasarkan run, jurnal, badan hukum dan digest snapshot. ID itu diperlukan sebelum pemroses yang ada dapat memverifikasi penerbitan dan menyimpan PDF. [Handoff cetak ulang](../../grpc/hrm/39-cetak-ulang-slip-gaji-mantan-karyawan.md).

## Konfirmasi jurnal yang diproyeksikan HRM

HRM menyediakan `HandlePayrollPosting` untuk kandidat event privat `payroll.posted` versi 1. Transport perlu memasok identitas Finance yang sudah diautentikasi, run, dan ID jurnal. Sebelum mengubah status, `ConfirmPayrollPosting` meminta verifier Finance mencocokkan jurnal dengan run yang diajukan, waktu pengajuan, badan hukum, periode, grup, mata uang, arah settlement, ID snapshot, dan digest input serta output. Finance mengembalikan waktu posting otoritatif. HRM mengunci run, menyimpan status `posted` dan ID jurnal, serta menerima pengulangan jurnal yang sama tanpa mengganti referensi. Finance tetap memiliki jurnal dan transfer. Kontrak transport dan provider Finance masih menunggu sinkronisasi. [Handler HRM](../../../backend/services/human_resource/infrastructure/inbox/payroll_posting.go)

Waktu penerbitan yang dikembalikan verifier harus setidaknya sama dengan waktu pengajuan posting HRM apabila tercatat. Finance juga perlu membuktikan bahwa penerbitan mengikuti posting jurnal aktual. HRM menyimpan referensi jurnal dan status posting, sedangkan waktu posting otoritatif diperiksa saat proyeksi. Uji PostgreSQL menolak penerbitan di antara persetujuan dan pengajuan posting tanpa menulis PDF.

HRM juga menyediakan `HandlePayrollPayment` untuk kandidat event privat `payroll.result_paid` versi 1. Verifier Finance perlu membuktikan penyelesaian individual berdasarkan ID hasil, pekerja, pembayaran, run dan jurnal. Keputusan harus cocok dengan badan hukum, mata uang, arah settlement, gaji bersih, dan digest hasil. Konfirmasi satu hasil memperbarui status dan referensi pembayaran pekerja. Status agregat menunggu bukti settlement run. [Handler pembayaran](../../../backend/services/human_resource/infrastructure/inbox/payroll_payment.go)

`HandlePayrollSettlement` menerima kandidat event privat `payroll.settled` versi 1. HRM memeriksa bahwa setiap hasil sudah memiliki konfirmasi individual dan verifier Finance membuktikan bahwa batch serta event final mencakup jurnal, digest run, dan seluruh pasangan ID hasil serta referensi pembayaran. HRM menyimpan status agregat `paid`, ID batch, ID event settlement, dan waktu settlement. Sinkronisasi perlu menetapkan bagaimana Finance menyediakan referensi penyelesaian tepercaya untuk hasil yang tidak melakukan transfer, misalnya nominal nol. [Handler settlement](../../../backend/services/human_resource/infrastructure/inbox/payroll_settlement.go)

Implementasi provider dilanjutkan saat Finance dibangun. Pasang verifier jurnal, penerbitan, pembayaran individual, dan settlement pada startup HRM. Finance memiliki penerbitan dan transfer. Uji pengiriman ulang event yang sama, kegagalan sementara pada sebagian PDF, dan korelasi referensi dengan snapshot HRM. [Titik masuk event](../../grpc/hrm/41-penerimaan-event-payroll-finance.md)
