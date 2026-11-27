# List operasi

Schema Gateway memuat 52 Query dan 55 Mutation HRM serta jalur Relay `node` untuk 24 entitas. Sebanyak 119 nama RPC pada matriks wiring cocok dengan pemanggilan adapter dan deklarasi proto. [Matriks wiring](26-status-wiring-gateway-hrm.md).

## Butuh GraphQL Query dan Mutation

| Service | Kandidat Query | Kandidat Mutation | RPC sudah dipanggil Gateway | Kandidat RPC tersisa |
| --- | ---: | ---: | ---: | ---: |
| Leave | 28 | 13 | 15 | 26 |
| Payroll | 88 | 26 | 23 | 91 |
| Personnel | 52 | 23 | 22 | 53 |
| Access Request | 12 | 3 | 3 | 12 |
| Workforce Planning | 36 | 20 | 27 | 29 |
| Offboarding | 19 | 12 | 12 | 19 |
| 16 service lain | 434 | 227 | 0 | 661 |
| Total | 669 | 324 | 102 | 891 |

Jumlah kandidat berasal dari inventaris 1.011 RPC. Sebanyak 18 RPC nonaktif dikeluarkan dari kolom kandidat. Satu field GraphQL bisa memilih beberapa RPC berdasarkan filter. Angka kandidat tersisa mengukur kontrak gRPC yang belum terhubung. Pemilihan field yang perlu dipublikasikan mengikuti kebutuhan produk dan persetujuan pemilik domain. [Inventaris per service](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Izin yang sudah dipetakan tercatat pada handoff IAM. Uji HTTPS dengan assignment resmi IAM dan fixture dua badan hukum telah lulus untuk pekerja dan workforce plan. [Handoff IAM](28-izin-iam-untuk-wiring-gateway.md) · [Uji HTTPS](../../e2e/gateway/34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md).

## Dependency workflow

- Payroll dan slip mandiri memerlukan bukti penerbitan Finance dan akses privat DRM untuk PDF. Query Gateway sudah mengembalikan metadata slip dan ID versi. [Finance](25-kebutuhan-integrasi-finance-payroll.md) · [DRM](24-kebutuhan-integrasi-drm.md).
- Personnel sudah memuat pembuatan draf kontrak, pencatatan tanda tangan, dan aktivasi. Uji positif tanda tangan memerlukan verifier DRM. Pengaitan principal memerlukan bukti provisioning IAM. [Personnel](17-operasi-personnel.md).
- Workforce Planning sudah memuat alur draf hingga persetujuan. Uji positif persetujuan memerlukan verifier anggaran Finance. [Handoff anggaran](../../notes/finance_accounting/2-human_resource-workforce-budget.md).

## Verifikasi

Jumlah field dihitung dari blok Query dan Mutation pada schema domain HRM. Nama RPC dari tabel wiring dibandingkan dengan deklarasi proto dan referensi method klien pada adapter Go di luar file pengujian. Pemeriksaan ini membuktikan konsistensi inventaris. Hasil runtime dan otorisasi dicatat pada laporan E2E masing-masing.

## Rekonsiliasi setelah rollout gRPC HRM

Matriks terkini mencakup 91 field khusus HRM dan 102 nama RPC. `LinkWorkerPrincipal` dipanggil oleh alur IAM melalui gRPC privat. Scope hitungan tabel mengikuti field HRM dan Relay `node`.

Pada rollout terdahulu, seluruh 89 field pada schema khusus HRM cocok dengan entri matriks dan nama method resolver. Seluruh 100 nama RPC di matriks ditemukan pada deklarasi proto dan referensi adapter Go, termasuk method expression yang diteruskan ke helper command. Backend HRM development kini berjalan pada image checkout dengan baseline `000001`. Seluruh field diuji untuk guard sesi dan CSRF. Uji berakun berikutnya memakai katalog uji lokal dan assignment resmi IAM. [Audit rollout](../../audit/human_resource/184-2026-11-27-020305-kompatibilitas-rollout-grpc.md) · [Uji guard HTTPS](../../e2e/gateway/84-hrm-query-mutation-guard-seluruh-schema.md).


## Status HTTPS berakun

- Seluruh 42 Query diuji dengan role lokal. Daftar kosong, detail dengan ID yang tidak ada, guard MFA, dan guard scope self menghasilkan respons sesuai skenario. [Matriks Query](../../e2e/gateway/95-hrm-query-seluruh-schema-superadmin-https.md).
- 47 Mutation sebelum penambahan pembuatan kandidat diuji dengan CSRF valid memakai akun superadmin tanpa MFA. Hasil mencakup validasi input dan penolakan MFA. [Matriks Mutation](../../e2e/gateway/96-hrm-mutation-seluruh-schema-superadmin-https.md).
- Alur penyimpanan positif sudah dibuktikan untuk draf workforce plan dan pekerja, termasuk duplikat dan koreksi. [Workforce Planning](../../e2e/gateway/97-hrm-mutation-referensi-badan-hukum-workforce.md) · [Personnel](../../e2e/gateway/98-hrm-query-mutation-pekerja-dan-guard-offboarding.md).
- Assignment berscope badan hukum membatasi detail, daftar dan total paginasi. [Personnel](../../e2e/gateway/34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md) · [Workforce Planning](../../e2e/gateway/99-hrm-query-workforce-plan-scope-badan-hukum.md).

Uji operasi sensitif memerlukan sesi MFA yang valid. Fixture MFA sebaiknya memakai principal uji dengan faktor sendiri atau database IAM terisolasi, agar konfigurasi autentikasi superadmin development terjaga.

[Bukti HTTPS per Mutation](30-bukti-https-mutation-gateway.md) mencatat 42 jalur berhasil dan lima operasi yang masih memerlukan dependency runtime.

[Bukti HTTPS per Query](31-bukti-https-query-gateway.md) mencatat 42 jalur dengan data dan dua Query slip yang masih memerlukan integrasi Finance dan DRM.
