# Leader backend — Finance & Accounting — Perbaikan korelasi lintas management

- **ID** — 5
- **Owner** — Agent Finance & Accounting sebagai pelaksana, leader backend sebagai peninjau. Perubahan di service lain dikoordinasikan dengan pemilik pada matriks.
- **Priority** — Kerjakan temuan P1 lebih dahulu. P2 menjaga kebenaran jalur pada konfigurasi produksi.
- **Status** — Siap dikerjakan.
- **Baseline** — Working tree lokal ERP Retail pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`. Baca ulang source yang dirujuk karena beberapa service masih merupakan perubahan lokal.

## Task

Posting payroll HR dan notifikasi alokasi pembayaran ke Procurement memiliki jalur implementasi. Dua risiko operasional terpisah adalah bukti valuasi WMS yang belum dikirim saat probe sukses dan event untuk publisher umum yang tidak terpilih saat relay Procurement aktif. Selesaikan tiap temuan melalui entrypoint aplikasi atau workload asli hingga hasilnya dapat dibaca kembali dari pemilik fakta. [Audit Finance & Accounting](../../audit/finance_accounting/8-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Peta korelasi seluruh management](../../audit/gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md)

## Request

### FA-01 — P1 · Valuasi WMS tanpa pengiriman efek dan replay lintas item kurang ketat

- file: [finance/client.go L82](../../../backend/services/inventory_warehouse/internal/integration/finance/client.go#L82) · [integration/service.go L210](../../../backend/services/inventory_warehouse/internal/service/integration/service.go#L210) · [inventory_cost/service.go L63](../../../backend/services/finance_accounting/internal/service/inventory_cost/service.go#L63) · [handler/inventory_cost_handler.go L20](../../../backend/services/finance_accounting/transport/grpc/handler/inventory_cost_handler.go#L20) · [security/policy.go L358](../../../backend/services/finance_accounting/internal/security/policy.go#L358). Path menunjukkan titik baca/perubahan; ubah kode milik Finance & Accounting dan sepakati kontrak pasangan dengan owner terkait.
- function: Perbaiki `ApplyInventoryCostEvent` dan kontrak `DeliverOutboxEvent` bersama owner WMS. Tetapkan sumber nilai/jurnal di Finance dan bukti movement/effect di WMS.
- logic: Ikat key bisnis pada item, effect/original, quantity, tanggal, dimensi relevan dan valuasi yang dikonfirmasi Finance. Normalisasi decimal. Pisahkan not-found dari kesalahan pencarian. Jika fakta dan nilai belum lengkap, simpan receipt pending atau tolak sementara. Verifikasi ack ID dan hasil melalui lookup owner. Proses backlog lama lewat rekonsiliasi ledger secara idempoten.
- error_handling: Replay data sama mengembalikan cost entry yang sama. Key sama dengan item, jumlah atau fakta berbeda menghasilkan konflik tanpa perubahan saldo. Timeout setelah commit aman diulang.
- security: Hanya workload beridentitas WMS di allowlist method terkait boleh memasok fakta fisik. Hanya Finance menghitung/menyetujui nilai; jangan menerima delta bebas dari browser.
- expected_result: Setiap efek yang memenuhi kriteria valuasi menghasilkan tepat satu entry dan reversal tertaut. Transfer custody tanpa perubahan ownership tidak membuat valuasi tambahan. Probe transport tidak mengubah status delivery.
- testing: Uji receipt lalu ownership, transfer internal, reversal, dua item dengan key sama, error lookup, restart sesudah posting. Bandingkan source effect dan cost entry serta saldo sebelum/sesudah.

### FA-02 — P1 · Filter relay privat dapat menyisihkan tujuan webhook yang aktif

- file: [cmd/wiring.go L297](../../../backend/services/finance_accounting/cmd/wiring.go#L297) · [integration/integration_repository.go L211](../../../backend/services/finance_accounting/repository/integration/integration_repository.go#L211) · [procurement/client.go L69](../../../backend/services/finance_accounting/internal/integration/procurement/client.go#L69). Path menunjukkan titik baca/perubahan; ubah kode milik Finance & Accounting dan sepakati kontrak pasangan dengan owner terkait.
- function: Perbaiki pemilihan repo di `buildOutboxRelay`, `procurementInvoiceRoute.Publish` dan filter `SelectPendingBatchForUpdate`.
- logic: Definisikan kepemilikan jenis event dan tujuan. Dalam mode privat saja klaim hanya rute aktif. Dalam mode gabungan klaim pula event tujuan umum yang sah. Jika satu event diwajibkan ke lebih dari satu tujuan, simpan progres dan retry per tujuan; pisahkan data payroll yang wajib privat. Pertahankan ordering/dependency invoice sebelum alokasi sesuai consumer.
- error_handling: Tujuan wajib yang mati tetap pending atau terminal dengan alasan yang dapat direkonsiliasi. Sukses log diagnostik tidak menjadi bukti consumer produksi.
- security: Gunakan allowlist tipe dan payload minimum; event HR dan AP khusus tidak jatuh ke webhook umum tanpa kontrak yang disetujui.
- expected_result: Saat webhook dan relay invoice aktif bersamaan, satu event umum sampai ke webhook, satu event AP masuk inbox PS, dan event HR tetap ke HR bila rutenya aktif.
- testing: Uji kombinasi standby, webhook saja, PS saja, PS+webhook, HR+PS+webhook, tujuan wajib mati dan restart. Periksa jumlah pending, ack per tujuan dan hasil receipt.

### FA-03 — P2 · Kredensial client EF tidak diteruskan dan binding respons longgar

- file: [cmd/wiring.go L119](../../../backend/services/finance_accounting/cmd/wiring.go#L119) · [ef/client.go L48](../../../backend/services/finance_accounting/internal/integration/ef/client.go#L48) · [grpc/procurement_workload.go](../../../backend/services/enterprise_foundation/transport/grpc/procurement_workload.go). Path menunjukkan titik baca/perubahan; ubah kode milik Finance & Accounting dan sepakati kontrak pasangan dengan owner terkait.
- function: Teruskan konfigurasi TLS ke `ef.NewClient` dan cocokkan response ID pada keempat lookup. Validasi owner entity bila use case menyediakannya.
- logic: Baca status aktif/deleted dan hubungan unit/location dari EF; gunakan scope API EF sesuai kebutuhan. Pertahankan kecocokan ID pada mapping domain. Koordinasi listener EF-01.
- error_handling: Tolak ID/kind/owner keliru sebagai bukti invalid dan bedakan dari upstream unavailable; jangan mengembalikan success memakai input UUID.
- security: Pertahankan mTLS peer Finance dan propagasi aktor terverifikasi, tanpa alternatif plaintext otomatis pada profil produksi.
- expected_result: Mutasi yang membutuhkan fakta organisasi memakai fakta objek yang diminta dan gagal tertutup bila owner tidak dapat dibuktikan.
- testing: Uji dua legal entity dan dua party aktif, reply silang, owner unit salah, sertifikat/kunci parsial dan operasi AP dengan EF tersedia.

### FA-04 — P1 · Terbitkan payroll dari alur Finance dan simpan penerbit yang sebenarnya

- file: [Rancangan operasi Finance](../../grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L163), [SDL payroll](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql), [handler](../../../backend/services/finance_accounting/transport/grpc/handler/payroll_handler.go#L112), [service payroll](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L559), [policy](../../../backend/services/finance_accounting/internal/security/policy.go#L347).
- function: Implementasikan command bisnis run-level Finance sesuai matriks melalui resolver/adapter dan handler terpisah untuk manusia. Pertahankan RPC workload privat. Gunakan satu kebijakan domain untuk state dan idempotensi publikasi run.
- logic: Validasi owner entity, posting final dan otorisasi penerbit. Gunakan principal aktor sah dan identitas workload yang dipetakan pada jalur otomatis. Simpan publication ID, run, penerbit, waktu dan audit secara atomik dengan outbox; pertahankan identitas penerbit asal saat membaca hasil lama. Bedakan status published, settlement dan ketersediaan dokumen HR.
- error_handling: Tanpa posting atau identitas penerbit sah, tolak tanpa event. Timeout setelah commit mengembalikan publication ID yang sama pada retry. HR/DRM offline menyisakan pengiriman yang dapat dipulihkan.
- security: Grant Finance pada legal entity dan recheck sensitif/MFA untuk operator. Hanya identitas sistem yang dipetakan eksplisit boleh menerbitkan via jalur otomatis. Hindari substitusi UUID resource sebagai principal dan delegasi penerbitan kepada HR.
- expected_result: Petugas Finance dapat menerbitkan run yang sah satu kali. Audit mengidentifikasi penerbit yang benar; HR menerima event untuk penerbitan PDF tanpa otorisasi pengeluaran dana.
- testing: Uji dua penerbit paralel, scope salah, run belum diposting, lost ACK, principal workload tak terpetakan dan HR/DRM mati. Baca kembali publication serta audit asli setelah retry.

## Action

1. Cocokkan baseline, kode, proto/SDL, status working tree dan kontrak lawan. Sepakati ID operasi/event, owner, scope, waktu, versi, ACK dan bentuk respons gagal sebelum menyalakan jalur baru. Pertahankan kontrak pemanggil lama selama rollout.
2. Kerjakan blok Request sesuai dependency. Tambahkan fixture kecil untuk kasus yang belum terliput. Unit test server gRPC hanya di layer service. Buktikan constraint/efek persisten melalui entrypoint normal dengan database Docker disposable. Untuk Gateway uji resolver/adapter serta HTTPS. Dilarang menulis grant IAM langsung ke DB.
3. Dari `backend/services/finance_accounting` jalankan `go vet ./...` dan `go test -race -count=1 ./...` pada modul yang berubah. Bila kontrak berubah jalankan generator resmi module; untuk Gateway gunakan `GOWORK=off go generate ./...` di `backend/gateway` dan periksa diff. Uji Caddy di `https://erp.localhost:8443/api/graphql` melalui dua replika. Login superadmin, pilih permission katalog, buat role, pasang permission dan assignment scoped via GraphQL IAM. Gunakan aktor berbeda bila SoD disyaratkan. Cabut grant sementara setelah tes.
4. Untuk jalur workload, cek listener, identitas peer, source event/request, aggregate/effect, versi, correlation ID dan status owner sebelum/sesudah retry. Bedakan receipt pending dari efek diterapkan; lampirkan bukti hasil aktual, termasuk kondisi gagal/tertunda. Hindari PII, voucher, isi slip atau secret pada log.
5. Laporkan per ID file yang diubah, kontrak pasangan, kasus lulus/gagal, rollback/replay dan SHA. Tandai DONE hanya setelah seluruh kriteria aktif terbukti, sesuai aturan repo. Tim yang menunggu dependency dapat meneruskan temuan independen.

## Context

- [agents/arsitektur.md](../../arsitektur.md) · [agents/boundary.md](../../boundary.md) · [agents/users.md](../../users.md) · [agents/how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [agents/how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md) · [agents/instruction_template.md](../../instruction_template.md) · [AGENTS.md](../../../AGENTS.md) · [docs/README.md](../../../docs/README.md)
- [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md) · [Konteks Finance & Accounting](../../context/finance_accounting/1-context.md)
- [Instruksi Gateway](../gateway/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi IAM](../iam/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Enterprise Foundation](../enterprise_foundation/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi HRM](../hrm/6-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Procurement & Sourcing](../procurement_sourcing/2-leader-perbaikan-korelasi-gateway-grpc-instructions.md)
