# Pembatalan wawancara

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Recruiter lain membatalkan | PermissionDenied |
| Recruiter terkait membatalkan wawancara planned | Status cancelled |
| Membatalkan wawancara yang sudah dibatalkan | FailedPrecondition |
| Membuat wawancara lain di slot yang sama | Berhasil |
| Memeriksa riwayat panel dan outbox | Panelis tetap tercatat dan satu event pembatalan tersimpan |
