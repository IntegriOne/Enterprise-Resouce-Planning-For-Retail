# DDD Bounded Context, Data Ownership dan Inter-Service

> **Baseline 12 November 2026.** Kebijakan keamanan/izin harus cocok dengan [agents/arsitektur.md](../../agents/arsitektur.md); ownership ditentukan oleh [boundary management](../../agents/boundary.md). Status runtime perlu dibuktikan dari kode/tes terbaru.

## Hard invariants
**Service yang memiliki fakta bisnis mengoreksi faktanya.** SQL, DSN, join, FK dan migration lintas database serta impor Go `internal` management lain **dilarang**. Satu database per management boleh berbagi server PostgreSQL fisik dengan kredensial isolasi; tidak ada kewajiban satu cluster per microservice. Referensi lintas service berupa ID opaque + owner/source + version/status/expiry. Untuk registrasi katalog RBAC agar otomatis tersimpan di production, masing-masing management hanya boleh mendaftarkan manifest deklaratif statis di `backend/shared/rbac/catalogs/{management}.json` (tanpa logika bisnis) agar dieksekusi satu pintu secara idempoten oleh IAM ke database `iam`.

| Fakta | Management pemilik | Konsumen harus |
|---|---|---|
| Human/workload principal, sessions, access grants | IAM | Gunakan contract/verified context |
| Legal entity, unit, supplier/customer party | Enterprise Foundation | Referensi entity ID dan scope, bukan FK |
| Item, varian dan revision produk | PLM | Ambil versi/rilis yang benar |
| Purchase order, ASN dan komitmen pembelian | Procurement | Gunakan PO lines dan status owner |
| Stock, receiving, hold, pick/pack/dispatch | WMS | Pakai stock/hold actual qty/version |
| Harga, promosi dan eligibility | Pricing | Ambil keputusan harga plus currency/expiry |
| Sales order, acceptance, cancellation, return decision | OMS | Simpan order ref & source version |
| Channel listing, cart, quote dan order submission | Commerce | Tidak membuat sales order kedua |
| AP/AR, receipt uang, GL, pajak/buku | Finance | Upload bukti ≠ payment verified |
| Worker lifecycle, kontrak kerja | HRM | Worker ref ≠ principal IAM |
| CRM lead/customer marketing | CRM | Consent/source dan audience |
| Case/ticket pelanggan | Customer Service | Case scope; tidak mengubah order/receipt secara sepihak |
| Dokumen, versi, hash, retention | DRM | Reference terikat owner dan ACL |

## Kontrak RPC minimum
Audience + workload identity, action/permission, realm/legal entity/party/object IDs, request/response schema version, source revision, stable idempotency key, expiry/timeout/deadline, error classification dan recovery path. Unknown result karena timeout ≠ business rejected. Penerima harus dapat mencari hasil command yang sama.

## Transaksi asinkron
1. Writer mengubah aggregate dan menulis outbox dalam **satu transaksi** PostgreSQL.
2. Worker mengirim event; duplikat dan out-of-order diperbolehkan oleh transport.
3. Consumer memeriksa source/aggregate version/digest, dedupe lewat inbox dan menulis efek bisnisnya sendiri secara lokal.
4. Reconciler membandingkan fakta owner dan memperbaiki read model melalui command yang sah. Same version/different payload adalah conflict, bukan overwrite.

## Contoh
Pelanggan memesan dari website: Commerce menyimpan submission `PENDING`, OMS menerima/menolak sales order, WMS menjalankan hold/fulfillment, Finance menentukan receipt, Commerce menampilkan status owner berversi. Tidak ada shared transaction antar-DB; jangan membalas accepted hanya karena event terkirim.
