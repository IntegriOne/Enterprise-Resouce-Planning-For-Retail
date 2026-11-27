# Bukti dokumen pengakuan kebijakan

Tes gRPC mereproduksi penyimpanan UUID bukti tanpa verifikasi DRM ketika pekerja mengakui kebijakan. Validator menolak referensi bukti baru sampai kontrak verifier tersedia. Pengakuan tanpa lampiran tetap tercatat untuk pekerja dengan principal yang sesuai. Percobaan berlampiran yang ditolak tidak menghabiskan kesempatan pengakuan, dan percobaan pengulangan setelah sukses masih ditolak. Referensi historis tidak dipindahkan atau ditulis ulang. Integrasi Policy dengan PostgreSQL sementara, race detector, dan vet modul lulus.
