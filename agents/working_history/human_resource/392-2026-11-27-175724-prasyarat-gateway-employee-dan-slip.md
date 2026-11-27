# Prasyarat Gateway employee dan slip

Build terisolasi memetakan dependensi Gateway lokal yang belum tersedia di main. Commit `491a684` menambah paket bukti penautan akun pekerja dan menolak kunci pendek serta UUID kosong. Commit `aa87b28` membatasi izin baca slip mandiri pada grant self dan principal yang sama. Tes race dan vet lulus untuk kedua paket. Daftar sisa dependensi dicatat di katalog operasi HRM.
