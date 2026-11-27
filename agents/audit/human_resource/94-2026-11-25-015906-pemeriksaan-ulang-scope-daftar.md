# Pemeriksaan ulang scope daftar

Resource List kini membaca ulang scope setelah menghitung total dan menolak respons jika filter berubah. Langkah ini menjaga metadata pagination ketika halaman kosong sehingga pemeriksaan izin per item tidak berjalan. Pemeriksaan revisi actor tetap dilakukan sebelum respons.

Tes regresi mereproduksi penerimaan respons dengan scope yang berubah pada policy tiruan. Setelah perbaikan tes ditolak dengan Forbidden. Tes modul dengan race detector, go vet, dan integrasi Workforce Planning gRPC ke PostgreSQL lulus.
