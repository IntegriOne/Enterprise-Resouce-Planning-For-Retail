# Pengujian Mutation preboarding pekerja

Suite `TestHRMWorkerPreboardingMutationsThroughGateway` menggunakan sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Pendaftaran dengan grant create | preboarding; input RPC tanpa status, principal IAM, atau assignment |
| Nomor pekerja duplikat | CONFLICT |
| Pengosongan tanggal awal masa kerja dengan Relay ID | ID diteruskan sebagai UUID; mask hanya memuat tanggal yang dikosongkan |
| HRM menolak koreksi karena referensi terkait | CONFLICT |
| Grant read dipakai untuk update | FORBIDDEN |

Uji HTTPS dengan grant IAM resmi masih tertunda.
