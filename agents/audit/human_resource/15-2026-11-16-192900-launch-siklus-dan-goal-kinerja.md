# Audit launch siklus dan goal kinerja

`LaunchPerformanceCycle` mengubah draft menjadi open dalam transaksi yang mengunci siklus, menulis actor/waktu dari server dan outbox. Permission `human_resource.performance_cycles.launch` dipisah dari update serta dinyatakan sensitif. CRUD umum tidak dapat mengisi field peluncuran atau mengubah status. Validator `hr_performance_goals` mengunci worker aktif, siklus terbuka dan goal induk dalam urutan tersebut; kode goal unik per worker/siklus, tanggal due dibatasi periode, dan perubahan goal yang sudah dipakai child/pengukuran/skor ditolak. `actual_value` tidak dapat diisi melalui CRUD draft.

Tes integrasi gRPC `bufconn` → service → repository → PostgreSQL `_test` lulus untuk create/finder/update/launch siklus, create/get/finder/update/delete goal, penolakan sebelum launch, duplikasi, spoofed actual value, serta outbox tiga event per agregat untuk mutasi berhasil. `go generate`, fmt, vet, race test, build dan Docker health lulus.

Batas: policy integrasi adalah tiruan global. Belum ada grant IAM resmi, scope parsial, HMAC/mTLS runtime pada jalur tes, Gateway E2E, maupun workflow self/manager review, kalibrasi dan finalisasi siklus. Validator produksi kini mencakup 14 dari 78 resource berkode mutasi; 64 lainnya tetap fail-closed.
