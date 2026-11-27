# List operasi

Penyelarasan berlaku untuk seluruh Query, Mutation, dan tipe SDL di `backend/gateway/graphql/human_resource/`. Pada pemeriksaan awal ada 49 Query dan 55 Mutation berawalan `hrm` di worktree, termasuk edit requisition yang sedang diuji. Hanya tipe pekerja yang mengimplementasikan `Node`. Belum ada tipe `Connection` di SDL domain ini.

## Penjelasan kontrak

Ikuti pola `schema.graphql`, `iam_user.graphql`, `iam_rbac.graphql`, dan `ef_organization_unit.graphql`. Nama ekspor menyebut entitas bisnis, misalnya `Worker`, `workers`, `Requisition`, `requisitions`, dan `updateRequisition`. Hilangkan awalan `HRM` atau `hrm` dari nama tipe dan operasi serta periksa benturan nama pada schema gabungan.

## Butuh GraphQL Query

Entitas persisten memakai `Node` dengan `id` global yang dapat dibaca kembali melalui `node(id: ...)`. Daftar berpaginasi memakai `Connection` dan `Edge` dengan `edges { cursor node }`, `nodes`, `pageInfo`, dan `total_count`. Argumen Relay memakai `after`, `before`, `first`, dan `last`. Objek hasil perhitungan atau receipt diperiksa menurut identitas dan kebutuhan refetch masing-masing.

## Butuh GraphQL Mutation

Operasi domain dideklarasikan melalui `extend type Mutation`. Nama tipe input dan payload mengikuti operasinya dan payload mengimplementasikan `MutationResult` sesuai schema utama. Nama dan bentuk request konsumen serta bukti E2E ikut dimigrasikan.

## RBAC dan data UI

Pertahankan pemeriksaan permission dan scope pada operasi setelah pergantian nama. ID global hanya untuk referensi. UI menggunakan angka atau kode dokumen, nama pekerja, jabatan, dan unit sebagai label.

## Status penyelarasan nama

Worktree telah mengganti 49 nama Query, 55 nama Mutation, dan 128 nama tipe. Alias lama tidak dipublikasikan pada schema baru. Resolver, dokumen GraphQL frontend, dan pengujian Gateway dalam repositori ikut memakai nama baru. Paginasi Relay dan cakupan Node di luar pekerja masih perlu diterapkan. Bukti HTTPS dengan nama lama merekam kontrak pada saat masing-masing tes dijalankan.

[Uji HTTPS nama baru](../../e2e/gateway/153-hrm-query-mutation-nama-sdl-tanpa-prefix.md).

## Status penyelarasan paginasi

Ke-24 Query daftar memakai argumen `after`, `before`, `first`, dan `last`. Koneksi mengembalikan `edges`, `nodes`, `pageInfo`, dan `total_count`. RPC internal masih menerima limit dan offset. Gateway menggunakan codec cursor, parser, dan algoritme jendela yang sama dengan Enterprise Foundation. Frontend pekerja, kandidat, dan slip gaji membaca `nodes` dan memakai cursor dari `pageInfo` untuk navigasi.

Sebanyak 24 entitas dengan Query detail mengimplementasikan `Node` dan mendukung refetch melalui `node(id: ...)` dengan RBAC yang sama. Tipe dengan detail menerima UUID atau GID bertipe cocok. `_id` menyediakan UUID untuk referensi internal. Snapshot slip milik pengguna dan respons posting payroll mengikuti operasi khusus masing-masing. Transfer assignment mendukung Query detail, riwayat Relay per pekerja, dan Node.

[Bukti HTTPS paginasi Relay](../../e2e/gateway/154-hrm-query-koneksi-relay-requisition.md).

[Verifikasi refetch, browser, dan rollout dev](../../e2e/gateway/155-hrm-query-node-refetch-dan-ui-pekerja.md).

Pemeriksaan HTTPS seluruh root field setelah migrasi menolak 51 Query tanpa sesi dan 55 Mutation tanpa CSRF. Skrip membaca nama langsung dari SDL. [Laporan guard HTTPS](../../e2e/gateway/158-hrm-query-mutation-guard-kontrak-relay.md).

Skrip guard Gateway sekarang memeriksa kontrak SDL sebelum mengirim request HTTPS. Pemeriksaan menolak kemunculan `HRM` tanpa membedakan kapitalisasi di nama operasi atau tipe, memeriksa `id: ID!` pada Node, serta memeriksa bentuk Edge, Connection, argumen Relay, dan payload Mutation. `MyPayslip` tetap memakai query privat per sesi untuk identitas hasil. Pemeriksaan SDL terakhir menemukan 24 Node dan 24 Connection pada 52 Query serta 55 Mutation.

[Verifikasi SDL dan guard HTTPS](../../e2e/gateway/168-hrm-query-mutation-kontrak-sdl-relay.md).
