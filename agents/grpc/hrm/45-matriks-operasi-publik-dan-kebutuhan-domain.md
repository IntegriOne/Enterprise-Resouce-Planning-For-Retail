# List operasi HRM untuk remediasi November 2026

**Cakupan terukur.** Pada checkout yang diaudit terdapat **176 query dan 326 mutation** GraphQL domain. Seluruh 502 root mempunyai resolver dan method adapter HRM. Inventaris privat domain memuat 1.018 RPC pada 22 service. Empat RPC workload untuk payroll Finance dan akses IAM berada di dua kontrak terpisah. Angka 674 pada inventaris domain adalah jumlah RPC yang diklasifikasikan sebagai baca. Satu field connection publik dapat memilih beberapa finder berdasarkan filter. [Audit HRL-01 sampai HRL-12](../../../agents/audit/human_resource/199-2026-11-29-000433-audit-logika-operasi-dan-otorisasi-hrm.md).

## Penjelasan operasi yang memerlukan perubahan kontrak

| Kebutuhan | Butuh GraphQL | Perubahan yang diminta | RBAC dan hasil untuk UI |
| --- | --- | --- | --- |
| Cetak ulang oleh HR | Query usulan `employeePayslipPeriods` + mutation tersedia `reissueEmployeePayslip` | Lookup paginasi periode terbit dari worker_id, lalu kirim worker_id dan result ID hasil pilihan petugas ke mutation cetak | HR berscope employer + MFA. Tampilkan periode, tanggal publikasi dan status. Batasi metadata sesuai kebutuhan cetak. HRL-05/10. |
| Profil dan kontrak sendiri | Query usulan `myEmploymentProfile` dan `myPublishedContracts` | Profil minimal dari principal sesi. Berikan versi kontrak yang sudah dirilis dengan pemeriksaan baca DRM | Pegawai aktif dengan permission self yang tepat. Hindari field internal HR dan ID teknis sebagai label. HRL-10. |
| Cuti sendiri | Query usulan `myLeaveBalances` | Entitlement aktif, satuan, periode dan saldo, hasil posting usage milik Workforce | Pegawai membaca miliknya dan tidak memasukkan worker pihak lain. HRL-10/12. |
| Hire dari penawaran | Mutation usulan `hireAcceptedApplicant` | Hubungkan offer accepted dan lamaran ke satu relasi worker dengan bukti dan idempotency | HR berwenang atas requisition dan entitas tujuan. Kembalikan nomor pekerja dan status onboarding. HRL-09. |
| Gaji yang disetujui dan diterbitkan | Pertahankan mutation `approveCompensationRecord`, `freezePayrollInput`, `createPayrollRun`, `calculatePayrollRun`, `approvePayrollRun`, `requestPayrollPosting` | Ubah seleksi, klaim kewajiban, engine, roster, digest dan transaksi approval pada jalur sekarang | Maker/checker terpisah. HRM mengajukan rincian. Finance sendiri menerbitkan dan membayar. HRL-01–04. |
| Slip pribadi | Pertahankan query `myPayslips`, `myPayslip`, `myPayslipPdf` | Dukung bukti posting dengan jurnal atau tanpa efek akuntansi. Sinkronkan publikasi Finance, PDF DRM dan status pembayaran per hasil | Status published, belum bayar, sebagian, selesai tanpa transfer dan lunas perlu dapat dibedakan sesuai bukti. Mantan pekerja melalui HR. HRL-05/06. |
| Kerahasiaan kasus | Pertahankan query `employeeRelationsCase`, `employeeRelationsCases`, pembacaan event dan `node` bersama | Terapkan ACL dan redaksi sebelum baca/count. Turunkan scope dari penugasan server | Investigator atau oversight berwenang melihat metadata sesuai kelas. HRL-07. |
| Bukti tindakan HR | Perbaiki mutation yang tercantum pada HRL-08 | Pasang verifier yang digunakan setiap cabang wajib. Sertakan owner penerbit bukti, status/verifikasi dan tanggal keputusan | Izin tindakan dan SoD per-domain. Berikan pesan yang menjelaskan tahap gagal tanpa isi dokumen. |
| Perubahan scope kerja | Gunakan workflow pengajuan yang ada. Tambahkan kontrak workload privat change/revoke + hasil efektif | Outbox dan status yang direkonsiliasi menurut intent dan versi hubungan | HR mengajukan, IAM mengelola grant. HRL-11. |
| Pembayaran dan debit cuti | Kontrak workload privat | Proyeksi bukti per result dari Finance dan posting usage/cancellation dari Workforce ke ledger | mTLS per owner, replay aman, default tertutup. HRL-06/12. |

