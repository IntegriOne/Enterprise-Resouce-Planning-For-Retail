# Leader backend — HRM — Perbaikan korelasi lintas management

- **ID** — 6
- **Owner** — Agent HRM sebagai pelaksana, leader backend sebagai peninjau. Perubahan di service lain dikoordinasikan dengan pemilik pada matriks.
- **Priority** — Kerjakan temuan P1 lebih dahulu. P2 menjaga kebenaran jalur pada konfigurasi produksi.
- **Status** — Siap dikerjakan.
- **Baseline** — Working tree lokal ERP Retail pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`. Baca ulang source yang dirujuk karena beberapa service masih merupakan perubahan lokal.

## Task

Rute keluar terminasi, posting payroll dan pelepasan anggaran tersedia. Penerimaan publikasi payroll dari Finance sudah mencapai handler HR, tetapi penyimpanan dan pembacaan PDF DRM tidak dirangkai di entrypoint yang diperiksa. Pengiriman pengajuan akun ke IAM masih dipicu Gateway. Selesaikan tiap temuan melalui entrypoint aplikasi atau workload asli hingga hasilnya dapat dibaca kembali dari pemilik fakta. [Audit HRM](../../audit/human_resource/198-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Peta korelasi seluruh management](../../audit/gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md)

## Request

### HR-01 — P1 · Pengajuan akses tersimpan tanpa jaminan sampai di antrean IAM

- file: [accessrequest/service.go L42](../../../backend/services/human_resource/internal/service/accessrequest/service.go#L42) · [outbox/store.go L53](../../../backend/services/human_resource/infrastructure/outbox/store.go#L53) · [cmd/wiring.go L145](../../../backend/services/human_resource/cmd/wiring.go#L145) · [human_resource/access_request.go L69](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L69). Path menunjukkan titik baca/perubahan; ubah kode milik HRM dan sepakati kontrak pasangan dengan owner terkait.
- function: Tambahkan rute pengiriman pengajuan terverifikasi ke IAM-02 di transaksi `CreateAccessRequest`, pemilih event `Store.SelectDue`, publisher dan wiring runtime. Gunakan outbox domain yang sudah tersedia bila kontraknya mencukupi.
- logic: Simpan immutable request ID, worker/entity, lifecycle evidence, pemohon, catalog, waktu dan correlation. Penerimaan IAM memberi receipt persisten. Pada batch recovery baca record lama yang pending, cocokkan ke IAM sebelum kirim ulang. Integrasikan dengan GW-02.
- error_handling: Kesalahan sementara menyisakan event pending dengan backoff. Konflik payload tetap terlihat untuk rekonsiliasi. Cegah jalur log/webhook umum menghabiskan event privat ketika IAM mati.
- security: Jangan menyimpan sesi browser atau credential pada outbox. Worker hanya dapat meneruskan bukti permohonan. Grant tetap milik IAM.
- expected_result: Restart di antara commit HR dan delivery IAM berakhir pada satu receipt. Request lama yang sah dapat dipulihkan tanpa permohonan manual baru.
- testing: Uji gagal sebelum/ sesudah commit, IAM mati selama batch, replay dua worker, konflik dengan katalog/worker berbeda, request historis dan retry saat jalur langsung juga aktif.

### HR-02 — P1 · Hasil publikasi gaji belum mencapai dokumen yang dapat dicetak

- file: [payroll/service.go L558](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L558) · [inbox/payroll_publication.go](../../../backend/services/human_resource/infrastructure/inbox/payroll_publication.go) · [document_records/payslip.go L51](../../../backend/services/human_resource/infrastructure/integration/document_records/payslip.go#L51) · [service/registry.go L180](../../../backend/services/human_resource/internal/service/registry.go#L180) · [payroll/payslip_reissue.go L27](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L27) · [payroll/payslip.go L57](../../../backend/services/human_resource/internal/service/payroll/payslip.go#L57) · [cmd/main.go L94](../../../backend/services/human_resource/cmd/main.go#L94). Path menunjukkan titik baca/perubahan; ubah kode milik HRM dan sepakati kontrak pasangan dengan owner terkait.
- function: Koordinasikan verifier publikasi Finance dan storage/read versi privat DRM, kemudian pasang ketiga opsi pada bootstrap. Sediakan rekonsiliasi publikasi lama yang tertunda saat DRM tersedia.
- logic: Verifikasi publication ID, run, result dan worker. Buat satu PDF deterministik per versi hasil melalui kunci idempoten. Pertahankan versi dokumen yang sudah dibuat bila proses berhenti sebelum pencatatan HR. Pegawai aktif dapat membaca slip sendiri; mantan pegawai melalui HR yang memiliki scope dan audit.
- error_handling: DRM belum siap tetap menjadi dependency tertunda. Pastikan kegagalan satu dokumen dapat dilanjutkan, dead letter dapat direplay dan petugas dapat memulihkan isi hilang. Jangan mengubah status pelunasan atau membuat publikasi Finance baru ketika retry.
- security: Verifikasi hak akses pada setiap baca. Batasi ukuran/PDF, periksa versi dan digest. Jangan simpan isi slip dalam log, URL publik atau payload notifikasi.
- expected_result: Setelah Finance menerbitkan payroll, slip tersedia melalui jalur yang diizinkan. Replay publikasi tidak menggandakan versi; reissue yang sah mengembalikan PDF yang tepat.
- testing: Uji Finance published saat DRM mati lalu pulih, batch parsial, dua callback paralel, versi hilang, self pegawai aktif, mantan pegawai sendiri ditolak, HR scope benar/salah. Tunda klaim live sampai owner DRM tersedia.

### HR-03 — P2 · Validasi organisasi di EF belum memakai profil transport terverifikasi

- file: [enterprise_foundation/client.go L82](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L82) · [cmd/wiring.go L78](../../../backend/services/human_resource/cmd/wiring.go#L78) · [grpc/procurement_workload.go](../../../backend/services/enterprise_foundation/transport/grpc/procurement_workload.go). Path menunjukkan titik baca/perubahan; ubah kode milik HRM dan sepakati kontrak pasangan dengan owner terkait.
- function: Perluas `Dial` dan config EF saja, pertahankan `VerifyWorkerReferences` serta `VerifyPositionReferences`.
- logic: Periksa SAN dan owner facts termasuk relasi unit–entity, lokasi dan waktu berlaku yang sekarang dipakai HR. Lakukan rollout bersama EF-01.
- error_handling: Sertifikat salah atau EF unavailable menghentikan mutasi yang wajib verifikasi; pertahankan deadline dan klasifikasi error.
- security: Kunci workload HR dan sertifikat Gateway dikelola terpisah. Jaga signature caller existing.
- expected_result: CreateWorker dengan entitas sah tetap berfungsi; invalid/expired scope dan kanal tak tepercaya ditolak.
- testing: Uji createWorker dan posisi via HTTPS, unit beda entity, hire date sebelum incorporation date, sertifikat salah dan operasi lokal saat EF gagal.

## Action

1. Cocokkan baseline, kode, proto/SDL, status working tree dan kontrak lawan. Sepakati ID operasi/event, owner, scope, waktu, versi, ACK dan bentuk respons gagal sebelum menyalakan jalur baru. Pertahankan kontrak pemanggil lama selama rollout.
2. Kerjakan blok Request sesuai dependency. Tambahkan fixture kecil untuk kasus yang belum terliput. Unit test server gRPC hanya di layer service. Buktikan constraint/efek persisten melalui entrypoint normal dengan database Docker disposable. Untuk Gateway uji resolver/adapter serta HTTPS. Dilarang menulis grant IAM langsung ke DB.
3. Dari `backend/services/human_resource` jalankan `go vet ./...` dan `go test -race -count=1 ./...` pada modul yang berubah. Bila kontrak berubah jalankan generator resmi module; untuk Gateway gunakan `GOWORK=off go generate ./...` di `backend/gateway` dan periksa diff. Uji Caddy di `https://erp.localhost:8443/api/graphql` melalui dua replika. Login superadmin, pilih permission katalog, buat role, pasang permission dan assignment scoped via GraphQL IAM. Gunakan aktor berbeda bila SoD disyaratkan. Cabut grant sementara setelah tes.
4. Untuk jalur workload, cek listener, identitas peer, source event/request, aggregate/effect, versi, correlation ID dan status owner sebelum/sesudah retry. Bedakan receipt pending dari efek diterapkan; lampirkan bukti hasil aktual, termasuk kondisi gagal/tertunda. Hindari PII, voucher, isi slip atau secret pada log.
5. Laporkan per ID file yang diubah, kontrak pasangan, kasus lulus/gagal, rollback/replay dan SHA. Tandai DONE hanya setelah seluruh kriteria aktif terbukti, sesuai aturan repo. Tim yang menunggu dependency dapat meneruskan temuan independen.

## Context

- [agents/arsitektur.md](../../arsitektur.md) · [agents/boundary.md](../../boundary.md) · [agents/users.md](../../users.md) · [agents/how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [agents/how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md) · [agents/instruction_template.md](../../instruction_template.md) · [AGENTS.md](../../../AGENTS.md) · [docs/README.md](../../../docs/README.md)
- [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md) · [Konteks HRM](../../context/human_resource/1-context.md)
- [Instruksi Gateway](../gateway/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi IAM](../iam/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Enterprise Foundation](../enterprise_foundation/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Finance & Accounting](../finance_accounting/5-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Procurement & Sourcing](../procurement_sourcing/2-leader-perbaikan-korelasi-gateway-grpc-instructions.md)
