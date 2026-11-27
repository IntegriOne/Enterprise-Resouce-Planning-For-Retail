# Benturan jadwal wawancara lintas lamaran

Tes gRPC dengan PostgreSQL sementara mereproduksi jadwal bertumpang tindih antara dua lamaran kandidat yang sama. Validator sekarang memeriksa seluruh jadwal aktif kandidat menggunakan kueri berbatas satu hasil di dalam transaksi yang telah mengunci kandidat. Benturan hanya mengembalikan konflik generik. Tes regresi Recruitment juga mencakup slot bersebelahan, RPC serentak pada satu lamaran, dan penolakan reschedule yang bertumpang tindih.
