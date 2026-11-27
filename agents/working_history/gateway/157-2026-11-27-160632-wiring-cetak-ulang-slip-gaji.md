# Wiring cetak ulang slip gaji

Mutation GraphQL memanggil RPC privat cetak ulang. Gateway menuntut grant badan hukum dengan MFA, membatasi satu PDF per operasi, dan memvalidasi respons. Regresi Gateway dengan fixture gRPC dan race detector lulus. Adapter Finance dan DRM masih dibutuhkan untuk jalur sukses runtime.
