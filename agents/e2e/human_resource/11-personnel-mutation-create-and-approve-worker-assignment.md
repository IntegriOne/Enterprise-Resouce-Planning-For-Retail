# Pemeriksaan masa kerja manajer melalui gRPC Personnel

Tes memakai PostgreSQL sementara, Redis TLS terisolasi, dan race detector. Tanggal masuk manajer diatur langsung di database tes.

| Kasus | Hasil |
| --- | --- |
| Buat penempatan yang dimulai sebelum tanggal masuk manajer | FailedPrecondition |
| Buat permintaan setelah tanggal masuk manajer diselaraskan | Draf tersimpan tanpa persetujuan |
| Setujui draf setelah tanggal masuk manajer dikoreksi menjadi lebih akhir | FailedPrecondition |
| Setujui setelah periode kembali sesuai | Persetujuan tersimpan dengan ID dan waktu penyetuju |
| Ulangi persetujuan yang sudah berhasil | FailedPrecondition dan total outbox tetap dua event |
