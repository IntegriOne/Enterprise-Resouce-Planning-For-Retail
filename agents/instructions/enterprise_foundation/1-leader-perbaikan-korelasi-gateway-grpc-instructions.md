# Leader backend — Enterprise Foundation — Perbaikan korelasi lintas management

- **ID** — 1
- **Owner** — Agent Enterprise Foundation sebagai pelaksana, leader backend sebagai peninjau. Perubahan di service lain dikoordinasikan dengan pemilik pada matriks.
- **Priority** — Kerjakan temuan P1 lebih dahulu. P2 menjaga kebenaran jalur pada konfigurasi produksi.
- **Status** — Siap dikerjakan.
- **Baseline** — Working tree lokal ERP Retail pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`. Baca ulang source yang dirujuk karena beberapa service masih merupakan perubahan lokal.

## Task

EF sudah menjadi sumber keputusan kelayakan scope dan pihak. Listener mTLS yang ditemukan khusus melayani tiga RPC Procurement. Konsumen lain memerlukan profil akses yang cocok, serta Gateway memerlukan bukti hasil penautan kontak yang tahan retry. Selesaikan tiap temuan melalui entrypoint aplikasi atau workload asli hingga hasilnya dapat dibaca kembali dari pemilik fakta. [Audit Enterprise Foundation](../../audit/enterprise_foundation/1-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Peta korelasi seluruh management](../../audit/gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md)

## Request

### EF-01 — P1 · Kontrak transport bukti untuk tiap pemanggil

- file: [grpc/procurement_workload.go](../../../backend/services/enterprise_foundation/transport/grpc/procurement_workload.go) · [grpc/server.go L72](../../../backend/services/enterprise_foundation/transport/grpc/server.go#L72) · [scopeverifier/ef_verifier.go](../../../backend/services/iam/infrastructure/scopeverifier/ef_verifier.go) · [enterprise_foundation/client.go L82](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L82) · [ef/client.go L48](../../../backend/services/finance_accounting/internal/integration/ef/client.go#L48). Path menunjukkan titik baca/perubahan; ubah kode milik Enterprise Foundation dan sepakati kontrak pasangan dengan owner terkait.
- function: Perluas konfigurasi listener/guard terautentikasi untuk profil IAM, HRM, Finance dan Gateway. Pertahankan `procurementWorkloadGuard` dan pembatasan RPC yang sudah aktif.
- logic: Petakan principal ke method yang benar-benar dibutuhkan per owner. IAM memerlukan ValidateLegalEntityScopeTarget, ValidateOrganizationUnitScopeTarget, ValidateFacilityScope dan ValidatePartyRoleScope. HR/Finance memerlukan lookup reference terverifikasi, termasuk unit/location sesuai use case. Pertahankan `target_id`, jenis, status aktif dan owner entity yang relevan. Sinkronkan kontrak TLS dengan GW-03, IAM-01, HR-03 dan FA-03.
- error_handling: Peer atau method di luar pemetaan ditolak. Scope tidak aktif dan layanan tidak tersedia memberi hasil yang dapat dibedakan. Matikan hanya jalur yang konfigurasinya rusak.
- security: Cocokkan sertifikat, workload dan konteks aktor terdelegasi sesuai jalur. Jangan membuka seluruh method EF ke sertifikat yang ditandatangani CA sama.
- expected_result: Setiap pemanggil dapat membaca fakta wajibnya saja melalui profil yang disepakati. Lookup live cukup untuk keputusan segar; dokumentasikan consumer, ack dan replay bila ada proyeksi event baru.
- testing: Uji matriks peer × method, kind/UUID pada respons, kredensial kosong, profil PS yang sudah berjalan dan entity/unit nonaktif.

### EF-02 — P1 · Pembuktian hasil tautan portal setelah timeout

- file: [enterprise_foundation/portal_provisioning.go L109](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L109) · [party/contact.go L57](../../../backend/services/enterprise_foundation/services/party/contact.go#L57) · [party/contact.go L292](../../../backend/services/enterprise_foundation/services/party/contact.go#L292) · [party/contact.go L410](../../../backend/services/enterprise_foundation/services/party/contact.go#L410). Path menunjukkan titik baca/perubahan; ubah kode milik Enterprise Foundation dan sepakati kontrak pasangan dengan owner terkait.
- function: Dukung recovery GW-01 melalui lookup existing berdasarkan IAM user dan pemeriksaan party/contact; tambahkan identitas operasi durable bila lookup yang ada belum membedakan attempt.
- logic: Buktikan binding exact untuk jalur add-contact dan existing-contact. Pengulangan binding yang sama mengembalikan hasil semula, konflik terhadap user atau party lain harus eksplisit. Pertahankan status pending_link sampai aktivasi sah. Ikat audit/outbox pada perubahan yang benar-benar terjadi.
- error_handling: Timeout mengizinkan reconcile/read-back. Hasil lookup gagal bukan bukti user belum tertaut. Tolak replay lintas party, hindari kontak tambahan ketika respons hilang.
- security: Batasi akses status pada party yang diotorisasi dan data minimum. Email bukan identitas operasi. Pertahankan revokasi kontak dan evaluasi portal oleh owner.
- expected_result: Dua panggilan dengan binding sama berakhir pada satu tautan dan satu perubahan bisnis; Gateway bisa mengetahui hasil commit secara otoritatif.
- testing: Uji add vs existing, commit lalu respons hilang, request paralel, status revoked, party lain dan user yang sudah terikat. Verifikasi audit serta jumlah contact yang dapat dibaca.

## Action

1. Cocokkan baseline, kode, proto/SDL, status working tree dan kontrak lawan. Sepakati ID operasi/event, owner, scope, waktu, versi, ACK dan bentuk respons gagal sebelum menyalakan jalur baru. Pertahankan kontrak pemanggil lama selama rollout.
2. Kerjakan blok Request sesuai dependency. Tambahkan fixture kecil untuk kasus yang belum terliput. Unit test server gRPC hanya di layer service. Buktikan constraint/efek persisten melalui entrypoint normal dengan database Docker disposable. Untuk Gateway uji resolver/adapter serta HTTPS. Dilarang menulis grant IAM langsung ke DB.
3. Dari `backend/services/enterprise_foundation` jalankan `go vet ./...` dan `go test -race -count=1 ./...` pada modul yang berubah. Bila kontrak berubah jalankan generator resmi module; untuk Gateway gunakan `GOWORK=off go generate ./...` di `backend/gateway` dan periksa diff. Uji Caddy di `https://erp.localhost:8443/api/graphql` melalui dua replika. Login superadmin, pilih permission katalog, buat role, pasang permission dan assignment scoped via GraphQL IAM. Gunakan aktor berbeda bila SoD disyaratkan. Cabut grant sementara setelah tes.
4. Untuk jalur workload, cek listener, identitas peer, source event/request, aggregate/effect, versi, correlation ID dan status owner sebelum/sesudah retry. Bedakan receipt pending dari efek diterapkan; lampirkan bukti hasil aktual, termasuk kondisi gagal/tertunda. Hindari PII, voucher, isi slip atau secret pada log.
5. Laporkan per ID file yang diubah, kontrak pasangan, kasus lulus/gagal, rollback/replay dan SHA. Tandai DONE hanya setelah seluruh kriteria aktif terbukti, sesuai aturan repo. Tim yang menunggu dependency dapat meneruskan temuan independen.

## Context

- [agents/arsitektur.md](../../arsitektur.md) · [agents/boundary.md](../../boundary.md) · [agents/users.md](../../users.md) · [agents/how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [agents/how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md) · [agents/instruction_template.md](../../instruction_template.md) · [AGENTS.md](../../../AGENTS.md) · [docs/README.md](../../../docs/README.md)
- [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md) · [Konteks Enterprise Foundation](../../context/enterprise_foundation/1-context.md)
- [Instruksi Gateway](../gateway/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi IAM](../iam/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi HRM](../hrm/6-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Finance & Accounting](../finance_accounting/5-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Procurement & Sourcing](../procurement_sourcing/2-leader-perbaikan-korelasi-gateway-grpc-instructions.md)
