# Uji penerbitan slip serentak

Regresi PostgreSQL mengamati pemroses kedua menunggu kunci run saat simpan pertama belum selesai. Kedua pemroses menerima versi PDF yang sama dengan satu penyimpanan. Tes lulus dengan race detector, termasuk saat renderer PDF digunakan.
