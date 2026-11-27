# Aturan domain terminasi pekerja

Personnel memiliki fungsi domain yang membentuk perubahan status pekerja dari rencana offboarding yang berjalan dan disetujui secara independen. Fungsi memeriksa identitas pekerja, tanggal efektif, masa kerja, dan urutan waktu persetujuan. Hasil mengambil tanggal serta alasan dari rencana dan mengosongkan pointer penempatan primer tanpa mengubah input.

Tes domain dengan race detector dan `go vet ./...` lulus. Penyimpanan atomik, penutupan kontrak dan penempatan, serta RPC masih terbuka.
