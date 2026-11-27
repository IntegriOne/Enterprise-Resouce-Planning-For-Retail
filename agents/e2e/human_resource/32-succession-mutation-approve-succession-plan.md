# Persetujuan rencana suksesi

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy dan verifier panel tiruan.

| Kasus | Hasil |
| --- | --- |
| Rencana tanpa nominasi | FailedPrecondition |
| Verifier menolak mandat | PermissionDenied |
| Verifier menyetujui mandat dan kandidat sah | Principal dan waktu approval tersimpan |
| Penghapusan kandidat setelah approval | FailedPrecondition |
| Mengulang approval | FailedPrecondition |

Tes integrasi juga memeriksa bahwa verifier menerima ID nominasi, ID pekerja, kode kesiapan, dan tanggal asesmen dari nominasi yang disahkan.
