# Validitas jabatan saat pengajuan workforce

SubmitWorkforcePlan memeriksa ulang status dan masa berlaku jabatan untuk semua baris. Rentang kebutuhan digabung per jabatan. Lock diambil menurut urutan ID agar rencana yang memakai lebih dari satu jabatan mengambil lock secara konsisten. Validator pembuatan dan perubahan baris memakai pemeriksaan jabatan yang sama.

Tes gRPC ke PostgreSQL membuktikan jabatan yang dinonaktifkan setelah draf dibuat menghalangi pengajuan. Pengajuan berhasil setelah jabatan aktif kembali. Jumlah outbox memastikan penolakan tidak mencatat event pengajuan. Tes modul dengan race detector dan go vet lulus.