Total usulan publik eksplisit pada tabel adalah **4 query baru dan 1 mutation baru**. Angka itu belum termasuk dalam baseline 502. Gunakan nama akhir sesuai konvensi SDL repo. Dokumentasikan setiap delta tambahan jika terbukti diperlukan. `node` umum berada di luar hitungan root domain HRM.

## Daftar RPC domain dan klasifikasi

Tautan per baris membuka daftar rinci RPC, permission dan action pada katalog existing. Kolom nonaktif menunjukkan RPC yang diklasifikasikan nonaktif oleh pemeriksa handoff. [Pemeriksa katalog](../../../backend/services/human_resource/tools/check_gateway_handoff.py).

| Service | RPC | Diklasifikasikan query | Diklasifikasikan mutation | Nonaktif |
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
| [RecruitmentService](20-operasi-recruitment.md) | 146 | 96 | 47 | 3 |
| [SuccessionService](21-operasi-succession.md) | 19 | 12 | 7 | 0 |
| [TalentService](22-operasi-talent.md) | 102 | 64 | 38 | 0 |
| [WorkforcePlanningService](23-operasi-workforceplanning.md) | 58 | 38 | 20 | 0 |
| **Total** | **1.018** | **674** | **326** | **18** |

RPC workload di luar tabel adalah `GetApprovedPayrollSnapshot`, `DeliverPayrollEvent`, `VerifyTerminationAccess` dan `VerifyWorkerAccountLink`. [Kontrak payroll privat](../../../backend/services/human_resource/contract/proto/human_resource/workload/payroll/v1/payroll_event.proto) · [Kontrak IAM privat](../../../backend/services/human_resource/contract/proto/human_resource/workload/termination/v1/termination_access.proto).

## Daftar aktual root GraphQL menurut SDL

Nama berikut diambil dari AST pada file sumber domain. Gunakan SDL untuk tipe input/output, nullable, filters dan pola connection. Alur cetak dan operasi yang membutuhkan penyedia bukti dapat terbit di schema meskipun jalur runtime belum dapat diselesaikan. [Instruksi perbaikan dan kriteria penerimaan](../../../agents/instructions/hrm/11-leader-remediasi-logika-operasi-dan-otorisasi-hrm-instructions.md).

### hrm_access_request.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_access_request.graphql)

- **Query (2)** — `employeeAccountRequest`, `workerEmployeeAccountRequest`.

- **Mutation (1)** — `requestEmployeeAccount`.

### hrm_accessrequest_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_accessrequest_entities.graphql)

- **Query (2)** — `accessRequest`, `accessRequests`.

- **Mutation (3)** — `createAccessRequest`, `updateAccessRequest`, `deleteAccessRequest`.

### hrm_analytics_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_analytics_entities.graphql)

- **Query (4)** — `humanCapitalMetricDefinition`, `humanCapitalMetricDefinitions`, `humanCapitalMetricSnapshot`, `humanCapitalMetricSnapshots`.

- **Mutation (6)** — `createHumanCapitalMetricDefinition`, `updateHumanCapitalMetricDefinition`, `deleteHumanCapitalMetricDefinition`, `createHumanCapitalMetricSnapshot`, `updateHumanCapitalMetricSnapshot`, `deleteHumanCapitalMetricSnapshot`.

### hrm_benefits_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_benefits_commands.graphql)

- **Mutation (2)** — `approveBenefitEnrollment`, `endBenefitEnrollment`.

### hrm_benefits_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_benefits_entities.graphql)

- **Query (4)** — `benefitPlan`, `benefitPlans`, `benefitEnrollment`, `benefitEnrollments`.

- **Mutation (6)** — `createBenefitPlan`, `updateBenefitPlan`, `deleteBenefitPlan`, `createBenefitEnrollment`, `updateBenefitEnrollment`, `deleteBenefitEnrollment`.

### hrm_candidate.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_candidate.graphql)

- **Query (2)** — `candidate`, `candidates`.

- **Mutation (3)** — `createCandidate`, `updateCandidate`, `deleteCandidate`.

### hrm_career_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_career_commands.graphql)

- **Mutation (4)** — `completeDevelopmentPlan`, `approveDevelopmentPlan`, `completeDevelopmentAction`, `approveCareerPath`.

