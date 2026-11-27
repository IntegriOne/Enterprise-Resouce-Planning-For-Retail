# Perbaikan paginasi HRM saat data berubah

HRM biasa membaca item dan jumlah dengan dua pernyataan. Gateway sebelumnya mengembalikan UNAVAILABLE bila keduanya berubah di tengah permintaan. Validasi kini memeriksa limit, offset dan rentang nilai sambil meneruskan total HRM. Daftar slip terbit tetap memakai kontrak total yang ketat. Regresi worker menguji halaman berisi satu item saat total terbaru dua.
