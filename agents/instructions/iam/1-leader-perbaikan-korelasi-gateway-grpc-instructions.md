# Leader backend — IAM — Perbaikan korelasi lintas management

- **ID** — 1
- **Owner** — Agent IAM sebagai pelaksana, leader backend sebagai peninjau. Perubahan di service lain dikoordinasikan dengan pemilik pada matriks.
- **Priority** — Kerjakan temuan P1 lebih dahulu. P2 menjaga kebenaran jalur pada konfigurasi produksi.
- **Status** — Siap dikerjakan.
- **Baseline** — Working tree lokal ERP Retail pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`. Baca ulang source yang dirujuk karena beberapa service masih merupakan perubahan lokal.

## Task

IAM sudah memverifikasi lima jenis scope eksternal di EF dan menyelesaikan provisioning pegawai dengan membaca linkage HR. Identitas target pada respons scope dan pemulihan pengiriman request pegawai masih memerlukan penguatan. Selesaikan tiap temuan melalui entrypoint aplikasi atau workload asli hingga hasilnya dapat dibaca kembali dari pemilik fakta. [Audit IAM](../../audit/iam/17-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md) · [Peta korelasi seluruh management](../../audit/gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md)

## Request

### IAM-01 — P1 · Keputusan scope belum dicocokkan dengan ID yang diminta

- file: [scopeverifier/ef_verifier.go L124](../../../backend/services/iam/infrastructure/scopeverifier/ef_verifier.go#L124) · [v1/common.proto L102](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/common.proto#L102) · [handler/mapper_common.go L253](../../../backend/services/enterprise_foundation/transport/grpc/handler/mapper_common.go#L253) · [cmd/wiring.go](../../../backend/services/iam/cmd/wiring.go). Path menunjukkan titik baca/perubahan; ubah kode milik IAM dan sepakati kontrak pasangan dengan owner terkait.
- function: Ubah `VerifyScope`, `verifyPartyRoleScope`, `validateDecision` dan config/dial EFVerifier. Pertahankan kontrak `ScopeOwnerVerifier` pada caller.
- logic: Teruskan expected UUID dan kind. Validasi UUID normalisasi dan kecocokan sebelum menulis user-role atau group-role. Cek legal entity apabila operasi juga mengetahui entity yang diminta; party global tidak diberi entity fiktif. Pakai kanal terautentikasi sesuai EF-01.
- error_handling: Keputusan target tidak layak tetap ditolak. Respons kosong, ID hilang/salah atau kind tidak cocok gagal tertutup, dengan error kontrak owner yang bisa dibedakan dari timeout. Pertahankan deadline headroom.
- security: Sumber fakta tetap EF; metadata caller ditandatangani dan identitas pemanggil diverifikasi pada listener. Jangan membuat grant/outbox untuk bukti salah.
- expected_result: Setiap assignment scoped yang berhasil memiliki keputusan bagi UUID yang persis sama. GLOBAL dan SELF yang valid tidak tergantung koneksi EF.
- testing: Uji lima kind dengan UUID sama/beda, reply nil, owner nonaktif, timeout, peer salah dan dua operasi GraphQL IAM `assignUserRole`/`assignGroupRole`. Pastikan grant dan revisi izin tidak berubah pada penolakan.

### IAM-02 — P1 · Jalur penerimaan provisioning untuk pemulihan durable

- file: [provisioning/employee.go](../../../backend/services/iam/services/provisioning/employee.go) · [handler/employee_provisioning_handler.go](../../../backend/services/iam/transport/grpc/handler/employee_provisioning_handler.go) · [cmd/requisition_mandate.go](../../../backend/services/iam/cmd/requisition_mandate.go) · [human_resource/access_request.go L100](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L100). Path menunjukkan titik baca/perubahan; ubah kode milik IAM dan sepakati kontrak pasangan dengan owner terkait.
- function: Sediakan penerimaan privat untuk relay HR-01 atau reconciler IAM yang menarik request HR terverifikasi. Pakai logika receipt/konflik bersama dengan `Receive`.
- logic: Ikat source request, worker, entity, pemohon, event lifecycle, katalog, waktu dan versi/digest. Verifikasi fakta langsung ke HR atau envelope workload yang tervalidasi. Otorisasi pembuat saat permohonan dan otorisasi operator saat membuat akun adalah dua keputusan. Receipt pending sendiri tidak memberikan hak akses.
- error_handling: Duplikat identik mengembalikan receipt yang sama, binding berubah menghasilkan konflik, owner unavailable tetap dapat diulang. Pisahkan readiness verifier worker-account dari nama opsi terminasi atau nyatakan dependency eksplisit di startup.
- security: Batasi workload/method dengan mTLS dan identitas yang tepat; tidak menyimpan cookie, password atau impersonasi sesi di payload. Pertahankan `Complete` dengan bukti account dari HR.
- expected_result: Satu request HR menghasilkan satu receipt walau jalur langsung dan relay bersaing. Akun hanya linked setelah hubungan di HR terverifikasi.
- testing: Uji concurrent delivery, respons receipt hilang, request dicabut, actor/worker/entity salah, dan verifier HR unavailable sebelum complete. Verifikasi pending tidak mengaktifkan akun.

## Action

1. Cocokkan baseline, kode, proto/SDL, status working tree dan kontrak lawan. Sepakati ID operasi/event, owner, scope, waktu, versi, ACK dan bentuk respons gagal sebelum menyalakan jalur baru. Pertahankan kontrak pemanggil lama selama rollout.
2. Kerjakan blok Request sesuai dependency. Tambahkan fixture kecil untuk kasus yang belum terliput. Unit test server gRPC hanya di layer service. Buktikan constraint/efek persisten melalui entrypoint normal dengan database Docker disposable. Untuk Gateway uji resolver/adapter serta HTTPS. Dilarang menulis grant IAM langsung ke DB.
3. Dari `backend/services/iam` jalankan `go vet ./...` dan `go test -race -count=1 ./...` pada modul yang berubah. Bila kontrak berubah jalankan generator resmi module; untuk Gateway gunakan `GOWORK=off go generate ./...` di `backend/gateway` dan periksa diff. Uji Caddy di `https://erp.localhost:8443/api/graphql` melalui dua replika. Login superadmin, pilih permission katalog, buat role, pasang permission dan assignment scoped via GraphQL IAM. Gunakan aktor berbeda bila SoD disyaratkan. Cabut grant sementara setelah tes.
4. Untuk jalur workload, cek listener, identitas peer, source event/request, aggregate/effect, versi, correlation ID dan status owner sebelum/sesudah retry. Bedakan receipt pending dari efek diterapkan; lampirkan bukti hasil aktual, termasuk kondisi gagal/tertunda. Hindari PII, voucher, isi slip atau secret pada log.
5. Laporkan per ID file yang diubah, kontrak pasangan, kasus lulus/gagal, rollback/replay dan SHA. Tandai DONE hanya setelah seluruh kriteria aktif terbukti, sesuai aturan repo. Tim yang menunggu dependency dapat meneruskan temuan independen.

## Context

- [agents/arsitektur.md](../../arsitektur.md) · [agents/boundary.md](../../boundary.md) · [agents/users.md](../../users.md) · [agents/how_to_code_grpc_managements.md](../../how_to_code_grpc_managements.md) · [agents/how_to_code_wiring_grpc_managements_to_gateway.md](../../how_to_code_wiring_grpc_managements_to_gateway.md) · [agents/instruction_template.md](../../instruction_template.md) · [AGENTS.md](../../../AGENTS.md) · [docs/README.md](../../../docs/README.md)
- [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md) · [Konteks IAM](../../context/iam/1-context.md)
- [Instruksi Gateway](../gateway/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Enterprise Foundation](../enterprise_foundation/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi HRM](../hrm/6-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Finance & Accounting](../finance_accounting/5-leader-perbaikan-korelasi-gateway-grpc-instructions.md) · [Instruksi Procurement & Sourcing](../procurement_sourcing/2-leader-perbaikan-korelasi-gateway-grpc-instructions.md)
