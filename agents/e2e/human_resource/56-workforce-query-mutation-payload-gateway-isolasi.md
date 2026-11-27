# Uji HTTPS payload workforce plan pada Gateway hasil isolasi

Gateway dibangun dari salinan `origin/main` dengan wiring HRM dan SDL payload baru. Uji berjalan melalui HTTPS dengan sertifikat development terverifikasi pada port 18443. Fixture badan hukum dibuat melalui Gateway Enterprise Foundation sementara pada port 18444. Kedua service bisnis memakai database PostgreSQL `*_test` baru. IAM dan Redis development dipakai untuk sesi dan akses sementara.

Login sebagai `superadmin`, buat role uji dan pasang permission `human_resource.workforce_plans.create`, `read`, dan `update` lewat GraphQL IAM, lalu login ulang. Mutation membawa token CSRF yang sah.

| Kasus | Hasil |
| --- | --- |
| Mutation `hrmCreateWorkforcePlan` | PASS. `was_successful: true` dan payload berisi draf dengan kode yang dikirim. |
| Mutation `hrmCreateWorkforcePlan` dengan kode dan versi yang sama | PASS. `was_successful: false`, error `CONFLICT`, dan objek hasil kosong. |
| Mutation `hrmUpdateWorkforcePlan` | PASS. Payload berisi ID draf yang sama dan `scenario_code: E2E-UPDATED`. |
| Query `hrmWorkforcePlan` setelah update | PASS. ID dan scenario sama dengan hasil Mutation. |
| Baca PostgreSQL HRM test | PASS. Satu baris `draft\|E2E-UPDATED`; percobaan duplikat tidak menambah baris. |
| Cabut penugasan role dan ketiga permission, hapus role, lalu logout | PASS melalui GraphQL IAM. |

Runner memakai `python3 /tmp/hrm-main-isolation.gcRxPH/run-positive.py` pada build lokal sebelum publikasi. Percobaan dengan namespace Redis arbitrer memberi `UNAVAILABLE` dari HRM. Uji lulus setelah kembali ke namespace development yang dikonfigurasi. Container dan kedua database sementara dibersihkan setelah uji. Cakupan laporan ini terbatas pada alur workforce plan, tanpa matriks lintas scope maupun konkurensi.
