# Konsistensi versi dokumen pekerja

Tes gRPC mereproduksi patch versi dokumen yang menyimpan digest lama. Mapper kini mewajibkan versi berbeda saat referensi dokumen berubah dan mewajibkan digest eksplisit saat referensi versi berubah. Tes PostgreSQL sementara membuktikan penolakan patch versi tanpa digest dan keberhasilan patch versi dengan digest. Verifikasi keaslian digest tetap menjadi tanggung jawab DRM. Tes Personnel dengan race detector dan vet modul lulus.
