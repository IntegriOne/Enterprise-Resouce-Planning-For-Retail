# Pengujian guard HTTPS untuk seluruh field HRM

Setiap field HRM pada schema domain diuji melalui Caddy lokal di `/api/graphql` dengan CA pengembangan. Permintaan dibuat dari nama field, tipe argumen wajib, dan bentuk input pada schema checkout. Nilai argumen memakai fixture. Setiap field dikirim dalam permintaan terpisah tanpa cookie sesi dan tanpa token CSRF. Untuk hasil bertipe objek, permintaan memilih `__typename`.

| Operasi | Jumlah | Hasil aktual |
| --- | ---: | --- |
| Query HRM | 38 | Semua mengembalikan `UNAUTHENTICATED` |
| Mutation HRM | 39 | Semua mengembalikan `FORBIDDEN` dengan pesan token CSRF absen |
| Kegagalan validasi schema | 0 | Semua field dikenali |

Replika lokal memakai image `sha256:43b02d4252b40d25c19e07edf9e2d524964521cb1f4019087c07b76a26db1971`. Query Relay `node` adalah field bersama. Pemeriksaan ini menguji field khusus HRM. Bukti ini mencakup pemuatan schema dan guard akses. Keberhasilan pemanggilan RPC dari sesi employee tetap memerlukan fixture IAM resmi.

## Pengulangan setelah rollout daftar tugas offboarding

Kedua replika sehat dengan image `sha256:e6d6bdb54b5ae7335f03907a9fccb418e515651558f7a0efabe73a153c16faa3`. Pemeriksaan diulang dengan metode yang sama. Seluruh 41 Query mengembalikan `UNAUTHENTICATED` dan seluruh 43 Mutation mengembalikan `FORBIDDEN` dengan pesan CSRF absen. Tidak ada kegagalan validasi schema.

## Pengulangan setelah rollout pembatalan offboarding

Kedua replika sehat pada image `sha256:b7886cd49e25ac2bf6e785b1934a003e3b7f55abb99601af72154dc6bfab5b79`. Permintaan terpisah melalui Caddy dengan CA lokal, Origin yang sesuai, tanpa cookie dan tanpa token CSRF mencakup seluruh 88 field khusus HRM. Seluruh 41 Query mengembalikan `UNAUTHENTICATED`. Seluruh 47 Mutation mengembalikan `FORBIDDEN` karena token CSRF absen. Tidak ada kegagalan validasi schema. Uji berakun tetap memerlukan fixture IAM resmi.

## Pengulangan setelah rollout detail tugas offboarding

Seluruh 42 Query mengembalikan `UNAUTHENTICATED` dan seluruh 47 Mutation mengembalikan `FORBIDDEN` akibat token CSRF absen. Kedua replika Gateway sehat. [Image dan uji detail tugas](93-hrm-query-detail-tugas-offboarding.md).

## Regresi schema onboarding terkini

Melalui Caddy HTTPS development, seluruh 44 Query HRM mengembalikan `UNAUTHENTICATED` bagi sesi anonim. Seluruh 47 Mutation mengembalikan `FORBIDDEN` saat CSRF tidak diberikan. Introspeksi mengenali `hrmRequestEmployeeAccount` dan `createEmployeeUser`, tanpa Mutation `hrmLinkWorkerPrincipal`. Pemeriksaan ini membuktikan schema dan guard. Pengujian bisnis tetap dicatat per operasi.

Runner tersimpan pada `backend/gateway/internal/e2e/scripts/check_hrm_schema_guards.py`. Runner menunggu satu jendela rate limit sebelum mengulangi hanya operasi yang mendapat throttle.

## Regresi setelah wiring kandidat

Runner lulus untuk 46 dari 46 Query dan 50 dari 50 Mutation melalui HTTPS development. Query anonim mengembalikan `UNAUTHENTICATED`. Mutation tanpa CSRF mengembalikan `FORBIDDEN`. Introspeksi HTTPS juga mengonfirmasi `createEmployeeUser` menerima `worker_id: ID!` dan schema tidak memuat `hrmLinkWorkerPrincipal`. Regresi Go dengan race detector lulus pada 9 paket Gateway dan 24 paket IAM.

Setelah penambahan Query PDF slip gaji, guard HTTPS lulus untuk 47 dari 47 Query dan 50 dari 50 Mutation. Permintaan yang terkena rate limit berhasil diuji ulang setelah satu jendela rate limit.
