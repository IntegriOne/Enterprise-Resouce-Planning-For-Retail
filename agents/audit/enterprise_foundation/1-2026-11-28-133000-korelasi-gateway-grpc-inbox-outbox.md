# Audit koneksi Enterprise Foundation — Gateway, gRPC dan korelasi bisnis

**Kesimpulan.** EF sudah menjadi sumber keputusan kelayakan scope dan pihak. Listener mTLS yang ditemukan khusus melayani tiga RPC Procurement. Konsumen lain memerlukan profil akses yang cocok, serta Gateway memerlukan bukti hasil penautan kontak yang tahan retry.

**Lingkup pemeriksaan.** Pembacaan statis checkout bersama pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`, termasuk source lokal yang belum di-commit. Bukti runtime terdahulu disebut terpisah bila tersedia. Tidak dilakukan transaksi HTTPS, transfer, deployment atau perubahan aplikasi pada audit ini.

## Jalur yang terkonfirmasi dari source

`ScopePolicy` mengembalikan kind, target_id dan eligibility; unit/facility bergantung pada owner legal entity. PS memeriksa status legal entity dan party serta role supplier/vendor. Perubahan party role dan kontak portal menulis outbox di transaksi service. Relay memiliki standby, log dan webhook. Mode log adalah diagnostik. Belum terbukti ada consumer IAM yang memproses perubahan ini melalui inbox; validasi scope saat pemberian role menggunakan pull. Catat kebutuhan invalidasi hanya bagi konsumen yang menyimpan proyeksi atau keputusan lama. [service/scope_policy.go L73](../../../backend/services/enterprise_foundation/domain/service/scope_policy.go#L73) · [party/role.go L104](../../../backend/services/enterprise_foundation/services/party/role.go#L104) · [party/role.go L253](../../../backend/services/enterprise_foundation/services/party/role.go#L253) · [party/contact.go L488](../../../backend/services/enterprise_foundation/services/party/contact.go#L488) · [cmd/wiring.go L140](../../../backend/services/enterprise_foundation/cmd/wiring.go#L140) · [foundation/client.go L79](../../../backend/services/procurement_sourcing/internal/integration/foundation/client.go#L79)

**Permukaan kontrak.** Parser gqlparser membaca 64 root Query dan 48 root Mutation dari SDL `backend/gateway/graphql/enterprise_foundation/`. Descriptor dari `buf build` memuat 112 RPC dalam 5 service proto milik Enterprise Foundation. RPC privat/callback ikut terhitung. Angka ini mengukur deklarasi dan tidak dipakai sebagai bukti E2E.

## Kecocokan dengan rancangan bisnis

Procurement §23 membutuhkan entity, unit, lokasi, party serta role; PLM §15/19 membutuhkan identitas entity dan brand owner; Inventory §2 bergantung pada lokasi dan batas pemilik stok. EF menyediakan lookup untuk fakta tersebut. PS telah mengonsumsi entity/party/role namun port PS belum mendukung unit/lokasi (PS-04). Tidak ada logic.md khusus EF; kelayakan role dan pemisahan owner dikonfirmasi dari implementasi EF. [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md)

## Temuan dan dampak

### EF-01 — P1 · Kontrak transport bukti untuk tiap pemanggil

Listener privat memverifikasi URI sertifikat Procurement dan membatasi tiga method legal entity/party/role. IAM membutuhkan validasi scope, HR membutuhkan fakta organisasi dan Finance membaca master EF. Ketiganya tidak dapat memakai listener Procurement apa adanya. Server EF biasa memakai interceptor konteks internal; opsi TLS pada constructor yang diperiksa diikat pada `ProcurementTLS`. [grpc/procurement_workload.go](../../../backend/services/enterprise_foundation/transport/grpc/procurement_workload.go) · [grpc/server.go L72](../../../backend/services/enterprise_foundation/transport/grpc/server.go#L72) · [scopeverifier/ef_verifier.go](../../../backend/services/iam/infrastructure/scopeverifier/ef_verifier.go) · [enterprise_foundation/client.go L82](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go#L82) · [ef/client.go L48](../../../backend/services/finance_accounting/internal/integration/ef/client.go#L48)

### EF-02 — P1 · Pembuktian hasil tautan portal setelah timeout

`CompletePortalUserProvisioning` memakai `RequestPartyPortalLink` untuk kontak yang ada dan `AddPartyContact` bila kontak baru. EF juga menyediakan `ListContactsByIAMUserID` dan fakta portal. Gateway saat ini dapat meminta kompensasi sebelum menggunakan sumber fakta tersebut. Efek yang ambigu juga perlu dibedakan dari akses portal yang benar-benar activated. [enterprise_foundation/portal_provisioning.go L109](../../../backend/gateway/internal/adapter/enterprise_foundation/portal_provisioning.go#L109) · [party/contact.go L57](../../../backend/services/enterprise_foundation/services/party/contact.go#L57) · [party/contact.go L292](../../../backend/services/enterprise_foundation/services/party/contact.go#L292) · [party/contact.go L410](../../../backend/services/enterprise_foundation/services/party/contact.go#L410)

## Tindak lanjut

[Instruksi per temuan, pemilik, pengujian dan kriteria penutupan](../../instructions/enterprise_foundation/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md). [Matriks lintas management dan perhitungan permukaan kontrak](../gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md).

**Jejak baseline lokal.** SHA-256 gabungan 14 file backend/logic yang ditautkan pada audit ini adalah `b66f0f484d32b1834a8ecd932f54aea179654c13b339dec31676a04c97f7ca46`. Urutkan path relatif unik, lalu untuk tiap path gabungkan UTF-8 path, byte NUL dan digest SHA-256 biner dari isi file. Hash gabungan membantu mencocokkan checkout ketika source lokal belum dipublikasikan.
