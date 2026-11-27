# Identitas dan tanda terima pengajuan akses

Tes gRPC mereproduksi perubahan idempotency key pada pengajuan pending dan penghapusan pengajuan pending yang sudah memiliki tanda terima IAM. Validator kini menjaga idempotency key dan membatasi CRUD pada pengajuan pending tanpa referensi request maupun keputusan IAM. Koreksi dan penghapusan draf lokal tanpa tanda terima tetap teruji. Database sementara memakai fixture tanda terima, bukan layanan IAM nyata. Tes validator, integrasi dengan race detector, dan vet bertag integration lulus.
