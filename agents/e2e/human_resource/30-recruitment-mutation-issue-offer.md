# Penerbitan penawaran

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy dan verifier DRM tiruan.

| Kasus | Hasil |
| --- | --- |
| Recruiter lain menerbitkan penawaran | PermissionDenied |
| Digest DRM tidak cocok | FailedPrecondition |
| Recruiter terkait memakai bukti cocok | Penawaran issued, versi dan digest tersimpan, lamaran berpindah ke offer |
