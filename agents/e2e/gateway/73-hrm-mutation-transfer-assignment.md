# Pengujian Mutation transfer assignment pekerja

| Kasus fixture | Hasil |
| --- | --- |
| Pengajuan dengan Relay ID pekerja dan kunci idempotensi | submitted |
| Pengulangan pengajuan yang sudah diterapkan | applied dengan ID assignment hasil |
| Pengulangan pengajuan yang sudah ditolak | rejected dengan ID assignment hasil null |
| Respons replay applied tanpa ID assignment hasil | UNAVAILABLE |
| Konflik kunci dari HRM | CONFLICT |
| Keputusan approve | applied dengan ID assignment hasil |
| Keputusan reject | rejected dengan ID assignment hasil null |
| Keputusan tanpa MFA | FORBIDDEN |
| Grant create dipakai untuk memutuskan | FORBIDDEN |

Uji HTTPS dengan grant IAM resmi masih tertunda.
