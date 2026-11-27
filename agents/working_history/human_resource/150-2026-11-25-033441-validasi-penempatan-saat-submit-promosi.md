# Validasi penempatan asal saat pengajuan promosi

SubmitPromotionRecommendation kini memvalidasi referensi penempatan primer dan periode penempatan asal melalui aturan domain yang dipakai saat penerapan. Draf tetap dapat disimpan untuk ditinjau. Pengajuan ditolak jika referensi primer sudah berubah atau tanggal efektif tidak memungkinkan penutupan periode asal.

Tes gRPC PostgreSQL sementara membuktikan penolakan saat referensi primer dihapus dari fixture kemudian sukses setelah dipulihkan. Tes security, service, race detector dan vet lulus.
