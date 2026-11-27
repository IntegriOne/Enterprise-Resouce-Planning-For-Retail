# List operasi

Inventaris 1015 RPC bisnis yang dideklarasikan pada 22 service privat HRM. Jumlah ini mengukur luas kontrak gRPC. Status penyelesaian workflow dan kesiapan Gateway dinilai melalui tes dan persetujuan IAM.
Lihat [status wiring Gateway](26-status-wiring-gateway-hrm.md) untuk field GraphQL yang sudah terhubung, izin dan bukti pengujian.

| Service | RPC | Kandidat Query | Kandidat Mutation | Nonaktif |
| --- | ---: | ---: | ---: | ---: |
| [AccessRequestService](2-operasi-accessrequest.md) | 15 | 12 | 3 | 0 |
| [AnalyticsService](3-operasi-analytics.md) | 16 | 10 | 6 | 0 |
| [BenefitsService](4-operasi-benefits.md) | 22 | 14 | 8 | 0 |
| [CareerService](5-operasi-career.md) | 56 | 37 | 19 | 0 |
| [CompensationService](6-operasi-compensation.md) | 26 | 19 | 7 | 0 |
| [CompetencyService](7-operasi-competency.md) | 36 | 24 | 12 | 0 |
| [EmployeeRelationsService](8-operasi-employeerelations.md) | 34 | 25 | 9 | 0 |
| [EngagementService](9-operasi-engagement.md) | 17 | 11 | 3 | 3 |
| [HRService](10-operasi-hrservice.md) | 28 | 18 | 10 | 0 |
| [LearningService](11-operasi-learning.md) | 30 | 18 | 12 | 0 |
| [LeaveService](12-operasi-leave.md) | 44 | 28 | 13 | 3 |
| [OffboardingService](13-operasi-offboarding.md) | 31 | 19 | 12 | 0 |
| [OnboardingService](14-operasi-onboarding.md) | 23 | 14 | 9 | 0 |
| [PayrollService](15-operasi-payroll.md) | 122 | 88 | 28 | 6 |
| [PerformanceService](16-operasi-performance.md) | 60 | 38 | 22 | 0 |
| [PersonnelService](17-operasi-personnel.md) | 80 | 54 | 23 | 3 |
| [PolicyService](18-operasi-policy.md) | 20 | 13 | 7 | 0 |
| [PromotionService](19-operasi-promotion.md) | 33 | 22 | 11 | 0 |
| [RecruitmentService](20-operasi-recruitment.md) | 145 | 95 | 47 | 3 |
| [SuccessionService](21-operasi-succession.md) | 19 | 12 | 7 | 0 |
| [TalentService](22-operasi-talent.md) | 102 | 64 | 38 | 0 |
| [WorkforcePlanningService](23-operasi-workforceplanning.md) | 57 | 37 | 20 | 0 |
| **Total** | **1016** | **672** | **326** | **18** |

## Butuh GraphQL

- Gunakan **Query** untuk detail berdasarkan ID, daftar, dan finder dengan filter yang tercantum. Teruskan paginasi ke HRM dan gunakan total dari responsnya agar scope berlaku pada hasil dan hitungan.
- Lewati RPC **Nonaktif** saat membuat resolver. Proto masih mendeklarasikannya, tetapi katalog HRM tidak menyediakan izin runtime untuk CRUD pada resource baca-saja.
- Gunakan **Mutation** untuk pembuatan, perubahan, penghapusan, serta perintah workflow. Petakan field mask dan bedakan nilai yang sengaja dikosongkan dari field yang tidak dikirim. Untuk perintah yang menerima idempotency key, gunakan kembali key yang sama saat mengulang permintaan yang sama.
- Seluruh akses HTTP aplikasi melalui `/api/graphql`. Mutation mengikuti pemeriksaan CSRF Gateway. Teruskan identitas dan metadata sesi tepercaya dari server melalui koneksi gRPC privat. Tentukan field GraphQL yang boleh terlihat berdasarkan kebutuhan tiap operasi, terutama untuk payroll, kesehatan, hubungan karyawan, dan dokumen personal.
- Klasifikasi pada tabel adalah masukan untuk desain schema. Pemilihan operasi yang boleh diterbitkan serta nama field GraphQL mengikuti kebutuhan produk dan katalog yang disetujui pemilik domain. Periksa perilaku CRUD terhadap aturan workflow sebelum menyediakan resolver.

## RBAC dan scope

