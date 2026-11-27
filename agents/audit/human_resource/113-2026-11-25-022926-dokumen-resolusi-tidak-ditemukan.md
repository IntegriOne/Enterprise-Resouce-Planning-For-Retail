# Dokumen resolusi tiket yang tidak ditemukan

Tes gRPC dengan PostgreSQL sementara mereproduksi respons Unavailable ketika verifier DRM tiruan melaporkan dokumen tidak ditemukan. Resolusi kini memakai klasifikasi error verifikasi bersama dan mengembalikan FailedPrecondition untuk kasus tersebut. Skenario yang sama melanjutkan dengan dokumen sah, menutup tiket, dan memeriksa jumlah event outbox.
