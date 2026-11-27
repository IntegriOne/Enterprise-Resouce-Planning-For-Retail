# Pengujian Query detail ledger cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| ID ada dan izin baca ledger | Urutan int64 dan delta desimal dipertahankan |
| ID tidak ditemukan | NOT_FOUND |
| Izin baca hak cuti saja | FORBIDDEN |

Tes detail dan regresi daftar tercakup dalam `TestHRMQueriesThroughGateway`. Uji HTTPS dengan grant resmi masih terbuka.
