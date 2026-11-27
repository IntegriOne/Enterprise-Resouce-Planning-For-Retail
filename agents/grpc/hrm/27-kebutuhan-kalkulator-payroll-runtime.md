# List operasi

## Penjelasan CalculatePayrollRun

Runtime HRM menyediakan kontrak `Calculator` dan titik injeksi `WithPayrollCalculator`. Paket domain merekonsiliasi nominal dengan baris komponen. Bootstrap saat ini belum memasang implementasi mesin kalkulasi. Gateway sudah meneruskan command dan status `UNAVAILABLE` saat dependency belum tersedia.

## Butuh GraphQL Mutation

`calculatePayrollRun(payroll_run_id: ID!)` sudah tersedia. HRM memvalidasi run draf, inisiator, periode terbuka, grup aktif, dan rule set yang disetujui. Respons sukses mengembalikan status calculated dan waktu kalkulasi.

## RBAC

Gunakan sesi employee dengan izin global `human_resource.payroll_runs.calculate`, MFA terkini, dan CSRF. HRM memeriksa bahwa pemanggil merupakan penginisiasi run.

## Kontrak runtime yang masih dibutuhkan

- Tetapkan sumber snapshot input immutable dan formatnya, termasuk pemilik data, pengambilan berdasarkan ID, dan verifikasi SHA-256 atas isi aktual.
- DRM menyediakan pembacaan privat versi immutable artefak aturan. Kalkulator memverifikasi SHA-256 terhadap digest yang melekat pada versi yang disetujui sebelum menjalankan aturan.
- Kalkulator menggunakan versi engine dan aturan yang diminta, periode, badan hukum, mata uang, presisi pembulatan, jenis run, dan arah settlement. Hasil mengikat ID run, aturan dan snapshot ke digest input, aturan, dan output serta hasil dan baris komponen.
- Pasang adapter melalui `WithPayrollCalculator` dan uji snapshot atau aturan tidak cocok serta kegagalan pembacaan. HRM menyimpan hasil dan outbox secara atomik setelah hasil lolos validasi.

Regresi `TestPayrollRPCAndApprovalWithPostgreSQL` dengan tag integration, race detector, dan database uji baru lulus. Tes memakai kalkulator fixture. Jalur produksi masih memerlukan sumber snapshot dan artefak serta engine yang disepakati. [Kontrak kalkulasi](../../../backend/services/human_resource/internal/service/payroll/calculation.go) · [Integrasi DRM](24-kebutuhan-integrasi-drm.md).

## Kriteria pemasangan adapter

- Startup server memanggil `BuildDependencies` tanpa opsi kalkulator. Adapter produksi harus dipasang secara eksplisit lewat `WithPayrollCalculator` setelah penyedia snapshot dan artefak tersedia. [Startup HRM](../../../backend/services/human_resource/cmd/main.go) · [Registry](../../../backend/services/human_resource/internal/service/registry.go).
- Adapter menerima `CalculationRequest` dan menghitung dari snapshot serta artefak yang digest-nya cocok dengan byte yang benar-benar dipakai. Adapter mengembalikan `CalculationResult` dengan ID run, rule set dan snapshot yang sama, digest input dan artefak yang sesuai, digest output 32 byte, serta hasil yang dapat direkonsiliasi beserta komponennya. Hasil harus berstatus `calculated` dan bebas referensi pembayaran serta slip. [Kontrak Go](../../../backend/services/human_resource/internal/service/payroll/calculation.go).
- Uji proses utuh memakai run yang dibuat oleh operator uji. Verifikasi respons `calculated` melalui HTTPS, nilai hasil dan baris komponen, presisi mata uang, serta penyimpanan hasil dan outbox. Ulangi dengan snapshot atau artefak berbeda dan pastikan run tetap draft tanpa hasil kalkulasi. Jalur positif di gRPC dengan kalkulator fixture kembali lulus dengan race detector. Runtime standar melalui Gateway sudah terbukti mengembalikan `UNAVAILABLE` dan mempertahankan draf. [Uji HTTPS](../../e2e/gateway/107-hrm-query-mutation-draf-payroll-mfa.md) · [Uji gRPC](../../e2e/human_resource/54-query-mutation-regresi-grpc-postgresql.md).

## Penyelarasan snapshot dengan pembuatan run

