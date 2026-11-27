# Wiring verifier rekrutmen

Composition root HRM meneruskan adapter rekrutmen melalui opsi WithRecruitmentVerifiers. Tes gRPC–PostgreSQL kini membuat service lewat registry yang sama dengan startup. Verifier tiruan membuktikan alur lengkap, sementara registry tanpa verifier menolak persetujuan penawaran dengan Unavailable. Go test dan go vet lulus.
