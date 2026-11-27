# Pengujian Query kebijakan cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Detail dengan izin global baca kebijakan | Kode kebijakan dan jatah opsional kosong |
| Daftar dengan izin global baca kebijakan | Satu kebijakan dan total 1 |
| Daftar dengan izin baca pekerja saja | FORBIDDEN |

Ketiga kasus lulus dalam `TestHRMQueriesThroughGateway` dengan race detector. Uji HTTPS dengan akun dan grant resmi masih terbuka.
