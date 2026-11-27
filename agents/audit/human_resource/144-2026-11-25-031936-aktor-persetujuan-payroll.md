# Aktor persetujuan payroll

Tes aplikasi mereproduksi persetujuan payroll oleh actor workload ketika policy tiruan mengembalikannya. Command kini mewajibkan actor manusia dan tetap melarang self-approval. Tes membuktikan penolakan tidak mengubah status atau outbox, dan persetujuan manusia yang sah tetap berjalan dalam integrasi PostgreSQL sementara. Tes race dan vet bertag integration lulus. Tahap calculate, submit_review, dan request_posting masih tercantum sebagai rancangan dan belum memiliki command runtime.
