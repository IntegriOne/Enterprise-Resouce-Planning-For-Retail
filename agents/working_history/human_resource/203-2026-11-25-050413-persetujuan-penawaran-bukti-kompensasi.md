# Persetujuan penawaran dengan bukti kompensasi

`ApproveOffer` memisahkan pembuat dan penyetuju serta mengikat versi kompensasi terverifikasi pada penawaran, lamaran, pekerjaan, badan hukum, dan mata uang. Status, persetujuan, dan outbox disimpan atomik. Tes gRPC–PostgreSQL dan race detector memakai verifier tiruan. Adapter DRM belum tersedia sehingga runtime menolak approval yang membutuhkan bukti eksternal.
