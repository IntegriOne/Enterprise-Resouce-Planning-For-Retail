# List operasi

`human_resource.offboarding.v1.OffboardingService` memiliki 31 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/offboarding/v1/offboarding.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CompleteOffboardingPlan` | Menyelesaikan entri `offboarding_plans` melalui transisi workflow HRM. | Mutation | `CompleteOffboardingPlanRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.complete` |
| `CancelOffboardingPlan` | Membatalkan rencana yang sudah disetujui dan belum dimulai, dengan alasan dan riwayat audit. | Mutation | `CancelOffboardingPlanRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.cancel` |
| `StartOffboardingPlan` | Memulai entri `offboarding_plans` melalui transisi workflow HRM. | Mutation | `IDRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.start` |
| `ApproveOffboardingPlan` | Menyetujui rencana draf yang sudah memiliki minimal satu tugas melalui persetujuan independen. | Mutation | `IDRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.approve` |
| `CreateOffboardingPlan` | Membuat entri `offboarding_plans`. | Mutation | `CreateOffboardingPlanRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.create` |
| `UpdateOffboardingPlan` | Mengubah entri `offboarding_plans` sesuai field mask. | Mutation | `UpdateOffboardingPlanRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.update` |
| `DeleteOffboardingPlan` | Menghapus entri `offboarding_plans` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.offboarding_plans.delete` |
| `GetOffboardingPlanByID` | Mengambil entri `offboarding_plans` berdasarkan ID. | Query | `IDRequest` → `OffboardingPlanResponse` | `human_resource.offboarding_plans.read` |
| `GetAllOffboardingPlans` | Mendaftar entri `offboarding_plans` dengan paginasi. | Query | `PageRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByWorkerID` | Menyaring daftar `offboarding_plans` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByPlanNumber` | Menyaring daftar `offboarding_plans` berdasarkan `plan_number`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByReasonCode` | Menyaring daftar `offboarding_plans` berdasarkan `reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByStatus` | Menyaring daftar `offboarding_plans` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByExitInterviewDocumentVersionReferenceID` | Menyaring daftar `offboarding_plans` berdasarkan `exit_interview_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByFinalPayrollRunID` | Menyaring daftar `offboarding_plans` berdasarkan `final_payroll_run_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByFinanceClearanceReferenceID` | Menyaring daftar `offboarding_plans` berdasarkan `finance_clearance_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByApprovedByPrincipalID` | Menyaring daftar `offboarding_plans` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `FindOffboardingPlanByCreatedByPrincipalID` | Menyaring daftar `offboarding_plans` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingPlansResponse` | `human_resource.offboarding_plans.read` |
| `CompleteOffboardingTask` | Menyelesaikan entri `offboarding_tasks` melalui transisi workflow HRM. | Mutation | `CompleteOffboardingTaskRequest` → `OffboardingTaskResponse` | `human_resource.offboarding_tasks.complete` |
| `WaiveOffboardingTask` | Mengecualikan entri `offboarding_tasks` melalui transisi workflow HRM. | Mutation | `WaiveOffboardingTaskRequest` → `OffboardingTaskResponse` | `human_resource.offboarding_tasks.waive` |
| `CreateOffboardingTask` | Membuat entri `offboarding_tasks`. | Mutation | `CreateOffboardingTaskRequest` → `OffboardingTaskResponse` | `human_resource.offboarding_tasks.create` |
| `UpdateOffboardingTask` | Mengubah entri `offboarding_tasks` sesuai field mask. | Mutation | `UpdateOffboardingTaskRequest` → `OffboardingTaskResponse` | `human_resource.offboarding_tasks.update` |
| `DeleteOffboardingTask` | Menghapus entri `offboarding_tasks` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.offboarding_tasks.delete` |
| `GetOffboardingTaskByID` | Mengambil entri `offboarding_tasks` berdasarkan ID. | Query | `IDRequest` → `OffboardingTaskResponse` | `human_resource.offboarding_tasks.read` |
| `GetAllOffboardingTasks` | Mendaftar entri `offboarding_tasks` dengan paginasi. | Query | `PageRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |
| `FindOffboardingTaskByOffboardingPlanID` | Menyaring daftar `offboarding_tasks` berdasarkan `offboarding_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |
| `FindOffboardingTaskByTaskCode` | Menyaring daftar `offboarding_tasks` berdasarkan `task_code`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |
| `FindOffboardingTaskByAssignedPrincipalID` | Menyaring daftar `offboarding_tasks` berdasarkan `assigned_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |
| `FindOffboardingTaskByStatus` | Menyaring daftar `offboarding_tasks` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |
| `FindOffboardingTaskByOwnerConfirmationReferenceID` | Menyaring daftar `offboarding_tasks` berdasarkan `owner_confirmation_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |
| `FindOffboardingTaskByWaivedByPrincipalID` | Menyaring daftar `offboarding_tasks` berdasarkan `waived_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffboardingTasksResponse` | `human_resource.offboarding_tasks.read` |

## Butuh GraphQL

19 kandidat Query dan 12 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

## Prasyarat penyelesaian

- `CompleteOffboardingTask` memerlukan tugas pending milik HRM dan verifier yang mengikat bukti pada tugas, rencana, dan principal. Tugas milik management lain memerlukan jalur konfirmasi dari pemiliknya.
- `CompleteOffboardingPlan` memerlukan rencana berjalan, pekerja berstatus `terminated` dengan tanggal yang cocok dan sudah tercapai, run terminasi yang sudah disetujui dan memuat hasil pekerja, tugas wajib yang sudah diselesaikan atau dikecualikan secara sah, serta verifikasi clearance Finance dan IAM.
- Registry HRM menerima kedua adapter melalui `WithOffboardingVerifiers`. Startup standar belum memasangnya, sehingga operasi terkait dapat mengembalikan `Unavailable`. [Registry HRM](../../../backend/services/human_resource/internal/service/registry.go)

Gunakan RPC `TerminateWorker` pada Personnel untuk mencatat status, tanggal, dan alasan sebelum menutup rencana. [Operasi terminasi](17-operasi-personnel.md#terminasi-pekerja)

Tugas pending milik HRM tetap dapat diselesaikan atau dikecualikan saat rencana berjalan setelah terminasi tercatat. Untuk pekerja berstatus `terminated`, tanggal terminasi harus cocok dengan rencana.

Hasil payroll terminasi tidak boleh menunjukkan tahap posting atau pembayaran yang lebih maju daripada status run.

Verifier clearance menerima ID hasil pekerja, badan hukum, mata uang, serta digest SHA-256 run dan kalkulasi hasil pekerja. Adapter perlu mencocokkan semua referensi dengan bukti clearance dari pemilik.

Saat rencana dibuat, diubah, disetujui, atau dimulai, hari kerja terakhir harus sama dengan atau setelah tanggal mulai bekerja jika tanggal tersebut tercatat. Draf lama yang melanggar aturan tanggal ini tetap bisa dihapus jika belum memiliki tugas.

`CompleteOffboardingPlan` juga memeriksa kesamaan badan hukum pemberi kerja pada run payroll final dan data pekerja. Perbedaan menghasilkan `FailedPrecondition` sebelum verifier clearance dipanggil.

Jika data pekerja sudah memiliki kode alasan terminasi, kode tersebut harus cocok dengan kode pada rencana yang disetujui. Data lama tanpa kode tetap mengikuti pemeriksaan status dan tanggal terminasi.

`CancelOffboardingPlan` memerlukan MFA, alasan, dan pelaku selain pekerja yang menjadi subjek rencana. Pembatalan menjaga metadata persetujuan serta daftar tugas, dan membuka slot untuk rencana pengganti. Rencana yang sudah berjalan memerlukan workflow koreksi tersendiri.

Hasil payroll individual berstatus `paid` dapat dipakai saat run terminasi masih `posted` jika referensi pembayaran Finance valid. HRM meneruskan referensi tersebut ke verifier clearance untuk dicocokkan dengan bukti pembayaran Finance sebelum rencana ditutup.

## Sisa wiring Gateway

Dua belas dari 31 RPC sudah dipakai Gateway, termasuk Query detail tugas. Sembilan belas RPC berikut masih perlu keputusan pemetaan. [Matriks operasi yang terhubung](26-status-wiring-gateway-hrm.md).

| RPC | Jumlah | Rencana Gateway dan pemeriksaan akses |
| --- | ---: | --- |
| `GetAllOffboardingPlans`, `GetAllOffboardingTasks` | 2 | Daftar lintas pekerja atau rencana memerlukan keputusan produk tentang visibilitas. Daftar saat ini mewajibkan ID pekerja atau rencana. |
| Finder rencana untuk nomor, alasan dan status | 3 | Evaluasi sebagai filter daftar operator HRM. Kontrak finder mendukung satu kriteria per panggilan. Perlu aturan ketika beberapa filter dikirim bersamaan. |
| Finder rencana untuk versi wawancara keluar, payroll final, clearance Finance, approver, dan pembuat | 5 | Tentukan kebutuhan pengguna dan scope tiap referensi sebelum publikasi. Referensi DRM dan Finance tetap tunduk pada aturan pemilik. |
| Finder tugas untuk kode, principal penerima, status, bukti konfirmasi, dan principal yang mengecualikan | 5 | Tentukan filter yang dibutuhkan pada daftar tugas. Izin daftar global yang ada tidak memberi akses mandiri ke tugas berdasarkan principal. |
| `UpdateOffboardingPlan`, `DeleteOffboardingPlan`, `UpdateOffboardingTask`, `DeleteOffboardingTask` | 4 | Petakan grant `update` dan `delete` per resource bersama IAM. Batasi input koreksi ke kolom bisnis yang dapat diedit, pertahankan perbedaan field absen dan null dalam field mask, dan uji penolakan HRM untuk status terkunci. Perubahan status tetap melalui mutation workflow yang sudah tersedia. |

Operasi daftar global, finder tambahan, dan koreksi memerlukan pemetaan paparan atau grant sebelum rollout.

## Handoff verifier IAM

Pada clearance, `ActorPrincipalID` mengidentifikasi operator yang menutup rencana. `WorkerPrincipalID` berasal dari relasi IAM pada pekerja yang dikunci di transaksi HRM. Nilai kosong berarti pekerja tidak memiliki relasi IAM yang tercatat. Adapter perlu membedakan kondisi tersebut dari kegagalan membaca bukti. GraphQL tetap menerima ID rencana, payroll final, dan clearance Finance.

Kebutuhan kontrak IAM adalah bukti yang mengikat akun pekerja, rencana offboarding, tanggal efektif dan keputusan akses. Kebijakan IAM perlu menetapkan izin kerja yang dicabut, perlakuan sesi, serta jalur akses slip gaji mandiri setelah hubungan kerja berakhir. HRM menyerahkan tindakan terhadap akun kepada IAM. Verifier produksi clearance masih memerlukan bukti IAM dan Finance.

Kebutuhan verifier Finance untuk penutupan rencana dicatat di [handoff clearance Finance](38-kebutuhan-finance-clearance-offboarding.md).

Jalur positif gRPC untuk konfirmasi tugas dan clearance rencana sudah diuji dengan verifier fixture pada PostgreSQL disposable. Bukti runtime produksi tetap terbuka. [Regresi offboarding](../../e2e/human_resource/58-offboarding-mutation-grpc-clearance-postgresql.md).

Deaktivasi akun setelah terminasi sudah diuji melalui HTTPS dan mTLS. Jalur clearance penutupan rencana masih memerlukan kontrak pemilik. [Operasi deaktivasi otomatis](40-deaktivasi-akun-setelah-terminasi.md)

### Sumber konfirmasi penyelesaian tugas

Kontrak `ConfirmationVerifier` membutuhkan bukti yang dapat dibaca berdasarkan `owner_confirmation_reference_id`. Bukti harus mengikat ID tugas, ID rencana, management pemilik, dan principal pelaku. Tabel tugas HRM menyimpan referensi setelah verifikasi berhasil. Kontrak data saat ini belum memiliki mekanisme penerbitan bukti tersebut. Runtime memerlukan penyedia bukti sebelum command penyelesaian bisa diaktifkan. ID tugas, perubahan status manual, dan UUID yang dikirim operator sendiri tidak membuktikan adanya konfirmasi pemilik.

`CompleteOffboardingTask` saat ini hanya menerima tugas pending milik HRM pada rencana yang berjalan. Jalur tugas milik management lain masih memerlukan mekanisme konfirmasi dari masing-masing pemilik. Jalur `WaiveOffboardingTask` menyimpan alasan pengecualian dan identitas operator sesuai kebijakan yang sudah tersedia. [Kontrak dan pemeriksaan](../../../backend/services/human_resource/internal/service/offboarding/complete_task.go)

Event `hr_offboarding_tasks.complete` mencatat pelaku pada envelope, serta ID rencana, management pemilik, dan ID konfirmasi terverifikasi pada `details`. Event ditulis dalam transaksi yang sama dengan status tugas. Ini memberi jejak audit penyelesaian. Penerbit bukti HRM tetap diperlukan sebelum verifikasi dapat diaktifkan di runtime standar.
