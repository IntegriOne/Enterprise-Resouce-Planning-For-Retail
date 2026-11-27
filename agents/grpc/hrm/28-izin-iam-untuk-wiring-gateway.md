# List operasi

Matriks berikut mencatat izin operasi Gateway HRM melalui pemetaan gRPC. Manifest resmi kini menyertakan 63 izin HRM, termasuk `access_requests.create` dan `access_requests.read` untuk provisioning employee. Tes otomatis memeriksa pemetaan dan sensitivitas untuk setiap izin resmi.

Lima izin menunggu dependency operasi positif. `employment_contracts.record_signature` memerlukan DRM. `offboarding_plans.complete` dan `offboarding_tasks.complete` memerlukan bukti pemilik. `payroll_runs.calculate` memerlukan kalkulator serta artefak. `workforce_plans.approve` memerlukan verifikasi anggaran Finance. Penautan principal memakai `iam.users.write` pada jalur IAM privat, sehingga kode historis `human_resource.workers.link_principal` tidak didaftarkan.

## RBAC

Kode di bawah memakai prefiks `human_resource.`. Semua operasi menggunakan sesi employee. Kolom MFA menunjukkan batas minimum runtime. Grant yang ditandai sensitif oleh IAM tetap mengikuti aturan autentikasi yang berlaku.

| Cakupan | MFA minimum | Izin |
| --- | --- | --- |
| global atau badan hukum | Sesuai katalog IAM | `workers.read`, `workforce_plans.read`, `workforce_plan_lines.read` |
| global | Sesuai katalog IAM | `workers.create`, `workers.update`, `access_requests.create`, `access_requests.read` |
| global | Sesuai katalog IAM | `offboarding_plans.read`, `offboarding_plans.create` |
| global | Sesuai katalog IAM | `offboarding_tasks.create`, `offboarding_tasks.read` |
| global | Wajib | `offboarding_tasks.complete`, `offboarding_tasks.waive` |
| global | Wajib | `offboarding_plans.approve`, `offboarding_plans.start`, `offboarding_plans.complete`, `offboarding_plans.cancel` |
| global | Wajib | `workers.activate`, `workers.terminate`, `workers.apply_primary_assignment` |
| self milik principal sesi | Sesuai katalog IAM | `payslips.read_self` |
| global | Sesuai katalog IAM | `payroll_groups.read`, `payroll_periods.read`, `payroll_rule_sets.read`, `payroll_runs.read` |
| global | Sesuai katalog IAM | `jobs.read`, `job_families.read`, `positions.read` |
| global | Wajib | `positions.activate`, `positions.create`, `positions.update`, `positions.delete` |
| global | Wajib | `employment_contracts.read`, `employment_contracts.create`, `employment_contracts.record_signature`, `employment_contracts.delete`, `employment_contracts.update` |
| global | Wajib | `worker_assignments.approve`, `worker_assignments.transfer` |
| global | Sesuai katalog IAM | `worker_assignments.read`, `worker_assignments.create` |
| global | Wajib | `payroll_results.read`, `payroll_result_lines.read` |
| global | Wajib | `payroll_runs.create`, `payroll_runs.calculate`, `payroll_runs.submit_review`, `payroll_runs.approve`, `payroll_runs.cancel`, `payroll_runs.request_posting` |
| global | Sesuai katalog IAM | `leave_policies.read`, `leave_entitlements.read`, `leave_ledger_entries.read`, `leave_adjustments.read`, `leave_adjustments.create` |
| global | Wajib | `leave_adjustments.approve`, `leave_adjustments.reject`, `leave_entitlements.post`, `leave_entitlements.close` |
| global | Sesuai katalog IAM | `workforce_plans.create`, `workforce_plans.update`, `workforce_plan_lines.create`, `workforce_plan_lines.update` |
| global | Wajib | `workforce_plans.submit`, `workforce_plans.cancel`, `workforce_plans.approve`, `workforce_plans.supersede`, `workforce_plans.delete`, `workforce_plan_lines.delete` |

Untuk uji awal akses pekerja, gunakan akun uji dengan satu grant baca pekerja di badan hukum A dan fixture pekerja A serta B. Untuk mutasi payroll dan koreksi cuti, pengusul dan pengambil keputusan memakai akun berbeda sesuai aturan HRM. Grant slip mandiri memerlukan verifikasi scope self dan kepemilikan versi DRM.

## Butuh GraphQL Query dan Mutation

Nama field, RPC, scope dan bukti uji tercatat dalam [matriks wiring Gateway](26-status-wiring-gateway-hrm.md). Aktivasi memerlukan manifest resmi yang disetujui dan assignment lewat IAM. [Handoff IAM](../../notes/iam_management/1-human_resource-permission-handoff.md) · [Tes HTTPS pekerja](../../e2e/gateway/34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md).

## Aktivasi tahap awal untuk uji pekerja

Aktivasi awal memakai `human_resource.workers.read` dan satu role uji baca pekerja. Gunakan `assignUserRole` untuk mengikat role ke akun employee pada scope `legal_entity` badan hukum A. Verifier IAM perlu menerima badan hukum A sebagai target yang aktif di Enterprise Foundation. Akun uji tidak diberi grant global atau grant pekerja dari badan hukum B.