### hrm_career_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_career_entities.graphql)

- **Query (10)** — `developmentPlan`, `developmentPlans`, `developmentAction`, `developmentActions`, `careerPath`, `careerPaths`, `careerPathTransition`, `careerPathTransitions`, `workerCareerAspiration`, `workerCareerAspirations`.

- **Mutation (15)** — `createDevelopmentPlan`, `updateDevelopmentPlan`, `deleteDevelopmentPlan`, `createDevelopmentAction`, `updateDevelopmentAction`, `deleteDevelopmentAction`, `createCareerPath`, `updateCareerPath`, `deleteCareerPath`, `createCareerPathTransition`, `updateCareerPathTransition`, `deleteCareerPathTransition`, `createWorkerCareerAspiration`, `updateWorkerCareerAspiration`, `deleteWorkerCareerAspiration`.

### hrm_compensation_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_compensation_commands.graphql)

- **Mutation (1)** — `approveCompensationRecord`.

### hrm_compensation_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_compensation_entities.graphql)

- **Query (4)** — `compensationBand`, `compensationBands`, `compensationRecord`, `compensationRecords`.

- **Mutation (6)** — `createCompensationBand`, `updateCompensationBand`, `deleteCompensationBand`, `createCompensationRecord`, `updateCompensationRecord`, `deleteCompensationRecord`.

### hrm_competency_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_competency_entities.graphql)

- **Query (8)** — `workerQualification`, `workerQualifications`, `competency`, `competencies`, `jobCompetency`, `jobCompetencies`, `workerCompetencyAssessment`, `workerCompetencyAssessments`.

- **Mutation (12)** — `createWorkerQualification`, `updateWorkerQualification`, `deleteWorkerQualification`, `createCompetency`, `updateCompetency`, `deleteCompetency`, `createJobCompetency`, `updateJobCompetency`, `deleteJobCompetency`, `createWorkerCompetencyAssessment`, `updateWorkerCompetencyAssessment`, `deleteWorkerCompetencyAssessment`.

### hrm_employeerelations_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_employeerelations_entities.graphql)

- **Query (6)** — `employeeRelationsCase`, `employeeRelationsCases`, `employeeRelationsEvent`, `employeeRelationsEvents`, `collectiveAgreement`, `collectiveAgreements`.

- **Mutation (9)** — `createEmployeeRelationsCase`, `updateEmployeeRelationsCase`, `deleteEmployeeRelationsCase`, `createEmployeeRelationsEvent`, `updateEmployeeRelationsEvent`, `deleteEmployeeRelationsEvent`, `createCollectiveAgreement`, `updateCollectiveAgreement`, `deleteCollectiveAgreement`.

### hrm_engagement_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_engagement_entities.graphql)

- **Query (4)** — `engagementSurvey`, `engagementSurveys`, `engagementSurveyResult`, `engagementSurveyResults`.

- **Mutation (3)** — `createEngagementSurvey`, `updateEngagementSurvey`, `deleteEngagementSurvey`.

### hrm_hrservice_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_hrservice_commands.graphql)

- **Mutation (4)** — `resolveServiceRequest`, `closeServiceRequest`, `cancelServiceRequest`, `assignServiceRequest`.

### hrm_hrservice_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_hrservice_entities.graphql)

- **Query (4)** — `serviceRequest`, `serviceRequests`, `serviceRequestEvent`, `serviceRequestEvents`.

- **Mutation (6)** — `createServiceRequest`, `updateServiceRequest`, `deleteServiceRequest`, `createServiceRequestEvent`, `updateServiceRequestEvent`, `deleteServiceRequestEvent`.

### hrm_job.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_job.graphql)

- **Query (5)** — `job`, `jobsByIds`, `jobs`, `jobFamily`, `jobFamilies`.

- **Mutation (2)** — `createJobFamily`, `createJob`.

### hrm_learning_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_learning_commands.graphql)

- **Mutation (3)** — `startLearningEnrollment`, `completeLearningEnrollment`, `withdrawLearningEnrollment`.

### hrm_learning_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_learning_entities.graphql)

- **Query (6)** — `learningCourse`, `learningCourses`, `learningEnrollment`, `learningEnrollments`, `learningCourseCompetency`, `learningCourseCompetencies`.

