# Batas saldo saat persetujuan koreksi cuti

Suite gRPC mereproduksi persetujuan koreksi yang sudah diketahui melampaui saldo maksimum. Validator kini memproyeksikan saldo setelah satu koreksi dengan snapshot entitlement dan policy yang dikunci, lalu menolak hasil negatif atau di atas maksimum. Approval tidak mencadangkan saldo. Posting masih memvalidasi ulang sehingga koreksi yang disetujui tetap dapat gagal jika saldo berubah. Tes Leave dan migrasi PostgreSQL sementara dengan race detector lulus, demikian pula vet modul.
