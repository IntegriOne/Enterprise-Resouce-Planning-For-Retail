# Audit gerbang integrasi HRM lintas paket

## Cakupan

Pada checkout `81360dc`, seluruh 16 paket Go HRM yang memiliki `*_integration_test.go` dijalankan satu per satu dengan `-race -tags=integration -count=1`. Setiap paket memakai database PostgreSQL Docker bernama `_test` yang baru dibuat dan langsung dihapus setelah tes. Paket yang lulus: `db`, `infrastructure/outbox`, serta handler gRPC `analytics`, `benefits`, `career`, `compensation`, `competency`, `leave`, `offboarding`, `onboarding`, `performance`, `personnel`, `recruitment`, `succession`, `talent`, dan `workforceplanning`.

## Hasil

Empat belas paket langsung lulus. Suksesi dan posisi menolak nama database awal karena tes secara sengaja meminta awalan `hrm_disposable_` dan `HRM_TEST_DATABASE_DISPOSABLE=1`; keduanya lulus setelah dijalankan ulang dengan database sekali pakai yang memenuhi syarat tersebut. Hasil akhir: **16/16 paket lulus**. Container HRM tetap `running healthy`.

## Batas klaim

Masing-masing paket membuktikan kasus yang tertulis pada tesnya, bukan seluruh 84 resource atau seluruh transisi bisnis. Sebagian tes memakai actor, policy, verifier owner, atau fixture SQL tiruan. Hasil ini tidak membuktikan grant IAM, E2E Gateway, verifikasi Foundation/Finance/DRM produksi, atau 38 validator mutasi yang belum terdaftar. Tidak ada skema atau kode bisnis yang diubah oleh audit ini.