Gateway mewajibkan `input_snapshot_reference_id` dan `input_digest` saat `createPayrollRun`. Jalur kalkulasi membaca keduanya dari run yang sudah disimpan. Handoff implementasi perlu menetapkan siapa yang membekukan input, bagaimana operator memilih snapshot yang cocok untuk periode dan grup, serta cara kalkulator mengambil bytes yang sama dari penyedia snapshot. Pemilihan hanya dapat menampilkan ringkasan yang diizinkan, seperti periode, grup, waktu pembekuan, dan jumlah pekerja. ID serta digest dipakai sebagai referensi teknis.

PostgreSQL sudah mengikat run, periode, dan grup lewat foreign key komposit untuk badan hukum, yurisdiksi, dan mata uang. Percobaan mengubah mata uang periode secara terpisah ditolak oleh constraint. Suite gRPC payroll lulus ulang dengan race detector dan PostgreSQL disposable, termasuk penolakan digest input dan artefak yang berbeda. Verifikasi tersebut memakai kalkulator fixture. [Schema GraphQL](../../../backend/gateway/graphql/human_resource/hrm_payroll_posting.graphql) · [Migrasi HRM](../../../backend/services/human_resource/db/migrations/000001_hrm.up.sql).

Kalkulasi menolak hasil yang sudah membawa referensi penerbitan slip, dokumen slip, atau pembayaran. Regresi gRPC menyuntikkan referensi penerbitan dari kalkulator fixture dan memastikan HRM menolaknya sebelum kalkulasi sah yang berikutnya.

`NewVerifiedCalculator` menyediakan adapter untuk `WithPayrollCalculator`. Adapter membaca snapshot berdasarkan ID dan artefak aturan berdasarkan ID versi, lalu memverifikasi SHA-256 atas salinan byte sebelum memanggil mesin. Uji menolak perubahan pada snapshot atau artefak sebelum engine berjalan. Penyedia snapshot, adapter pembacaan DRM, dan engine aturan produksi masih perlu tersedia untuk wiring startup. [Adapter verifikasi](../../../backend/services/human_resource/internal/service/payroll/verified_calculator.go)

Suite gRPC PostgreSQL juga memasang `NewVerifiedCalculator` dengan penyedia byte dan engine fixture. Snapshot atau aturan yang berubah ditolak sebelum engine dipanggil. Snapshot asli dan snapshot pengganti masing-masing lulus dengan digest byte aktual. Tes mempertahankan pemeriksaan atas korelasi digest pada event pengajuan posting.

Sebelum membaca snapshot atau menjalankan mesin, `CalculatePayrollRun` memeriksa ulang integritas draf terhadap referensi aktif, termasuk rentang badan hukum EF. Regresi gRPC dengan PostgreSQL dan verifier EF fixture membuktikan bahwa perubahan keputusan EF setelah draf dibuat menghentikan kalkulasi sebelum mesin dipanggil.

Pembaca snapshot harus mengembalikan ID snapshot, periode payroll, grup payroll, dan badan hukum dari metadata yang tersimpan bersama byte. Kalkulator memeriksa keempat referensi sebelum memeriksa digest dan membaca artefak aturan. Uji membuktikan byte dan digest yang sah dari grup lain tetap ditolak sebelum engine berjalan. Penyimpanan dan penerbitan snapshot oleh HRM masih perlu diimplementasikan.

HRM menyediakan tabel `hr_payroll_input_snapshots` dan repository insert/read. PostgreSQL mengikat rekaman ke periode, grup, badan hukum, yurisdiksi, dan mata uang yang sama. Constraint mencocokkan digest dengan byte aktual, dan trigger menolak update serta delete. Tes disposable mencakup pemasangan dan upgrade baseline, digest salah, grup salah, baca kembali, dan percobaan mutasi. Command untuk menerbitkan snapshot dari input payroll yang telah disetujui masih diperlukan.

Repository `Capture` membekukan profil payroll dalam cakupan periode terbuka, kompensasi aktif atau historis yang masih relevan, dan penyesuaian yang sudah disetujui. Satu pernyataan PostgreSQL membaca dan menyimpan snapshot, menghitung SHA-256, serta mengembalikan metadata. Token rekening tidak masuk ke isi snapshot. Uji PostgreSQL mencakup periode tanpa profil yang memenuhi syarat dan pembacaan profil yang berhasil dibekukan. Pemasangan command terotorisasi dan pemilihan snapshot lewat Gateway masih terbuka.
