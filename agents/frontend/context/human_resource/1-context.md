# Konteks frontend Human Resource

## Apa yang sudah dikerjakan

- Workspace **SDM → Pekerja** menyediakan pencarian nomor pekerja, status hubungan kerja, tanggal masuk, dan pengajuan akun IAM. Detail menampilkan nama dan badan hukum dari Enterprise Foundation sesuai izin. ID teknis digunakan oleh GraphQL dan tidak ditampilkan pada area kerja.
- Status pengajuan dapat dibaca ulang setelah penautan IAM. Saat pengiriman gagal, data yang sudah tersimpan tetap tersedia untuk percobaan ulang. Saat status IAM belum bisa diverifikasi, halaman menyediakan penyegaran status. [Uji browser](../../e2e/human_resource/1-mutation-pengajuan-akun-pekerja.md).

## Apa yang belum dikerjakan

- Slip gaji mandiri memerlukan konfirmasi penerbitan Finance dan mekanisme baca PDF privat DRM. [Handoff Finance](../../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [Handoff DRM](../../../grpc/hrm/24-kebutuhan-integrasi-drm.md).

## Apa yang harus dikerjakan selanjutanya

- Setelah dependency tersedia, verifikasi pengguna bisa membuka dan mencetak slip sendiri dengan scope SELF melalui endpoint GraphQL yang sama.
