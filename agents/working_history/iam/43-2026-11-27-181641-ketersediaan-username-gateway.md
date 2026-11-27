# Ketersediaan username untuk Gateway

Commit `ffe8f7d` mempublikasikan kontrak dan handler IAM untuk ketersediaan username. Hasil hanya memuat status dan username ternormalisasi. Tes gRPC membuktikan respons akun employee dan customer yang sudah mengambil username konsisten tanpa atribut akun. Lint kontrak, race suite terkait, dan vet lulus. Commit `87956c3` juga mempublikasikan modul katalog RBAC deklaratif tanpa otomatis memasukkan izin HRM ke baseline superadmin.
