# Pengujian Query riwayat saldo cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Izin global baca ledger | Urutan int64 dan delta saldo desimal dipertahankan sebagai string |
| Izin baca hak cuti saja | FORBIDDEN |
| Respons gRPC berisi entri dari hak cuti berbeda | UNAVAILABLE |

Kasus integrasi tercakup dalam `TestHRMQueriesThroughGateway` dengan race detector. Uji HTTPS dengan grant resmi masih terbuka.

| Kasus filter koreksi | Hasil |
| --- | --- |
| Koreksi yang sudah dibukukan | Entri dengan ID koreksi yang sama dan total 1 |
| Koreksi tanpa entri ledger | Daftar kosong dan total nol |
| ID hak cuti dan ID koreksi dikirim bersama | INVALID_ARGUMENT |

Regresi gRPC HRM dengan PostgreSQL membuktikan finder mengembalikan total nol setelah persetujuan dan sebelum pembukuan. Setelah pembukuan dan percobaan ulang, finder mengembalikan tepat satu entri jenis adjustment yang merujuk ke koreksi dan hak cuti yang benar. Race detector dan vet integrasi lulus.
