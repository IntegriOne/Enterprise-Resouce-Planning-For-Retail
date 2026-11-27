# Pengujian Query detail tugas offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Sesi employee dengan grant global baca tugas | Tugas berstatus pending |
| ID tugas tidak ditemukan oleh HRM | NOT_FOUND |
| ID respons berbeda dari ID permintaan | UNAVAILABLE |
| ID permintaan tidak valid | INVALID_ARGUMENT |
| Sesi hanya memiliki grant baca rencana | FORBIDDEN |

Uji Gateway melalui fixture gRPC lulus dengan race detector untuk Query detail dan daftar tugas.

## HTTPS lokal

Kedua replika sehat pada image `sha256:b845e0af37486ed202cf77543a36eebe66597852c5b6f9103014f1d799c79aa4`. Caddy HTTPS mengenali Query detail tugas dan menolak permintaan anonim dengan `UNAUTHENTICATED`. Uji berakun dengan grant resmi IAM masih tertunda. Vet Gateway lulus.
