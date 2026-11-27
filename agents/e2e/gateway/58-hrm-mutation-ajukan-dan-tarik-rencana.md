# Pengujian Mutation pengajuan dan penarikan rencana tenaga kerja

| Kasus gRPC fixture | Hasil |
| --- | --- |
| Pengajuan dengan grant submit dan MFA | review |
| Penarikan dengan grant cancel dan MFA | cancelled |
| Transisi tidak diizinkan HRM | CONFLICT |
| Grant submit dipakai untuk penarikan | FORBIDDEN |
| Penarikan tanpa MFA | FORBIDDEN |

Semua panggilan memakai CSRF. HRM memeriksa bahwa pelaku adalah pembuat rencana dan status awal memenuhi aturan. Persetujuan masih memerlukan verifier anggaran Finance.
