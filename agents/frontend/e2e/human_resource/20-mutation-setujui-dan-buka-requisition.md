# Uji browser persetujuan dan pembukaan requisition

Chrome mengakses stack disposable melalui HTTPS. Superadmin masuk dengan MFA dan memperoleh grant melalui GraphQL IAM. Dua unit dibuat di Enterprise Foundation. HRM memeriksa mandat melalui mTLS. Untuk uji ini saja, binary HRM memasang verifier anggaran fixture yang menghasilkan keputusan berhasil atau gangguan sesuai skenario.

| Kasus | Hasil |
| --- | --- |
| Buka pengajuan dari unit tanpa grant spesifik | Aksi Setujui tidak tampil meski superadmin memiliki role GLOBAL. |
| Ajukan persetujuan saat verifier anggaran tidak tersedia | Pesan verifikasi muncul dan status tetap Menunggu persetujuan. |
| Setujui pengajuan yang memenuhi syarat | Tabel menampilkan status Disetujui. |
| Buka requisition dengan mandat dan bukti anggaran yang sama | Tabel menampilkan status Dibuka. Tombol Buka hilang. |
| Periksa detail | Teks tidak memuat `gid://`. |

TypeScript, lint direktori terkait, dan build lulus. Alur positif memakai fixture anggaran. Finance produksi masih diperlukan.
