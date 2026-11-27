# Persetujuan penawaran

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy serta verifier DRM tiruan.

| Kasus | Hasil |
| --- | --- |
| Pembuat menyetujui draf sendiri | PermissionDenied |
| Mata uang pada bukti kompensasi berbeda | FailedPrecondition |
| Penyetuju independen dengan bukti cocok | Status approved serta versi kompensasi dan principal persetujuan tersimpan |
