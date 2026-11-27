# Verifikasi kontak kandidat

CreateCandidate memeriksa bahwa kontak aktif, belum dihapus, dan terikat pada individu yang diverifikasi melalui Enterprise Foundation. Perubahan lewat CRUD mempertahankan kedua referensi. Tes adapter, wiring, validator dan gRPC dengan PostgreSQL serta fixture EF lulus dengan race detector. Go vet lulus pada paket yang diubah.
