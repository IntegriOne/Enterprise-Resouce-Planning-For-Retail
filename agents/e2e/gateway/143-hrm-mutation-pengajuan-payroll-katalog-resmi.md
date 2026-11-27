# Pengajuan payroll dengan katalog IAM resmi

Uji memakai dua employee dengan MFA pada stack Caddy HTTPS terisolasi. Superadmin memasang role dan permission lewat GraphQL. Operator kedua dibuat melalui pengajuan HRM dan provisioning IAM. Artefak aturan, kontrak bertanda tangan, saldo payroll terhitung, serta verifikasi identitas uji disiapkan sebagai fixture.

| Skenario | Hasil |
| --- | --- |
| Penginisiasi membuat draft dan mengirim run untuk review | Berhasil. |
| Penginisiasi menyetujui run yang sama | Ditolak oleh policy HRM. |
| Operator kedua menyetujui dan mengajukan posting | Berhasil. Run tetap approved. |
| Permintaan posting diulang | `CONFLICT`. |
| Pemeriksaan PostgreSQL | Satu event outbox dengan ID snapshot input dan digest SHA-256 input serta output yang sesuai. |

Percobaan awal meminta field payload yang salah. Runner disesuaikan ke `payroll_posting_request` lalu lulus. HRM development juga diperbarui dan lulus healthcheck serta uji baca HTTPS. Uji posting memakai stack disposable. Penerbitan, jurnal, dan pembayaran menunggu implementasi Finance.

Regresi diulang dengan image Compose aktif setelah perbaikan verifikasi snapshot penerbitan. Uji HTTPS membuktikan persetujuan independen, pengajuan posting sekali, serta penyimpanan ID snapshot dan kedua digest pada event outbox. Uji ini berhenti pada pengajuan. Pemeriksaan isi bukti penerbitan memakai verifier fixture pada regresi gRPC dan PostgreSQL.
