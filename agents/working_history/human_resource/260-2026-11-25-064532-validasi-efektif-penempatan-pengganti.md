# Validasi tanggal efektif penempatan pengganti

Penetapan penempatan primer menolak penempatan yang sudah disetujui jika awal periodenya mendahului tanggal mulai bekerja. Transfer dan promosi memakai aturan domain bersama yang menolak tanggal efektif pengganti sebelum tanggal mulai bekerja. Persamaan tanggal tetap diizinkan oleh aturan tanggal tersebut.

Tes domain Personnel, validator, service Personnel dan Promotion lulus dengan race detector. `go vet ./...` lulus. [Handoff Personnel](../../grpc/hrm/17-operasi-personnel.md)
