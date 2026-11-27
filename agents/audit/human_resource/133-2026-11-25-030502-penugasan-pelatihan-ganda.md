# Penugasan pelatihan ganda

Tes gRPC dan PostgreSQL sementara mereproduksi dua enrollment aktif untuk pekerja dan kursus yang sama dengan idempotency key berbeda. Create kini memeriksa status planned, enrolled, dan in_progress sambil memegang lock worker dan kursus. Penugasan aktif kedua ditolak dengan AlreadyExists. Enrollment historis berstatus terminal tidak ikut membatasi penugasan baru. Integrasi Learning, tes validator dengan race detector, dan vet modul lulus.
