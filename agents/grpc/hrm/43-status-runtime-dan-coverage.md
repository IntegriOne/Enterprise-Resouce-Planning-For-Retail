# List operasi

HRM mendeklarasikan 1.015 RPC bisnis pada 22 service gRPC privat. Tiga RPC workload tambahan melayani integrasi IAM dan Finance. Gateway menyediakan **52 Query dan 61 Mutation** HRM, di luar root bersama seperti `node`. [Inventaris gRPC](1-inventaris-operasi-grpc-hrm.md) · [Matriks Gateway](26-status-wiring-gateway-hrm.md)

## Penjelasan hasil verifikasi

| Cakupan | Bukti |
| --- | --- |
| Runtime gRPC HRM | 301 tes utama lulus di PostgreSQL disposable dengan race detector. Empat riwayat migrasi terverifikasi dan konflik data membatalkan perubahan skema serta ledger. |
| Gateway HTTPS | Guard sesi lulus untuk 52 Query. Guard CSRF lulus untuk 61 Mutation. |
| IAM, EF, dan HRM | Login dan MFA superadmin, grant lewat GraphQL, verifikasi scope badan hukum, provisioning berulang, penautan pekerja, aktivasi, terminasi, pencabutan sesi, dan clearance mTLS lulus dari arsip main. |

[Laporan HTTPS dari main](../../e2e/gateway/183-hrm-mutation-provisioning-dan-terminasi-main.md)

## Butuh GraphQL

Bukti keberhasilan bisnis per operasi dan skenario yang masih menunggu provider tercatat pada [52 Query](31-bukti-https-query-gateway.md) dan [61 Mutation](30-bukti-https-mutation-gateway.md). Guard HTTPS memverifikasi akses dan kontrak transport. Bukti positif per operasi mengikuti matriks masing-masing.

## RBAC

Pengujian memakai akun superadmin dan MFA. Pengikatan role, permission, dan scope uji dilakukan melalui GraphQL IAM. Slip mandiri membutuhkan principal employee tertaut dan grant SELF. Cetak ulang oleh petugas membutuhkan grant badan hukum dan MFA. [Matriks izin](28-izin-iam-untuk-wiring-gateway.md)

## Dependensi pemilik dan pekerjaan terbuka

- Finance memverifikasi anggaran requisition dan workforce, menerbitkan payroll, menangani transfer dan settlement, serta memberi clearance offboarding. HRM mencatat rincian dan hasil otorisasi. [Handoff Finance](25-kebutuhan-integrasi-finance-payroll.md)
- DRM menyimpan versi PDF privat dan immutable, memeriksa akses versi, serta menyediakan bukti kontrak bertanda tangan dan artefak aturan payroll. [Handoff DRM](24-kebutuhan-integrasi-drm.md) · [Cetak ulang slip](39-cetak-ulang-slip-gaji-mantan-karyawan.md)
- Kalkulator runtime memerlukan sumber snapshot immutable dan artefak aturan. Penyelesaian tugas offboarding lintas owner memerlukan konfirmasi pemilik tugas. [Kalkulator](27-kebutuhan-kalkulator-payroll-runtime.md) · [Offboarding](13-operasi-offboarding.md)

## Verifikasi setelah pekerjaan HRM dilanjutkan

Snapshot `4c9396d` lolos pemeriksaan inventaris terhadap proto, handler, dan izin. Seluruh 1.015 RPC bisnis tercatat tanpa selisih. Ulang uji HTTPS dengan `--termination --guards` lulus untuk 52 Query, 61 Mutation, provisioning idempoten, aktivasi dan terminasi, serta clearance mTLS. Uji gRPC offboarding dan validator pada PostgreSQL disposable juga lulus.

Provider Finance dan DRM ditunda sampai kedua management dibangun. Konfirmasi tugas offboarding milik HRM kini dapat dicatat oleh petugas HR. Pembekuan snapshot input payroll masih terbuka. Jalur positif untuk tahap yang membutuhkan provider tetap tercatat sebagai uji fixture. [Bukti tugas offboarding](13-operasi-offboarding.md#sumber-konfirmasi-penyelesaian-tugas) · [Snapshot payroll](27-kebutuhan-kalkulator-payroll-runtime.md#penyelarasan-snapshot-dengan-pembuatan-run)
