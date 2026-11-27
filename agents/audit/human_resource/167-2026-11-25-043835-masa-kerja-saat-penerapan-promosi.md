# Masa kerja saat penerapan promosi

Penerapan promosi sebelumnya masih berhasil pada pekerja berstatus aktif dengan tanggal terminasi yang sudah tercatat. Helper yang menutup penempatan sumber kini menolak pekerja dengan tanggal terminasi atau tanggal efektif yang mendahului tanggal rekrut. Aturan terminasi mengikuti validator persetujuan penempatan personel.

Regresi membuktikan penolakan sebelum penempatan lama atau riwayat lifecycle berubah. Skenario yang memenuhi syarat tetap menutup penempatan lama dan memasang penempatan baru. Tes unit Personnel dan Promotion dengan race detector serta go vet lulus.
