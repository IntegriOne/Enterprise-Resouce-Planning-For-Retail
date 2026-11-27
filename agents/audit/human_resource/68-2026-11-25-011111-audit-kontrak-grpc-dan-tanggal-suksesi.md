# Audit kontrak gRPC dan tanggal suksesi

- Pemeriksaan statis membandingkan 966 deklarasi RPC dengan nama method handler pada 22 domain. Setiap deklarasi memiliki padanan nama di direktori handler domain. RegisterAll memuat 22 registrasi domain. Pemeriksaan ini mengukur kelengkapan struktural dan masih perlu dilengkapi dengan verifikasi perilaku tiap workflow.
- Validator nominasi suksesi sebelumnya mengambil tanggal kalender bukti TIMESTAMPTZ menurut zona yang dibawa timestamp. Bukti pukul 23.00 UTC dapat terbaca sebagai pukul 06.00 WIB pada tanggal berikutnya dan ditolak. Validator kini membandingkan tanggal kesiapan dengan hari UTC untuk finalisasi asesmen, persetujuan rencana pengembangan, dan hari berjalan.
- Tes regresi gagal sebelum perubahan dengan `hrm.reference_conflict`. Tes unit seluruh modul dengan race detector dan `go vet` lulus.
- Persetujuan rencana suksesi masih belum memiliki command pada kontrak. Jalur verifikasi bukti DRM dan posting Finance membutuhkan API pemilik. Audit perilaku tiap workflow masih berlanjut.

## Command yang belum dipetakan dari matriks izin HRM

Perbandingan kode izin rancangan dengan katalog runtime menunjukkan kelompok berikut. Izin persetujuan pengajuan akses sengaja dikecualikan karena keputusan dan grant milik IAM.

| Domain | Workflow yang masih perlu verifikasi atau implementasi |
| --- | --- |
| Workforce Planning | Submit, approve dengan bukti anggaran, supersede dan cancel. |
| Personnel | Link principal, suspend, terminate, transfer penempatan, approval/aktivasi/akhir kontrak dan verifikasi/penolakan dokumen pekerja. |
| Recruitment | Persetujuan, penerbitan, penarikan, penerimaan dan penolakan offer. |
| Offboarding | Persetujuan, mulai dan penyelesaian rencana, serta penyelesaian dan waiver tugas. |
| Compensation dan Benefits | Submit, approve, reject dan aktivasi kompensasi, serta persetujuan dan akhir enrollment benefit. |
| Payroll dan Leave | Kalkulasi payroll, submit review dan permintaan posting Finance, serta penolakan koreksi dan penutupan hak cuti. |
| Learning dan Succession | Penyelesaian enrollment, verifikasi kualifikasi, approval rencana suksesi, nominasi dan penarikan kandidat sebagai command tersendiri. |
| Employee Relations dan Career | Penyelesaian kasus dan persetujuan rencana pengembangan. |
| Policy, HR Service dan Engagement | Publikasi versi kebijakan, penugasan/resolusi/penutupan permintaan, publikasi dan penutupan survei. |
| Access Request dan Talent | Submit/recommend pengajuan akses dan penutupan sesi kalibrasi. |

Daftar ini memakai [matriks HRM](../../roles-permissions/human_resource/1-roles-permissions.md) dan [katalog runtime](../../../backend/services/human_resource/infrastructure/security/catalog.go). Untuk mengaktifkan approval suksesi perlu bukti mandat panel dari scope otoritatif dan aturan konflik kepentingan yang dapat diuji. Tabel rencana saat audit menyimpan approver dan waktu approval, sedangkan identitas pembuat tidak tersedia pada baris rencana.

## Verifikasi integrasi

Build modul lulus. Eksekusi integrasi dengan race detector menghasilkan 64 paket lulus. Penjaga dua paket menolak nama database pada eksekusi pertama karena perlu prefiks `hrm_disposable_`. Kedua paket tersebut, Succession dan Workforce Planning, lulus saat diulang dengan database sementara bernama sesuai kontrak. Kedua database sementara telah dibersihkan. Tes integrasi suksesi menguji RPC rencana suksesi. Regresi zona waktu bukti kandidat diuji langsung pada validator.
