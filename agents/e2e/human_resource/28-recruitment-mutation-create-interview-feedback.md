# Pengiriman feedback wawancara

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy tiruan. Feedback untuk penutupan ditulis melalui RPC.

| Kasus | Hasil |
| --- | --- |
| Kirim sebelum wawancara dimulai | FailedPrecondition |
| Principal lain mengatasnamakan panelis | PermissionDenied |
| Panelis mengirim feedback miliknya | Berhasil |
| Principal lain mengubah atau menghapus feedback | PermissionDenied |
| Recruiter menutup wawancara dengan feedback RPC | Status completed |
| Panelis mengubah feedback setelah selesai | FailedPrecondition |
