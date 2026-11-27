# Uji browser penetapan kepala unit dengan nomor pekerja

Chrome headless membuka workspace HRM melalui Caddy pada stack terisolasi. Superadmin membuat unit EF serta pekerja HRM yang akunnya ditautkan IAM. Izin uji diterapkan melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Membuka detail unit | Kode dan nama unit tampil |
| Mengisi nomor pekerja dan memilih Tetapkan | UI mencari pekerja dan mengirim GID pekerja ke mutation EF |
| Pembacaan unit setelah aksi dari browser | UUID yang tersimpan cocok dengan pekerja yang dicari |
| Pemeriksaan teks halaman | Tidak ditemukan `gid://` atau label UUID Raw |
| Mengosongkan kepala unit lewat GraphQL | Referensi kembali kosong |

TypeScript, ESLint pada file yang diubah dan build frontend lulus.

Form penyuntingan detail mengirim `preserve_head_person: true` agar nama dan rincian dapat disimpan tanpa mengubah penetapan kepala. [Verifikasi kontrak HTTPS](../../../e2e/gateway/161-ef-mutation-perbarui-unit-pertahankan-kepala.md).

Setelah penetapan, Chrome juga memverifikasi panel menampilkan nama dari party Enterprise Foundation dan nomor pekerja dari HRM. Pemeriksaan teks halaman tetap tidak menemukan `gid://`. Jika nama party tidak tersedia, panel memakai nomor pekerja bila diizinkan. Jika detail pekerja tidak tersedia, panel menampilkan teks generik.
