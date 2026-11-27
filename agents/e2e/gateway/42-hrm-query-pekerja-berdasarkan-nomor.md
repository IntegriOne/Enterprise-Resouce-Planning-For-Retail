# Pengujian Query pekerja berdasarkan nomor

| Kasus Personnel gRPC fixture | Hasil |
| --- | --- |
| Nomor cocok dan grant badan hukum, total berubah sesudah pembacaan item | Satu pekerja yang ditemukan dan total terbaru 2 |
| Nomor hanya berisi spasi | INVALID_ARGUMENT |
| RPC mengembalikan nomor berbeda | UNAVAILABLE |

Kasus baru dan regresi daftar dicakup oleh `TestHRMQueriesThroughGateway` dengan race detector.
