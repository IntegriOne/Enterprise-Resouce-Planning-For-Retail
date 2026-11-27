# Validasi payroll final offboarding

Penutupan rencana kini mengunci run payroll HRM dan mensyaratkan jenis termination dengan metadata persetujuan serta status approved, posted, atau paid. Run yang salah tipe, belum disetujui, atau tidak memuat satu hasil pekerja yang sama ditolak sebelum verifikasi clearance. Keputusan Finance dan IAM tetap dibuktikan melalui port clearance. Kebutuhan kontrak dicatat pada catatan owner masing-masing.

Tes validator mencakup terminasi yang disetujui dan dibayar, payroll reguler, terminasi yang masih calculated serta metadata persetujuan yang belum lengkap. Jalur positif gRPC penutupan dengan fixture payroll dan verifier eksternal masih perlu dibuktikan.
