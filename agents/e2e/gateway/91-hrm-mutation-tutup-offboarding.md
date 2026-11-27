# Pengujian Mutation penutupan offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Payroll final dan clearance sesuai | completed |
| RPC menolak payroll final yang tidak cocok | CONFLICT |
| Verifier clearance tidak tersedia | UNAVAILABLE |
| Format ID clearance salah | INVALID_ARGUMENT |
| Grant start dipakai untuk menutup | FORBIDDEN |
| Sesi tanpa MFA | FORBIDDEN |

Pengujian memakai sesi employee, CSRF, dan fixture gRPC. Uji positif runtime memerlukan verifier clearance Finance dan IAM.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:4de366a5252289fe57ce1c9a5503b3c4a4c8fb9c4e2cd9280c494f44637eebe8`. Ketiga mutation penutupan rencana, penyelesaian tugas dan pengecualian tugas ditolak dengan `FORBIDDEN` tanpa token CSRF melalui Caddy. Suite race dan vet Gateway lulus.