- HRM memverifikasi sesi employee, revisi permission, proyeksi IAM, dan grant saat memproses RPC. Grant berasal dari IAM. Kode di inventaris adalah permission yang diperiksa HRM, bukan penetapan role, grup, atau hak pengguna. [Policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go)
- Scope `legal_entity` saat ini didukung untuk baca dan daftar pekerja, rencana tenaga kerja, dan baris rencana. HRM menerapkan filter sebelum paginasi. Jangan memakai filter bisnis yang dikirim klien sebagai bukti hak akses. [Implementasi scope](../../../backend/services/human_resource/infrastructure/security/policy.go)
- Sensitivitas mengikuti katalog IAM dan aturan minimum HRM. Operasi sensitif memerlukan MFA dan umur sesi paling lama sepuluh menit menurut policy HRM. Tampilkan kebutuhan reautentikasi sesuai respons sistem. [Katalog dan aturan step-up](../../../backend/services/human_resource/infrastructure/security/catalog.go)
- Aksi approve, reject, return, dan apply atas rekomendasi promosi memakai permission `promotion_decisions`. `RecordTalentCalibrationDecision` memakai `talent_calibration_sessions.approve`. `PostLedgerEntry` memakai `leave_entitlements.post`. Masing-masing kode memiliki prefiks `human_resource.`.

## Kesiapan implementasi

Sebagian RPC CRUD hanya berlaku pada draf atau ditolak oleh validator HRM. Jalur yang memerlukan bukti DRM atau settlement Finance dapat menolak permintaan ketika integrasi pemilik data belum tersedia. Cek laporan domain sebelum memasukkan operasi ke backlog Gateway. Workflow yang masih terbuka mencakup pemasangan verifier mandat panel suksesi dan kanal kandidat. [Konteks HRM](../../context/human_resource/1-context.md)

`cmd/main.go` memulai HRM tanpa memasang adapter opsional berikut. Pemeriksaan kesehatan gRPC dapat berhasil ketika operasi terkait masih mengembalikan `Unavailable`. [Entrypoint HRM](../../../backend/services/human_resource/cmd/main.go) · [Opsi registry](../../../backend/services/human_resource/internal/service/registry.go)

| Adapter runtime | Operasi Gateway yang terdampak |
| --- | --- |
| Verifier rekrutmen | Persetujuan dan pembukaan requisition, publikasi posting, verifikasi bukti asesmen, persetujuan dan penerbitan offer, serta pencatatan respons kandidat. [Rekrutmen](20-operasi-recruitment.md) |
| Verifier mandat panel | `ApproveSuccessionPlan`. [Suksesi](21-operasi-succession.md) |
| Verifier offboarding | `CompleteOffboardingTask` dan `CompleteOffboardingPlan`. [Offboarding](13-operasi-offboarding.md) |
| Kalkulator payroll | `CalculatePayrollRun`. [Payroll](15-operasi-payroll.md) |
| Verifier tanda tangan kontrak | `RecordSignedEmploymentContract` membutuhkan bukti versi dokumen dan tanda tangan para pihak dari DRM. [Personnel](17-operasi-personnel.md) |
| Verifier slip pribadi | `GetMyPayslip` dan `ListMyPayslips` membutuhkan fakta penerbitan Finance dan akses privat DRM. Consumer Finance dan adapter penyimpanan PDF dibutuhkan agar slip tersedia. [Payroll](15-operasi-payroll.md) |
| Verifier Workforce Planning | `ApproveWorkforcePlan`. [Workforce Planning](23-operasi-workforceplanning.md) |
| Verifier bukti onboarding | `CompleteOnboardingTask`. [Onboarding](14-operasi-onboarding.md) |
| Verifier bukti resolusi HR Service | `ResolveServiceRequest`. [HR Service](10-operasi-hrservice.md) |

Startup HRM sudah memasang verifier pengaitan akun. `LinkWorkerPrincipal` dipanggil secara privat dari provisioning `createEmployeeUser` IAM dengan izin `iam.users.write` dan bukti yang mengikat akun dan pekerja. [Alur provisioning](33-alur-pengajuan-akun-employee.md).

`RequestPayrollPosting` mencatat permintaan pada outbox HRM. Status `approved` tetap berlaku sampai bukti jurnal Finance diterima. Gateway harus menampilkan permintaan dan konfirmasi jurnal sebagai status yang berbeda. [Payroll](15-operasi-payroll.md)

Kebutuhan bukti dokumen pada operasi HRM dirinci dalam [catatan integrasi DRM](24-kebutuhan-integrasi-drm.md). Kontrak jurnal, penerbitan, dan pembayaran dirinci dalam [catatan integrasi Finance](25-kebutuhan-integrasi-finance-payroll.md).

