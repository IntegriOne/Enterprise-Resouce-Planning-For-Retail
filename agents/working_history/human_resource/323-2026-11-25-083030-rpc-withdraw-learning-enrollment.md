# RPC withdraw learning enrollment

Menambahkan WithdrawLearningEnrollment untuk enrollment planned, enrolled atau in_progress. Transisi mencatat outbox, mempertahankan riwayat mulai dan menolak enrollment terminal. Tes gRPC PostgreSQL membuktikan withdrawal, penolakan pengulangan dan penolakan start setelah withdrawal. Inventaris mencocokkan 1.010 RPC dengan kontrak Gateway.