- **Mutation (9)** — `createLearningCourse`, `updateLearningCourse`, `deleteLearningCourse`, `createLearningEnrollment`, `updateLearningEnrollment`, `deleteLearningEnrollment`, `createLearningCourseCompetency`, `updateLearningCourseCompetency`, `deleteLearningCourseCompetency`.

### hrm_leave_crud_extra.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_leave_crud_extra.graphql)

- **Mutation (8)** — `createLeavePolicy`, `updateLeavePolicy`, `deleteLeavePolicy`, `createLeaveEntitlement`, `updateLeaveEntitlement`, `deleteLeaveEntitlement`, `updateLeaveAdjustment`, `deleteLeaveAdjustment`.

### hrm_leave_policy.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_leave_policy.graphql)

- **Query (8)** — `leavePolicy`, `leavePolicies`, `leaveEntitlement`, `leaveEntitlements`, `leaveLedgerEntry`, `leaveLedgerEntries`, `leaveAdjustment`, `leaveAdjustments`.

- **Mutation (5)** — `approveLeaveAdjustment`, `rejectLeaveAdjustment`, `createLeaveAdjustment`, `closeLeaveEntitlement`, `postLeaveAdjustment`.

### hrm_offboarding.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_offboarding.graphql)

- **Query (4)** — `offboardingPlan`, `offboardingPlans`, `offboardingTask`, `offboardingTasks`.

- **Mutation (8)** — `createOffboardingPlan`, `createOffboardingTask`, `completeOffboardingTask`, `waiveOffboardingTask`, `approveOffboardingPlan`, `startOffboardingPlan`, `cancelOffboardingPlan`, `completeOffboardingPlan`.

### hrm_offboarding_crud_extra.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_offboarding_crud_extra.graphql)

- **Mutation (4)** — `updateOffboardingPlan`, `deleteOffboardingPlan`, `updateOffboardingTask`, `deleteOffboardingTask`.

### hrm_onboarding_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_onboarding_commands.graphql)

- **Mutation (3)** — `startOnboardingPlan`, `completeOnboardingPlan`, `completeOnboardingTask`.

### hrm_onboarding_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_onboarding_entities.graphql)

- **Query (4)** — `onboardingPlan`, `onboardingPlans`, `onboardingTask`, `onboardingTasks`.

- **Mutation (6)** — `createOnboardingPlan`, `updateOnboardingPlan`, `deleteOnboardingPlan`, `createOnboardingTask`, `updateOnboardingTask`, `deleteOnboardingTask`.

### hrm_payroll_crud_extra.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_payroll_crud_extra.graphql)

- **Mutation (11)** — `createPayrollPeriod`, `updatePayrollPeriod`, `deletePayrollPeriod`, `createPayrollGroup`, `updatePayrollGroup`, `deletePayrollGroup`, `createPayrollRuleSet`, `updatePayrollRuleSet`, `deletePayrollRuleSet`, `updatePayrollRun`, `deletePayrollRun`.

### hrm_payroll_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_payroll_entities.graphql)

- **Query (6)** — `payrollProfile`, `payrollProfiles`, `payComponent`, `payComponents`, `payrollAdjustment`, `payrollAdjustments`.

- **Mutation (9)** — `createPayrollProfile`, `updatePayrollProfile`, `deletePayrollProfile`, `createPayComponent`, `updatePayComponent`, `deletePayComponent`, `createPayrollAdjustment`, `updatePayrollAdjustment`, `deletePayrollAdjustment`.

### hrm_payroll_posting.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_payroll_posting.graphql)

- **Query (8)** — `payrollRun`, `payrollRuns`, `payrollPeriod`, `payrollPeriods`, `payrollGroup`, `payrollGroups`, `payrollRuleSet`, `payrollRuleSets`.

- **Mutation (7)** — `requestPayrollPosting`, `calculatePayrollRun`, `submitPayrollRunForReview`, `approvePayrollRun`, `cancelPayrollRun`, `freezePayrollInput`, `createPayrollRun`.

### hrm_payroll_result.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_payroll_result.graphql)

- **Query (4)** — `payrollResult`, `payrollResults`, `payrollResultLine`, `payrollResultLines`.

### hrm_payslip.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql)

- **Query (3)** — `myPayslip`, `myPayslipPdf`, `myPayslips`.

- **Mutation (1)** — `reissueEmployeePayslip`.

### hrm_performance_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_performance_commands.graphql)

- **Mutation (7)** — `launchPerformanceCycle`, `closePerformanceCycle`, `submitPerformanceGoal`, `approvePerformanceGoal`, `rejectPerformanceGoal`, `submitPerformanceReview`, `finalizePerformanceReview`.

