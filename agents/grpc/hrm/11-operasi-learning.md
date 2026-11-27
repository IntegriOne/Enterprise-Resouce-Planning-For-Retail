# List operasi

`human_resource.learning.v1.LearningService` memiliki 27 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/learning/v1/learning.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateLearningCourse` | Membuat entri `learning_courses`. | Mutation | `CreateLearningCourseRequest` → `LearningCourseResponse` | `human_resource.learning_courses.create` |
| `UpdateLearningCourse` | Mengubah entri `learning_courses` sesuai field mask. | Mutation | `UpdateLearningCourseRequest` → `LearningCourseResponse` | `human_resource.learning_courses.update` |
| `DeleteLearningCourse` | Menghapus entri `learning_courses` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.learning_courses.delete` |
| `GetLearningCourseByID` | Mengambil entri `learning_courses` berdasarkan ID. | Query | `IDRequest` → `LearningCourseResponse` | `human_resource.learning_courses.read` |
| `GetAllLearningCourses` | Mendaftar entri `learning_courses` dengan paginasi. | Query | `PageRequest` → `LearningCoursesResponse` | `human_resource.learning_courses.read` |
| `FindLearningCourseByCourseCode` | Menyaring daftar `learning_courses` berdasarkan `course_code`, dengan paginasi. | Query | `StringFinderRequest` → `LearningCoursesResponse` | `human_resource.learning_courses.read` |
| `FindLearningCourseByCourseType` | Menyaring daftar `learning_courses` berdasarkan `course_type`, dengan paginasi. | Query | `StringFinderRequest` → `LearningCoursesResponse` | `human_resource.learning_courses.read` |
| `FindLearningCourseByProviderPartyReferenceID` | Menyaring daftar `learning_courses` berdasarkan `provider_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningCoursesResponse` | `human_resource.learning_courses.read` |
| `StartLearningEnrollment` | Menandai penugasan sebagai mulai dikerjakan dengan waktu server. | Mutation | `IDRequest` → `LearningEnrollmentResponse` | `human_resource.learning_enrollments.start` |
| `CompleteLearningEnrollment` | Mencatat penyelesaian administratif enrollment yang sedang berjalan. | Mutation | `IDRequest` → `LearningEnrollmentResponse` | `human_resource.learning_enrollments.complete` |
| `WithdrawLearningEnrollment` | Mengakhiri penugasan terbuka atau berjalan dan mempertahankan riwayatnya. | Mutation | `IDRequest` → `LearningEnrollmentResponse` | `human_resource.learning_enrollments.withdraw` |
| `CreateLearningEnrollment` | Membuat entri `learning_enrollments`. | Mutation | `CreateLearningEnrollmentRequest` → `LearningEnrollmentResponse` | `human_resource.learning_enrollments.create` |
| `UpdateLearningEnrollment` | Mengubah entri `learning_enrollments` sesuai field mask. | Mutation | `UpdateLearningEnrollmentRequest` → `LearningEnrollmentResponse` | `human_resource.learning_enrollments.update` |
| `DeleteLearningEnrollment` | Menghapus entri `learning_enrollments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.learning_enrollments.delete` |
| `GetLearningEnrollmentByID` | Mengambil entri `learning_enrollments` berdasarkan ID. | Query | `IDRequest` → `LearningEnrollmentResponse` | `human_resource.learning_enrollments.read` |
| `GetAllLearningEnrollments` | Mendaftar entri `learning_enrollments` dengan paginasi. | Query | `PageRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByWorkerID` | Menyaring daftar `learning_enrollments` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByCourseID` | Menyaring daftar `learning_enrollments` berdasarkan `course_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByEnrollmentStatus` | Menyaring daftar `learning_enrollments` berdasarkan `enrollment_status`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByAssignedByPrincipalID` | Menyaring daftar `learning_enrollments` berdasarkan `assigned_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByCredentialReferenceID` | Menyaring daftar `learning_enrollments` berdasarkan `credential_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByFundingReferenceID` | Menyaring daftar `learning_enrollments` berdasarkan `funding_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `FindLearningEnrollmentByIdempotencyKey` | Menyaring daftar `learning_enrollments` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `LearningEnrollmentsResponse` | `human_resource.learning_enrollments.read` |
| `CreateLearningCourseCompetency` | Membuat entri `learning_course_competencies`. | Mutation | `CreateLearningCourseCompetencyRequest` → `LearningCourseCompetencyResponse` | `human_resource.learning_course_competencies.create` |
| `UpdateLearningCourseCompetency` | Mengubah entri `learning_course_competencies` sesuai field mask. | Mutation | `UpdateLearningCourseCompetencyRequest` → `LearningCourseCompetencyResponse` | `human_resource.learning_course_competencies.update` |
| `DeleteLearningCourseCompetency` | Menghapus entri `learning_course_competencies` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `LearningCourseCompetencyKey` → `EmptyResponse` | `human_resource.learning_course_competencies.delete` |
| `GetLearningCourseCompetencyByID` | Mengambil entri `learning_course_competencies` berdasarkan ID. | Query | `LearningCourseCompetencyKey` → `LearningCourseCompetencyResponse` | `human_resource.learning_course_competencies.read` |
| `GetAllLearningCourseCompetencies` | Mendaftar entri `learning_course_competencies` dengan paginasi. | Query | `PageRequest` → `LearningCourseCompetenciesResponse` | `human_resource.learning_course_competencies.read` |
| `FindLearningCourseCompetencyByCourseID` | Menyaring daftar `learning_course_competencies` berdasarkan `course_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningCourseCompetenciesResponse` | `human_resource.learning_course_competencies.read` |
| `FindLearningCourseCompetencyByCompetencyID` | Menyaring daftar `learning_course_competencies` berdasarkan `competency_id`, dengan paginasi. | Query | `StringFinderRequest` → `LearningCourseCompetenciesResponse` | `human_resource.learning_course_competencies.read` |

## Butuh GraphQL

18 kandidat Query dan 12 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

Ketiga command diperuntukkan bagi petugas manusia dengan grant eksplisit dan MFA terbaru. HRM mencatat waktu mulai dan selesai dari jam server. Penyelesaian administratif tidak menerbitkan nilai atau kredensial. Bukti kelulusan, nilai, atau sertifikat membutuhkan workflow tersendiri.

`StartLearningEnrollment` memeriksa masa berlaku kursus pada waktu mulai. Penugasan yang masih valid saat dibuat tetap ditolak jika tanggal mulai sudah melewati tanggal akhir kursus.

Kursus dapat dinonaktifkan melalui `UpdateLearningCourse` setelah seluruh enrollment mencapai status akhir dan tindakan pengembangan yang memakai kursus sudah selesai. Riwayat enrollment dan kaitan kompetensi tetap bisa dirujuk, sementara penugasan baru terhadap kursus nonaktif ditolak. Perubahan definisi atau penghapusan kursus yang masih dirujuk tetap ditolak.

`WithdrawLearningEnrollment` menerima status `planned`, `enrolled`, atau `in_progress`. HRM mempertahankan waktu mulai jika sudah ada dan mencatat perubahan status pada outbox. Enrollment yang sudah mencapai status akhir tidak dapat ditarik ulang.
