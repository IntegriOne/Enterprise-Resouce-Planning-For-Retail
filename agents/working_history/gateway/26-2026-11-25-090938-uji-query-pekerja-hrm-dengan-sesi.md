# Pengujian Query pekerja HRM dengan sesi

Memverifikasi detail, Node, dan daftar pekerja dari endpoint GraphQL sampai Personnel gRPC fixture. Cookie kosong ditolak dan employee tanpa grant ditolak. Sesi fixture dengan grant badan hukum berhasil membaca ketiga jalur. Race test dan vet Gateway lulus. Pengujian memakai data terisolasi tanpa perubahan katalog atau grant IAM. [Matriks hasil](../../e2e/gateway/14-hrm-query-worker-node-dan-daftar.md).
