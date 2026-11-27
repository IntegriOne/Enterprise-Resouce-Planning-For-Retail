# Uji gRPC cetak ulang slip

Suite payroll lulus dengan race detector, PostgreSQL disposable, dan renderer DejaVu Sans. Finance dan DRM memakai fixture.

| Kasus | Hasil |
| --- | --- |
| PDF sudah tertaut | Versi yang sama dikembalikan tanpa simpan ulang, dan audit tercatat |
| PDF belum tertaut dan respons penyimpanan DRM hilang | HRM mengembalikan `Unavailable` tanpa meninggalkan referensi parsial |
| Retry cetak ulang setelah respons penyimpanan hilang | PDF identik dan kunci idempotensi yang sama digunakan. Versi awal dikembalikan dan ditautkan |
| Event penerbitan diulang setelah versi tertaut | PDF tidak disimpan ulang |
| Akses mandiri pekerja lain ke hasil yang sama | Ditolak |

Kasus ini mencakup PDF yang belum tertaut. Pemulihan versi DRM yang sudah tertaut tetapi hilang masih terbuka. [Handoff cetak ulang](../../grpc/hrm/39-cetak-ulang-slip-gaji-mantan-karyawan.md).

Regresi pemulihan memakai DRM fixture yang mengonfirmasi blob hilang dengan metadata versi masih ada. HRM merender ulang snapshot dan mengirim byte beserta digest. Fixture hanya menerima byte yang identik dengan versi awal. RPC mengembalikan PDF dan ID versi lama tanpa mengganti referensi payroll. Suite PostgreSQL dan race detector lulus.

Audit `hr_payslip_reissues.recorded` mencatat `reused`, `created`, atau `restored` berdasarkan cabang yang berhasil di dalam transaksi publikasi. Regresi memeriksa satu event untuk masing-masing hasil pada skenario uji.

Perubahan identitas aktor saat verifikasi Finance kedua disimulasikan sebelum HRM menulis PDF. RPC mengembalikan `PermissionDenied` tanpa menambah panggilan simpan. Command memeriksa ulang otorisasi sebelum memanggil DRM dan sesudah respons DRM sebelum menautkan versi.

Regresi menolak pasangan pekerja dan hasil payroll berbeda tanpa menulis PDF. Untuk versi yang sudah tertaut, respons DRM `Unavailable`, `Forbidden`, dan `NotFound` umum tidak memicu pemulihan dan tidak mengembalikan isi PDF. Jalur pemulihan hanya berjalan ketika adapter mengembalikan konfirmasi khusus bahwa blob hilang dan metadata versi masih tersedia.