Siapkan satu pekerja di masing-masing badan hukum dan jalankan tes HTTPS dengan enam input fixture yang tercatat. Kriteria lulus mencakup detail pekerja A, `NOT_FOUND` untuk pekerja B, dan hasil pagination yang tidak memuat baris di luar scope. Skenario ini lulus dengan role uji pada akun superadmin berjenis employee. Assignment global dicabut sebelum uji scope. [Skenario HTTPS](../../e2e/gateway/34-hrm-query-pekerja-dengan-grant-badan-hukum-https.md) · [Mutasi assignment IAM](../../../backend/gateway/graphql/iam/iam_rbac.graphql).

Runtime development memasang verifier Enterprise Foundation pada IAM dan HRM dengan target `enterprise-foundation:9002`. Uji paket verifier IAM lulus. Uji Gateway untuk `createEmployeeUser` juga lulus, termasuk penolakan saat `worker_id` diberikan tanpa provisioner. Untuk skenario baca pekerja berscope badan hukum, akun operator uji dan dua pekerja adalah fixture terpisah. Pembuatan akun employee mewajibkan `worker_id` dan `provisioning_request_id`. Scope tetap berasal dari assignment resmi. Uji akses berscope badan hukum dan uji penautan akun memiliki fixture tersendiri. [Kontrak pembuatan akun](../../../backend/gateway/graphql/iam/iam_user.graphql) · [Handoff bukti penautan](../../notes/iam_management/1-human_resource-permission-handoff.md).

Saat menyusun role untuk `postLeaveAdjustment`, sertakan `leave_entitlements.post` dengan MFA serta `leave_ledger_entries.read` dan `leave_adjustments.read`. HRM membaca ledger dan koreksi yang disetujui pada transaksi posting sebelum memperbarui saldo. Uji HTTPS tanpa izin baca koreksi mengembalikan `FORBIDDEN`. [Bukti posting](../../e2e/gateway/110-hrm-query-mutation-koreksi-cuti-dan-ledger-mfa.md).

Untuk izin slip, kirim `scope_type: SELF` dan `scope_id` yang cocok dengan akun target saat memanggil `assignUserRole`. IAM menolak assignment tanpa ID scope, dan Gateway menolak grant global untuk `myPayslip`. [Uji scope self](../../e2e/gateway/117-hrm-query-slip-gaji-scope-self-https.md).

## Katalog resmi cuti

Manifest HRM kini memuat sembilan izin cuti untuk pembacaan kebijakan, hak dan ledger serta pengajuan, pembacaan dan keputusan koreksi. Untuk posting, kode IAM `human_resource.leave_entitlements.post` dipetakan ke aksi gRPC `hr_leave_entitlements.post_leave_entry`. Tes memeriksa kode dan sensitivitas keduanya. Penugasan tetap melalui GraphQL.

## Bukti IAM untuk clearance offboarding

`completeOffboardingPlan` membutuhkan bukti bahwa IAM sudah menerapkan keputusan akses yang terkait dengan pekerja, rencana offboarding, dan tanggal efektif. HRM membedakan principal pekerja dari principal operator. Jika data HRM tidak memuat tautan principal, verifier perlu membedakan kondisi tersebut dari kegagalan membaca data IAM. Verifikasi dilakukan antarlayanan. Gateway tetap memakai mutation yang ada dan meneruskan referensi clearance Finance secara terpisah. [Kontrak HRM](13-operasi-offboarding.md#handoff-verifier-iam).

Owner memilih penutupan akses portal setelah terminasi. IAM perlu mencabut sesi lama dan menolak login berikutnya. Mantan karyawan meminta slip kepada HR untuk dicetak melalui operasi khusus. Uji penerimaan perlu membuktikan bahwa sesi lama dan login baru ditolak, serta bahwa verifier menolak bukti clearance yang merujuk pekerja atau rencana lain. Gangguan IAM harus mempertahankan status rencana dan outbox. [Handoff cetak ulang](39-cetak-ulang-slip-gaji-mantan-karyawan.md).

Perilaku yang sudah ada di HRM mencakup riwayat hubungan kerja. Jika IAM memberi sesi employee dengan izin SELF, HRM mencari seluruh record pekerja yang tertaut ke principal dan memastikan seluruhnya memiliki identitas orang yang sama. Status `terminated` pada satu record tidak menghentikan pembacaan slip riwayat bila principal juga memiliki hubungan kerja yang masih berjalan. Jika seluruh hubungan kerja sudah berakhir, HRM menolak daftar, detail, dan PDF mandiri meskipun sesi memiliki grant SELF. Tes PostgreSQL mencakup riwayat yang sah dan penolakan ketika dua record tertaut mengacu pada orang yang berbeda. Uji ini belum membuktikan kebijakan login setelah seluruh hubungan kerja berakhir. [Regresi payroll](../../../backend/services/human_resource/transport/grpc/handler/payroll/test/integration_test.go).
