# RPC respons penawaran terverifikasi

`AcceptOffer` dan `DeclineOffer` mencatat keputusan yang berasal dari verifier kandidat. Bukti terikat pada versi penawaran dan digest serta waktu yang sah. Tes gRPC–PostgreSQL dengan race detector memverifikasi penolakan bukti tak sah, penerimaan, penolakan, dan pencegahan keputusan ulang. Adapter kanal kandidat masih diperlukan.
