# Pengujian integrasi Query baris rencana tenaga kerja

Gateway menerima POST GraphQL dengan sesi employee fixture dan memakai Workforce Planning gRPC melalui bufconn.

| Kasus | Hasil |
| --- | --- |
| Detail dengan grant baca baris berscope badan hukum | ID rencana induk dan nilai FTE desimal |
| Daftar dengan filter ID induk | Satu baris dari RPC finder dan total 1 |
| Daftar tanpa filter | Halaman kosong dan total 0 dari RPC daftar |
| Daftar dengan grant baca rencana induk saja | `FORBIDDEN` |

Keempat kasus lulus dalam `TestHRMQueriesThroughGateway`. Tes ini memverifikasi jalur GraphQL ke RPC fixture. Cakupan badan hukum dengan data HRM dan akun IAM resmi perlu verifikasi HTTPS.
