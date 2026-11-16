# Kebutuhan pembuktian target scope untuk IAM

IAM membutuhkan kontrak gRPC privat dari pemilik objek sebelum memberikan role berscope `legal_entity`, `organization_unit`, `facility`, `supplier_party`, atau `vendor_party`. Saat ini IAM hanya bisa memeriksa format UUID; penugasan scope eksternal dibuat gagal tertutup (`iam.scope.owner_verification_unavailable`).

Operasi minimum yang dibutuhkan: verifikasi `scope_type` + `scope_id` terhadap objek otoritatif, jenis objek, status aktif, dan relasi induk yang relevan. Respons harus memuat ID kanonik, status, versi atau waktu pembacaan, dan hubungan legal entity/unit/party yang diperlukan untuk evaluasi scope. IAM memerlukan deadline, identitas workload, error `not found`/`inactive`/`unavailable` yang terpisah, serta semantik cache/freshness dan revoke ketika objek berubah. `facility` perlu dipastikan pemilik resminya bersama WMS; IAM tidak boleh membaca database owner secara langsung.

Setelah kontrak tersedia, IAM dapat memasang adapter verifier pada service role/group dan menguji grant sah, UUID salah jenis, objek nonaktif, serta kegagalan owner sebelum commit. Tidak ada grant yang dibuat dari catatan ini.
