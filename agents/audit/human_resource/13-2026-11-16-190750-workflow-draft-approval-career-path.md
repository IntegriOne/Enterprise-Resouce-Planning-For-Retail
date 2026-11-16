# Audit penutupan celah career path

## Temuan dan perbaikan

Audit sebelumnya membuktikan `CreateCareerPath` selalu gagal: database mewajibkan dua kolom approval sementara Guard menolak pengisian keduanya. Baseline `000001` kini menyediakan `status=draft`, `created_by_principal_id`, `created_at`, dan approval nullable. CHECK database memaksa pasangan status–approval serta pemisahan pembuat dan penyetuju. Create menulis audit dari actor server; input approval lama pada protobuf ditolak. `ApproveCareerPath` memakai permission `.approve` sensitif, row lock, recheck actor, SoD, validasi keluarga job/versi, transaksi dan outbox. Path approved tidak dapat diubah/dihapus melalui CRUD. Finder status dan pembuat tersedia pada repository, service, dan gRPC.

## Bukti

- Generator DBML resmi menghasilkan baseline 86 tabel dan 150 CHECK; `verify_schema.sh` lulus up, kasus integritas, down, dan up.
- Tes unit dan race seluruh modul, `go vet ./...`, `go build ./...` lulus.
- Tes gRPC `bufconn` dengan PostgreSQL `_test` lulus untuk draft, finder berhalaman, versi duplikat, penolakan spoofing approval, SoD, approval dua actor, penolakan perubahan approved, dan dua event outbox.
- Database development `human_resource` diperiksa tanpa baris bisnis, dibuat ulang sesuai kebijakan pra-rilis `000001`; container HRM kembali healthy dan migrasi versi 1 terpasang.

## Batas

Tes integrasi memakai policy tiruan, bukan grant IAM/scope parsial atau Gateway. Transisi career path masih fail-closed sampai verifikasi dokumen DRM tersedia. Approval path saat ini mengesahkan metadata path; syarat jumlah transisi minimum belum ditetapkan sebagai keputusan bisnis. Enam puluh enam resource mutasi HRM lain masih fail-closed.
