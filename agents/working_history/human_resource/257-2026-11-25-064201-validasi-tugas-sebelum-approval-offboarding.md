# Validasi tugas sebelum persetujuan offboarding

Rencana offboarding perlu minimal satu tugas sebelum dapat disetujui. Sebelumnya, rencana kosong bisa memperoleh persetujuan dan terkunci karena pembuatan tugas hanya diizinkan ketika rencana masih draf, sementara start mensyaratkan minimal satu tugas. Validator kini memeriksa jumlah tugas dalam transaksi persetujuan dan saat start.

Tes regresi validator dan tes gRPC dengan PostgreSQL sementara lulus menggunakan race detector. Tes gRPC membuktikan rencana kosong ditolak dan rencana yang sama dapat disetujui setelah tugas ditambahkan. `go vet ./...` lulus.

[Handoff offboarding](../../grpc/hrm/13-operasi-offboarding.md)
