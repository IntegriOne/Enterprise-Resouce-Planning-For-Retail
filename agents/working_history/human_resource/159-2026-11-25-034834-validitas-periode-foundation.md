# Validitas periode posisi pada adapter Foundation

Verifikasi posisi menolak periode terbalik atau tanggal akhir nol. Timestamp pendirian dan pembubaran badan hukum yang tidak valid juga ditolak sebelum konversi. Tes transport Dial dengan server fixture menolak periode terbalik dan timestamp pendirian rusak. Tes adapter dan cmd dengan race detector serta go vet lulus.
