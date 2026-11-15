# Code Architecture, Standar dan Kolaborasi

> **Pedoman engineering.** Ikuti [AGENTS.md](../../AGENTS.md), keputusan [arsitektur](../../agents/arsitektur.md), [ownership](../../agents/boundary.md) dan [scope](../01-product/vision-and-scope.md). Bukti tes terpisah dari rencana tes.

## DDD vertical slice
Urutan lazim: DBML/invariant → baseline migration `000001` (pra-rilis) → entity/value/status → service/use case → repository owner → proto/gRPC → GraphQL adapter/model/resolver → service tests dan E2E. Jangan membuat satu file besar untuk seluruh management; controller/resolver hanya boundary transport, bukan tempat bisnis.

Aturan: satu database penulis per aggregate, tidak ada cross-DB SQL/FK atau import package `internal` lain. Paket `backend/shared/` hanya untuk mekanisme lintas service: authcontext/authz, logging, retry/deadline, bukan aturan bisnis Finance atau OMS.

## Kontrak function
Nama method aktual sesuai source, bukan diciptakan agar documentation terlihat penuh. Tulis business invariant, input, output, reason/error code, permission/resource/scope, transaction boundary, idempotency, outbox dan recovery. Desimal uang/qty memakai presisi yang benar, bukan `float64`; waktu event UTC dan timezone bisnis diberi label; patch null/omitted/zero harus berbeda bila bermakna.

## Git satu branch (keputusan owner)
1. Cek `git status` dan `git fetch origin`; `git pull --ff-only origin main` hanya jika working tree aman.
2. Jangan edit file generated secara manual; koordinasikan `backend/gateway/gqlgen.yml`, `backend/contracts/gen/`, `backend/go.work`, root resolvers, `backend/VERSION` dan `backend/CHANGELOG.md`.
3. Commit hanya path scope tugas, format `type(management/domain): tujuan` dan pesan berisi bukti tes. Jangan `git add -A` pada workspace multi-agent tanpa audit.
4. Jika remote bergerak, `git fetch` kemudian rebase **hanya commit lokal yang belum dipush**, uji ulang, `git push origin main` non-force.
5. Cek SHA di GitHub, laporkan changed files, pass/fail/blocker dan working history; tidak mengklaim push sebelum terverifikasi.

Tidak boleh `git push --force`, membuat branch/PR baru sebagai alur rutin, reset hard/clean yang berpotensi menghapus kerja agen lain, atau menjalankan migrasi produksi tanpa owner.
