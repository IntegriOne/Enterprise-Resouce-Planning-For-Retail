# Pengujian Mutation penyelesaian dan pengecualian tugas offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Penyelesaian dengan bukti terverifikasi | completed |
| Verifier bukti tidak tersedia | UNAVAILABLE |
| Format ID bukti salah | INVALID_ARGUMENT |
| Pengecualian dengan alasan | waived |
| Alasan kosong | INVALID_ARGUMENT |
| Grant complete dipakai untuk waive | FORBIDDEN |
| Sesi tanpa MFA | FORBIDDEN |

Uji memakai sesi employee, CSRF, dan fixture gRPC. Penyelesaian runtime masih memerlukan verifier konfirmasi.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:4de366a5252289fe57ce1c9a5503b3c4a4c8fb9c4e2cd9280c494f44637eebe8`. Ketiga mutation penutupan rencana, penyelesaian tugas dan pengecualian tugas ditolak dengan `FORBIDDEN` tanpa token CSRF melalui Caddy. Suite race dan vet Gateway lulus.
