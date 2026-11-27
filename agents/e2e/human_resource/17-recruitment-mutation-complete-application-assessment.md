# Penyelesaian asesmen lamaran

Tes berjalan melalui gRPC dengan PostgreSQL sementara, policy dan verifier bukti tiruan, serta race detector.

| Kasus | Hasil |
| --- | --- |
| Pengguna lain menyelesaikan asesmen yang sudah memiliki penilai | PermissionDenied |
| Penilai menyelesaikan sebelum jadwal yang tercatat | FailedPrecondition |
| Penilai menyelesaikan saat waktu buka requisition masih di masa depan | FailedPrecondition |
| Penilai menyelesaikan setelah jadwal draf dikosongkan melalui RPC update | Status completed |
| Penyelesaian diulang dengan skor berbeda | FailedPrecondition |