### hrm_performance_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_performance_entities.graphql)

- **Query (10)** — `performanceCycle`, `performanceCycles`, `performanceGoal`, `performanceGoals`, `performanceReview`, `performanceReviews`, `performanceGoalUpdate`, `performanceGoalUpdates`, `performanceReviewGoalScore`, `performanceReviewGoalScores`.

- **Mutation (15)** — `createPerformanceCycle`, `updatePerformanceCycle`, `deletePerformanceCycle`, `createPerformanceGoal`, `updatePerformanceGoal`, `deletePerformanceGoal`, `createPerformanceReview`, `updatePerformanceReview`, `deletePerformanceReview`, `createPerformanceGoalUpdate`, `updatePerformanceGoalUpdate`, `deletePerformanceGoalUpdate`, `createPerformanceReviewGoalScore`, `updatePerformanceReviewGoalScore`, `deletePerformanceReviewGoalScore`.

### hrm_personnel_crud_extra.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_personnel_crud_extra.graphql)

- **Mutation (3)** — `deleteWorker`, `updateWorkerAssignment`, `deleteWorkerAssignment`.

### hrm_personnel_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_personnel_entities.graphql)

- **Query (6)** — `workerLifecycleEvent`, `workerLifecycleEvents`, `workerDocument`, `workerDocuments`, `emergencyContact`, `emergencyContacts`.

- **Mutation (6)** — `createWorkerDocument`, `updateWorkerDocument`, `deleteWorkerDocument`, `createEmergencyContact`, `updateEmergencyContact`, `deleteEmergencyContact`.

### hrm_policy_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_policy_commands.graphql)

- **Mutation (1)** — `acknowledgePolicy`.

### hrm_policy_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_policy_entities.graphql)

- **Query (4)** — `policy`, `policies`, `policyAcknowledgement`, `policyAcknowledgements`.

- **Mutation (6)** — `createPolicy`, `updatePolicy`, `deletePolicy`, `createPolicyAcknowledgement`, `updatePolicyAcknowledgement`, `deletePolicyAcknowledgement`.

### hrm_position.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_position.graphql)

- **Query (2)** — `position`, `positions`.

- **Mutation (4)** — `activatePosition`, `createPosition`, `updatePosition`, `deletePosition`.

### hrm_promotion_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_promotion_commands.graphql)

- **Mutation (5)** — `submitPromotionRecommendation`, `approvePromotionRecommendation`, `rejectPromotionRecommendation`, `returnPromotionRecommendation`, `applyPromotionRecommendation`.

### hrm_promotion_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_promotion_entities.graphql)

- **Query (4)** — `promotionRecommendation`, `promotionRecommendations`, `promotionDecision`, `promotionDecisions`.

- **Mutation (6)** — `createPromotionRecommendation`, `updatePromotionRecommendation`, `deletePromotionRecommendation`, `createPromotionDecision`, `updatePromotionDecision`, `deletePromotionDecision`.

### hrm_recruitment_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_recruitment_commands.graphql)

- **Mutation (15)** — `startApplicationScreening`, `shortlistApplication`, `rejectApplication`, `startInterview`, `completeInterview`, `cancelInterview`, `approveOffer`, `issueOffer`, `acceptOffer`, `declineOffer`, `withdrawOffer`, `publishJobPosting`, `closeJobPosting`, `withdrawJobPosting`, `completeApplicationAssessment`.

### hrm_recruitment_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_recruitment_entities.graphql)

- **Query (16)** — `application`, `applications`, `applicationStageHistory`, `applicationStageHistories`, `interview`, `interviews`, `interviewPanelist`, `interviewPanelists`, `interviewFeedback`, `interviewFeedbacks`, `offer`, `offers`, `jobPosting`, `jobPostings`, `applicationAssessment`, `applicationAssessments`.

- **Mutation (21)** — `createApplication`, `updateApplication`, `deleteApplication`, `createInterview`, `updateInterview`, `deleteInterview`, `createInterviewPanelist`, `updateInterviewPanelist`, `deleteInterviewPanelist`, `createInterviewFeedback`, `updateInterviewFeedback`, `deleteInterviewFeedback`, `createOffer`, `updateOffer`, `deleteOffer`, `createJobPosting`, `updateJobPosting`, `deleteJobPosting`, `createApplicationAssessment`, `updateApplicationAssessment`, `deleteApplicationAssessment`.