## Verifikasi sebelum wiring

- Tes integrasi gRPC pada 22 domain lulus dengan PostgreSQL sementara dan race detector. [Laporan audit](../../audit/human_resource/170-2026-11-25-054958-verifikasi-integrasi-grpc-seluruh-domain.md)
- Regresi terbaru juga lulus pada 22 domain, bersama unit test modul dan go vet. [Audit regresi](../../audit/human_resource/172-2026-11-25-061323-regresi-22-domain-hrm.md)

- Pencocokan ulang proto dengan tabel handoff menemukan 1015 RPC pada 22 service tanpa nama yang hilang atau duplikat, terdiri dari 672 kandidat Query, 325 kandidat Mutation, dan 18 RPC nonaktif. Pemeriksaan ini membandingkan nama RPC, tipe request dan response, klasifikasi Query atau Mutation, serta kode permission berdasarkan handler, service, dan katalog akses HRM.
- Jalankan `python3 backend/services/human_resource/tools/check_gateway_handoff.py` dari root repositori untuk memeriksa kembali tabel. Pemetaan berbasis pola source Go dan mencatat pengecualian workflow secara eksplisit.
- Tes policy runtime mencakup ketiga mutasi CRUD pada enam resource baca-saja. HRM menolak semua 18 operasi meski snapshot fixture memuat grant global yang cocok dan sesi sudah MFA.
- `TestRegisterAllMatchesHRMProtoServices` lulus dari checkout lokal dengan memeriksa pendaftaran setiap service, nama method, dan tipe streaming terhadap descriptor protobuf. Health check gRPC kontainer HRM yang sedang berjalan juga lulus. Tes pendaftaran memakai registry offline, sedangkan health check hanya memeriksa kesehatan kontainer. [Tes wiring](../../../backend/services/human_resource/cmd/wiring_test.go)
- Tes policy dan interceptor HRM lulus dengan race detector. Tes unit mencakup sesi dicabut, revisi izin berubah, MFA, scope badan hukum, dan error aman. Aktivasi grant di IAM dan wiring resolver HRM diperlukan untuk pengujian melalui Gateway. [Katalog HRM untuk IAM](../../roles-permissions/human_resource/1-roles-permissions.md)
- Tes Personnel memakai PostgreSQL dan Redis TLS sementara untuk membuktikan get/list ditolak setelah pencabutan token. Sesi dan grant berasal dari fixture. [Laporan scope pekerja](../../e2e/human_resource/8-personnel-query-worker-legal-entity-scope.md)
- Transport kontainer HRM lokal menerima health check bersertifikat dan menolak koneksi tanpa sertifikat klien. [Audit mTLS](../../audit/human_resource/171-2026-11-25-061227-verifikasi-mtls-grpc-hrm.md)
- Database development masih memiliki checksum baseline lama. Fresh install terbaru lulus pada database sementara. [Audit checksum](../../audit/human_resource/173-2026-11-25-061729-checksum-baseline-runtime-lokal.md)
- Uji scope badan hukum, sesi dicabut atau revisi berubah, ketiadaan permission, MFA kedaluwarsa, CSRF, kegagalan dependency dan transisi workflow yang ditolak melalui `/api/graphql`.
- Petakan status gRPC ke error GraphQL sesuai standar Gateway. Pertahankan pembedaan input tidak valid, sesi tidak sah, akses ditolak, konflik, prasyarat workflow gagal, dan dependency tidak tersedia.
- Inventaris ini mencakup RPC bisnis dari kontrak proto. Health check gRPC dipakai untuk operasional privat. [Bootstrap gRPC HRM](../../../backend/services/human_resource/transport/grpc/server.go)

## Verifikasi kesesuaian inventaris

`python3 backend/services/human_resource/tools/check_gateway_handoff.py` memeriksa proto, handler, klasifikasi dan izin terhadap tabel handoff. Hasil pada checkout aktif mencatat 22 service dan 1.015 RPC tanpa selisih. `TestRegisterAllMatchesHRMProtoServices` memeriksa bahwa registrasi server mencakup setiap metode pada descriptor proto dengan flag streaming yang sama. Verifikasi ini mengukur cakupan kontrak. Bukti perilaku tersedia pada [regresi gRPC](../../e2e/human_resource/54-query-mutation-regresi-grpc-postgresql.md) dan [matriks Gateway](26-status-wiring-gateway-hrm.md).
