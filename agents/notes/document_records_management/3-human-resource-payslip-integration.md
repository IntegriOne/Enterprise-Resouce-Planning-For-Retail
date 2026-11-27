# Integrasi DRM untuk HRM

Pekerjaan provider dijadwalkan saat pengembangan DRM. HRM telah menyediakan port untuk verifikasi bukti, penyimpanan, pembacaan, dan pemulihan versi PDF.

- Setelah Finance mengonfirmasi penerbitan, DRM menyimpan PDF privat dan immutable. Gunakan kunci idempotensi penerbitan dan hasil payroll, serta cocokkan digest SHA-256.
- Karyawan aktif memakai grant SELF untuk membuka slip sendiri. Mantan karyawan meminta petugas HR dengan grant badan hukum dan MFA. Jika PDF belum dibuat, jalur petugas membuatnya dari hasil terbit.
- Jika blob hilang tetapi metadata versi masih ada, pemulihan mempertahankan ID versi dan mewajibkan digest identik. Respons DRM harus membedakan keadaan tersebut dari dokumen tidak ditemukan biasa, akses ditolak, dan layanan tidak tersedia.
- Jadwalkan juga verifikasi versi kontrak bertanda tangan, dokumen rekrutmen, bukti onboarding dan HR Service, serta artefak aturan payroll.

[Titik pemasangan dan kontrak](../../grpc/hrm/24-kebutuhan-integrasi-drm.md) · [Cetak ulang](../../grpc/hrm/39-cetak-ulang-slip-gaji-mantan-karyawan.md)
