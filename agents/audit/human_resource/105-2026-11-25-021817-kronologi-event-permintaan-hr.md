# Audit kronologi event permintaan HR

Validator menerima komentar dengan waktu lebih awal dari pembukaan tiket. Tes regresi mengonfirmasi celah tersebut. Validator kini membandingkan waktu event dengan waktu pembukaan tiket yang dikunci dalam transaksi, sehingga pergeseran jam antarreplika tidak dapat menulis riwayat sebelum tiket dibuka.
