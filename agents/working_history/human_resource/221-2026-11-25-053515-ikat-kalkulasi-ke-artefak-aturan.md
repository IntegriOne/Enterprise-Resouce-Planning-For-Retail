# Ikat kalkulasi ke artefak aturan payroll

HRM mengirim ID versi dokumen serta digest artefak dari rule set ke kalkulator. Hasil dengan digest artefak berbeda ditolak sebelum hasil payroll tersimpan. Uji gRPC dengan PostgreSQL sementara dan race detector lulus. Adapter produksi tetap perlu membaca dan memverifikasi isi artefak dari pemilik dokumen.
