# Jadwal pengganti setelah pembatalan penempatan

Indeks unik PostgreSQL sekarang melewati penempatan batal agar jadwal pengganti yang sama dapat disimpan tanpa menghapus riwayat. Uji SQL terfokus membuktikan constraint persetujuan, penolakan periode primer bertumpuk, dan penerimaan jadwal pengganti. Tes validator kapasitas dengan race detector dan go vet lulus. Verifikasi skema penuh berhenti pada fixture payroll lama yang mencoba persetujuan tanpa review.
