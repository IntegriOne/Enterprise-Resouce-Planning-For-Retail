# Pengujian Query pengajuan koreksi saldo cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Grant global baca koreksi, pengajuan belum diputuskan | Delta desimal, waktu persetujuan dan penolakan kosong |
| Grant baca ledger saja | FORBIDDEN |
| Koreksi dari hak cuti berbeda | UNAVAILABLE |

Tes integrasi tercakup dalam `TestHRMQueriesThroughGateway` dengan race detector. Uji HTTPS dengan grant resmi masih terbuka.

| Kasus pemulihan draf | Hasil |
| --- | --- |
| Filter kunci idempotensi cocok | Satu koreksi ditemukan |
| Kunci pada respons berbeda | UNAVAILABLE |
| Filter kosong atau kedua filter dikirim | INVALID_ARGUMENT |

Kasus pemulihan draf lulus dengan race detector.

Regresi Leave gRPC dengan PostgreSQL memverifikasi bahwa pembuatan ulang dengan kunci yang sama mengembalikan AlreadyExists. Finder mengembalikan tepat satu draf asli dengan ID dan hak cuti yang sama. Kunci yang tidak dikenal menghasilkan daftar kosong dan total nol. Race detector dan vet integrasi lulus.
