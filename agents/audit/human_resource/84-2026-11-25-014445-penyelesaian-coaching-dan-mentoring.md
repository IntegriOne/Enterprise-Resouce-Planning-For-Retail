# Penyelesaian coaching dan mentoring

CompleteDevelopmentAction mencatat selesainya coaching atau mentoring lokal pada rencana aktif. Actor manusia harus memiliki izin step-up, pernah menyetujui rencana, dan tetap terhubung ke principal IAM mentor aktif yang tercatat. Command mengunci tugas, rencana dan mentor. Waktu server serta event outbox tersimpan atomik. Tugas dengan kursus, target tingkat kompetensi atau bukti dokumen memerlukan workflow verifikasi tersendiri.

Tes gRPC ke PostgreSQL lulus untuk penolakan rencana draf, penyelesaian oleh mentor, penolakan pengulangan dan perlindungan penghapusan. IAM menggunakan policy tiruan. Penyelesaian tindakan ini tidak mengubah hasil kursus, kompetensi, atau jabatan.
