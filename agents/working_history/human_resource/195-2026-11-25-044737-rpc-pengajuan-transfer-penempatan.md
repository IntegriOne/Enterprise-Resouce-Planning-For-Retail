# RPC pengajuan transfer penempatan

`SubmitWorkerAssignmentTransfer` tersedia pada gRPC Personnel privat. Service memeriksa penempatan primer sumber dan posisi tujuan dalam badan hukum yang sama serta mencatat principal dari policy. Retry identik mengembalikan record yang sama dan perubahan payload ditolak. Integrasi gRPC–PostgreSQL dengan race detector membuktikan pengajuan dan retry tanpa mutasi penempatan sumber. Go test dan vet modul lulus. Keputusan dan penerapan transfer masih terbuka.
