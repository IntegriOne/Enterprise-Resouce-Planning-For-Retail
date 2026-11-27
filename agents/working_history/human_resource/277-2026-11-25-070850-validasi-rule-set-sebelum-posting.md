# Validasi rule set sebelum pengajuan posting

Pengajuan posting menolak snapshot yang dihitung sebelum rule set disetujui. Validator memeriksa prasyarat sebelum perubahan dan event outbox tersimpan. Tes validator membuktikan penolakan dan jalur valid. Regresi gRPC PostgreSQL dengan race detector dan go vet lulus.
