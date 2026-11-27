# Batas satu PDF per operasi

Gateway memakai anggaran kompleksitas untuk membatasi satu pembacaan dan satu penyajian isi PDF per operasi GraphQL. Tes menunjukkan dua alias sebelumnya diterima. Setelah perbaikan, alias pada Query maupun kolom base64 ditolak, dan satu PDF masih berhasil melalui fixture gRPC.
