# Audit celah workflow career path

## Temuan

`hr_career_paths.approved_by_principal_id` dan `approved_at` wajib di DBML/migrasi. `specCareerPath.Guard` menolak keduanya pada `create`, sementara `CareerValidators` belum mendaftarkan validator path dan kontrak gRPC tidak memiliki command approval. Jadi `CreateCareerPath` selalu gagal: input dengan approval ditolak Guard; input tanpa approval gagal `entity.Validate` atau constraint SQL. Keberadaan RPC bukan bukti resource dapat dibuat.

`hr_career_path_transitions` juga membutuhkan referensi dan digest dokumen persyaratan yang dimiliki DRM. Tanpa kontrak pemeriksaan versi/digest dari DRM, mutasi transisi harus tetap fail-closed. Tidak ada perubahan pada tabel DRM atau seed IAM.

## Tindak lanjut

Putuskan workflow draft→approve untuk path: kolom approval nullable selama draft plus command approve dengan SoD, atau command create-and-approve yang memvalidasi mandat, actor, rubrik dan job family dalam satu transaksi. Setelah itu tambahkan validator, permission command, tes PostgreSQL dan E2E IAM/Gateway. Jangan menghapus Guard atau mengisi approval dari klien sekadar untuk meloloskan SQL.