### hrm_requisition.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_requisition.graphql)

- **Query (2)** — `requisition`, `requisitions`.

- **Mutation (8)** — `closeRequisition`, `approveRequisition`, `openRequisition`, `updateRequisition`, `deleteRequisition`, `rejectRequisition`, `submitRequisition`, `createRequisition`.

### hrm_succession_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_succession_commands.graphql)

- **Mutation (1)** — `approveSuccessionPlan`.

### hrm_succession_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_succession_entities.graphql)

- **Query (4)** — `successionPlan`, `successionPlans`, `successionCandidate`, `successionCandidates`.

- **Mutation (6)** — `createSuccessionPlan`, `updateSuccessionPlan`, `deleteSuccessionPlan`, `createSuccessionCandidate`, `updateSuccessionCandidate`, `deleteSuccessionCandidate`.

### hrm_talent_commands.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_talent_commands.graphql)

- **Mutation (11)** — `launchTalentReviewCycle`, `startTalentReviewCycleCalibration`, `closeTalentReviewCycle`, `submitTalentAssessment`, `withdrawTalentAssessment`, `finalizeTalentAssessment`, `openTalentCalibrationSession`, `recordTalentCalibrationDecision`, `approveTalentPoolMembership`, `rejectTalentPoolMembership`, `exitTalentPoolMembership`.

### hrm_talent_entities.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_talent_entities.graphql)

- **Query (18)** — `talentFramework`, `talentFrameworks`, `talentDimension`, `talentDimensions`, `talentReviewCycle`, `talentReviewCycles`, `talentAssessment`, `talentAssessments`, `talentAssessmentScore`, `talentAssessmentScores`, `talentCalibrationSession`, `talentCalibrationSessions`, `talentCalibrationDecision`, `talentCalibrationDecisions`, `talentPool`, `talentPools`, `talentPoolMembership`, `talentPoolMemberships`.

- **Mutation (27)** — `createTalentFramework`, `updateTalentFramework`, `deleteTalentFramework`, `createTalentDimension`, `updateTalentDimension`, `deleteTalentDimension`, `createTalentReviewCycle`, `updateTalentReviewCycle`, `deleteTalentReviewCycle`, `createTalentAssessment`, `updateTalentAssessment`, `deleteTalentAssessment`, `createTalentAssessmentScore`, `updateTalentAssessmentScore`, `deleteTalentAssessmentScore`, `createTalentCalibrationSession`, `updateTalentCalibrationSession`, `deleteTalentCalibrationSession`, `createTalentCalibrationDecision`, `updateTalentCalibrationDecision`, `deleteTalentCalibrationDecision`, `createTalentPool`, `updateTalentPool`, `deleteTalentPool`, `createTalentPoolMembership`, `updateTalentPoolMembership`, `deleteTalentPoolMembership`.

### hrm_worker.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_worker.graphql)

- **Query (8)** — `worker`, `workers`, `workerAssignmentTransfers`, `workerAssignmentTransfer`, `workerAssignment`, `workerAssignments`, `employmentContract`, `employmentContracts`.

- **Mutation (13)** — `createWorker`, `updateWorker`, `activateWorker`, `terminateWorker`, `applyPrimaryAssignment`, `approveWorkerAssignment`, `createWorkerAssignment`, `submitWorkerAssignmentTransfer`, `decideWorkerAssignmentTransfer`, `createEmploymentContract`, `recordSignedEmploymentContract`, `deleteEmploymentContract`, `updateEmploymentContract`.

### hrm_workforce_plan.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_workforce_plan.graphql)

- **Query (4)** — `workforcePlan`, `workforcePlans`, `workforcePlanLine`, `workforcePlanLines`.

- **Mutation (10)** — `submitWorkforcePlan`, `cancelWorkforcePlan`, `deleteWorkforcePlan`, `approveWorkforcePlan`, `supersedeWorkforcePlan`, `createWorkforcePlan`, `updateWorkforcePlan`, `updateWorkforcePlanLine`, `createWorkforcePlanLine`, `deleteWorkforcePlanLine`.

### hrm_workforceplanning_crud_extra.graphql

[SDL sumber](../../../backend/gateway/graphql/human_resource/hrm_workforceplanning_crud_extra.graphql)

- **Mutation (4)** — `updateJobFamily`, `deleteJobFamily`, `updateJob`, `deleteJob`.
