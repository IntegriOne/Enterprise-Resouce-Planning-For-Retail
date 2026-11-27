# Subjek kasus hubungan industrial dengan riwayat

Tes gRPC mereproduksi pengosongan subjek kasus yang telah memiliki event. Pembaruan subjek atau pihak pelapor kini memeriksa riwayat di bawah lock kasus dan menolak perubahan jika event sudah tercatat. Koreksi referensi pada kasus tanpa event tetap mengikuti validasi pekerja dan Foundation yang berlaku. Tes validator, integrasi PostgreSQL sementara dengan race detector, dan vet bertag integration lulus.
