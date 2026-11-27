# Pengajuan akun pekerja dari workspace HRM

## Lingkungan

Chrome headless melalui Caddy HTTPS dengan Gateway, HRM, IAM dan Enterprise Foundation. Pengujian menggunakan PostgreSQL dan Redis terisolasi. Role sementara dibuat, diberi permission, dan dipasang ke superadmin lewat GraphQL. Data dasar pekerja dan peristiwa perekrutan untuk pengujian disiapkan sebagai fixture di database HRM terisolasi.

## Skenario

| Skenario | Hasil |
| --- | --- |
| Petugas membuka daftar, memilih pekerja, dan mengajukan akun | Browser menampilkan nomor pekerja dan status Menunggu proses IAM. Pemeriksaan teks area kerja tidak menemukan awalan `gid://`. |
| Pengajuan sebelum pekerja memiliki peristiwa perekrutan terverifikasi | GraphQL menolak pengajuan. |
| Petugas mengirim ulang pengajuan | GraphQL mengembalikan referensi pengajuan yang sama. |
| IAM membuka antrean sebelum MFA atau mengirim referensi pekerja yang berbeda | GraphQL menolak akses atau perubahan. |
| IAM menerbitkan akun lalu mengulang mutation | GraphQL mengembalikan akun yang sama. |
| Petugas memeriksa status setelah IAM menautkan akun | GraphQL mengembalikan status `applied`. |
| Petugas membuka ulang halaman setelah penautan | Browser menampilkan Akun tertaut. Pemeriksaan teks area kerja kembali tidak menemukan awalan `gid://`. |
| Pemeriksaan data hasil pengujian | IAM menyimpan satu akun dan satu pengajuan. Principal yang ditautkan HRM cocok dengan akun IAM tersebut. |

Pembukaan ulang pada percobaan awal berhenti di indikator muat. Pengujian ulang pada sesi browser baru dengan cache browser dinonaktifkan berhasil untuk kedua tahap.

Pengujian ulang pada build frontend terbaru juga memeriksa baris pekerja saat pengajuan dan setelah penautan. Pada kedua tahap, label kerja tampil sebagai **Aktif**, tanggal masuk tampil sebagai **1 November 2026**, dan area kerja tetap tidak menampilkan `gid://`.

Tes klien GraphQL juga mencakup respons yang menyertakan pengajuan tersimpan dan status gagal kirim. Klien mempertahankan status pending sehingga halaman dapat menawarkan pengiriman ulang. Hasil ini berasal dari respons fixture pada tes unit. Skenario gangguan pengiriman belum dijalankan di browser.

## Pengiriman ulang dari browser

Gateway uji mengembalikan `UNAVAILABLE` satu kali setelah pengajuan tersimpan di HRM dan sebelum dikirim ke IAM. Browser menampilkan **Pengiriman ke IAM belum terkonfirmasi** dan **Pengajuan tersimpan**. Tombol **Kirim ulang ke IAM** berhasil. Setelah IAM memprosesnya, browser HRM menampilkan **Akun tertaut** tanpa ID teknis pada area kerja. Pemeriksaan PostgreSQL menemukan satu pengajuan HRM, satu penerimaan IAM, dan satu akun. Build Gateway yang menyimulasikan kegagalan hanya dipakai pada stack terisolasi.

Skenario berikutnya menyimulasikan kegagalan pembacaan status IAM setelah percobaan pengiriman awal. Browser tetap menampilkan pengajuan lokal dan pemberitahuan bahwa status IAM belum bisa diperiksa. Pengiriman ulang tidak ditawarkan selama status itu belum terverifikasi. Setelah refresh berhasil, browser menampilkan status pengiriman pending dan menyediakan tombol kirim ulang. Satu permintaan HRM dan satu penerimaan IAM tersimpan. Build normal Gateway dan frontend kemudian dipasang di development, kedua replika sehat, dan uji baca HTTPS superadmin berhasil.

## Pesan kegagalan daftar pekerja

Chrome headless membuka workspace melalui HTTPS pada stack terisolasi menggunakan sesi superadmin. Browser menerima satu respons GraphQL simulasi dengan `gid://` dalam pesan kesalahan daftar pekerja. Halaman menampilkan “Daftar pekerja gagal dimuat.” dan tombol “Muat ulang daftar”. Pemeriksaan teks area kerja tidak menemukan `gid://`. Setelah tombol ditekan, browser memuat baris pekerja dari Gateway uji dan melanjutkan pengajuan akun. Setelah persetujuan IAM dengan MFA, browser kembali menampilkan “Akun tertaut”, dengan satu akun dan satu penerimaan pengajuan tersimpan.

## Nama pekerja dan badan hukum pada daftar

Chrome membaca nama pekerja dan badan hukum dari Enterprise Foundation melalui HTTPS pada stack terisolasi. Label muncul pada baris sebelum detail dibuka. Uji kedua menyimulasikan penolakan kedua lookup. UI menampilkan keterangan nama tidak tersedia dan status akun tertaut tetap terbaca. Kedua fase tidak menampilkan `gid://`. Wiring IAM–HRM menyimpan satu akun dan satu penerimaan.
