# List operasi

`human_resource.recruitment.v1.RecruitmentService` memiliki 145 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/recruitment/v1/recruitment.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateRequisition` | Membuat entri `requisitions`. | Mutation | `CreateRequisitionRequest` → `RequisitionResponse` | `human_resource.requisitions.create` |
| `UpdateRequisition` | Mengubah entri `requisitions` sesuai field mask. | Mutation | `UpdateRequisitionRequest` → `RequisitionResponse` | `human_resource.requisitions.update` |
| `SubmitRequisition` | Mengajukan entri `requisitions` melalui transisi workflow HRM. | Mutation | `IDRequest` → `RequisitionResponse` | `human_resource.requisitions.submit` |
| `ApproveRequisition` | Menyetujui entri `requisitions` melalui transisi workflow HRM. | Mutation | `IDRequest` → `RequisitionResponse` | `human_resource.requisitions.approve` |
| `RejectRequisition` | Menolak entri `requisitions` melalui transisi workflow HRM. | Mutation | `RejectRequisitionRequest` → `RequisitionResponse` | `human_resource.requisitions.reject` |
| `OpenRequisition` | Membuka entri `requisitions` melalui transisi workflow HRM. | Mutation | `IDRequest` → `RequisitionResponse` | `human_resource.requisitions.open` |
| `CloseRequisition` | Menutup entri `requisitions` melalui transisi workflow HRM. | Mutation | `IDRequest` → `RequisitionResponse` | `human_resource.requisitions.close` |
| `DeleteRequisition` | Menghapus entri `requisitions` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.requisitions.delete` |
| `GetRequisitionByID` | Mengambil entri `requisitions` berdasarkan ID. | Query | `IDRequest` → `RequisitionResponse` | `human_resource.requisitions.read` |
| `GetAllRequisitions` | Mendaftar entri `requisitions` dengan paginasi. | Query | `PageRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByRequisitionNumber` | Menyaring daftar `requisitions` berdasarkan `requisition_number`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByWorkforcePlanLineID` | Menyaring daftar `requisitions` berdasarkan `workforce_plan_line_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByPositionID` | Menyaring daftar `requisitions` berdasarkan `position_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByJobID` | Menyaring daftar `requisitions` berdasarkan `job_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByOrganizationUnitReferenceID` | Menyaring daftar `requisitions` berdasarkan `organization_unit_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByHiringManagerPrincipalID` | Menyaring daftar `requisitions` berdasarkan `hiring_manager_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByRecruiterPrincipalID` | Menyaring daftar `requisitions` berdasarkan `recruiter_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByRequisitionType` | Menyaring daftar `requisitions` berdasarkan `requisition_type`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByStatus` | Menyaring daftar `requisitions` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByBudgetReferenceID` | Menyaring daftar `requisitions` berdasarkan `budget_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByWorkLocationReferenceID` | Menyaring daftar `requisitions` berdasarkan `work_location_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByApprovedByPrincipalID` | Menyaring daftar `requisitions` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByRejectionReasonCode` | Menyaring daftar `requisitions` berdasarkan `rejection_reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByRejectedByPrincipalID` | Menyaring daftar `requisitions` berdasarkan `rejected_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `FindRequisitionByIdempotencyKey` | Menyaring daftar `requisitions` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `RequisitionsResponse` | `human_resource.requisitions.read` |
| `CreateCandidate` | Membuat entri `candidates`. | Mutation | `CreateCandidateRequest` → `CandidateResponse` | `human_resource.candidates.create` |
| `UpdateCandidate` | Mengubah entri `candidates` sesuai field mask. | Mutation | `UpdateCandidateRequest` → `CandidateResponse` | `human_resource.candidates.update` |
| `DeleteCandidate` | Menghapus entri `candidates` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.candidates.delete` |
| `GetCandidateByID` | Mengambil entri `candidates` berdasarkan ID. | Query | `IDRequest` → `CandidateResponse` | `human_resource.candidates.read` |
| `GetAllCandidates` | Mendaftar entri `candidates` dengan paginasi. | Query | `PageRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `FindCandidateByCandidateNumber` | Menyaring daftar `candidates` berdasarkan `candidate_number`, dengan paginasi. | Query | `StringFinderRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `FindCandidateByPersonPartyReferenceID` | Menyaring daftar `candidates` berdasarkan `person_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `FindCandidateByContactReferenceID` | Menyaring daftar `candidates` berdasarkan `contact_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `FindCandidateBySourceType` | Menyaring daftar `candidates` berdasarkan `source_type`, dengan paginasi. | Query | `StringFinderRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `FindCandidateByCandidateStatus` | Menyaring daftar `candidates` berdasarkan `candidate_status`, dengan paginasi. | Query | `StringFinderRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `FindCandidateByCreatedByPrincipalID` | Menyaring daftar `candidates` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `CandidatesResponse` | `human_resource.candidates.read` |
| `CreateApplication` | Membuat entri `applications`. | Mutation | `CreateApplicationRequest` → `ApplicationResponse` | `human_resource.applications.create` |
| `UpdateApplication` | Mengubah entri `applications` sesuai field mask. | Mutation | `UpdateApplicationRequest` → `ApplicationResponse` | `human_resource.applications.update` |
| `StartApplicationScreening` | Memulai penyaringan pada entri `applications` melalui transisi workflow HRM. | Mutation | `IDRequest` → `ApplicationResponse` | `human_resource.applications.start_screening` |
| `ShortlistApplication` | Memajukan tahap entri `applications` melalui transisi workflow HRM. | Mutation | `ShortlistApplicationRequest` → `ApplicationResponse` | `human_resource.applications.advance` |
| `RejectApplication` | Menolak entri `applications` melalui transisi workflow HRM. | Mutation | `RejectApplicationRequest` → `ApplicationResponse` | `human_resource.applications.reject` |
| `DeleteApplication` | Menghapus entri `applications` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.applications.delete` |
| `GetApplicationByID` | Mengambil entri `applications` berdasarkan ID. | Query | `IDRequest` → `ApplicationResponse` | `human_resource.applications.read` |
| `GetAllApplications` | Mendaftar entri `applications` dengan paginasi. | Query | `PageRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByCandidateID` | Menyaring daftar `applications` berdasarkan `candidate_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByRequisitionID` | Menyaring daftar `applications` berdasarkan `requisition_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByApplicationNumber` | Menyaring daftar `applications` berdasarkan `application_number`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByStatus` | Menyaring daftar `applications` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByCurrentStageCode` | Menyaring daftar `applications` berdasarkan `current_stage_code`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByAssignedRecruiterPrincipalID` | Menyaring daftar `applications` berdasarkan `assigned_recruiter_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByDispositionReasonCode` | Menyaring daftar `applications` berdasarkan `disposition_reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByDecisionByPrincipalID` | Menyaring daftar `applications` berdasarkan `decision_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByHiredWorkerID` | Menyaring daftar `applications` berdasarkan `hired_worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationBySourcePostingID` | Menyaring daftar `applications` berdasarkan `source_posting_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `FindApplicationByIdempotencyKey` | Menyaring daftar `applications` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationsResponse` | `human_resource.applications.read` |
| `CreateApplicationStageHistory` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `CreateApplicationStageHistoryRequest` → `ApplicationStageHistoryResponse` | — |
| `UpdateApplicationStageHistory` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `UpdateApplicationStageHistoryRequest` → `ApplicationStageHistoryResponse` | — |
| `DeleteApplicationStageHistory` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `IdentityIDRequest` → `EmptyResponse` | — |
| `GetApplicationStageHistoryByID` | Mengambil entri `application_stage_history` berdasarkan ID. | Query | `IdentityIDRequest` → `ApplicationStageHistoryResponse` | `human_resource.application_stage_history.read` |
| `GetAllApplicationStageHistories` | Mendaftar entri `application_stage_history` dengan paginasi. | Query | `PageRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `FindApplicationStageHistoryByApplicationID` | Menyaring daftar `application_stage_history` berdasarkan `application_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `FindApplicationStageHistoryByPreviousStatus` | Menyaring daftar `application_stage_history` berdasarkan `previous_status`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `FindApplicationStageHistoryByNewStatus` | Menyaring daftar `application_stage_history` berdasarkan `new_status`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `FindApplicationStageHistoryByStageCode` | Menyaring daftar `application_stage_history` berdasarkan `stage_code`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `FindApplicationStageHistoryByChangedByPrincipalID` | Menyaring daftar `application_stage_history` berdasarkan `changed_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `FindApplicationStageHistoryByReasonCode` | Menyaring daftar `application_stage_history` berdasarkan `reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationStageHistoriesResponse` | `human_resource.application_stage_history.read` |
| `StartInterview` | Memulai wawancara pada jadwalnya dan memindahkan lamaran shortlist ke tahap interview. | Mutation | `IDRequest` → `InterviewResponse` | `human_resource.interviews.start` |
| `CompleteInterview` | Menutup wawancara berlangsung setelah setiap panelis mencatat feedback yang sah. | Mutation | `IDRequest` → `InterviewResponse` | `human_resource.interviews.complete` |
| `CancelInterview` | Membatalkan wawancara planned atau scheduled dan membebaskan slot tanpa menghapus panelis. | Mutation | `IDRequest` → `InterviewResponse` | `human_resource.interviews.cancel` |
| `CreateInterview` | Membuat entri `interviews`. | Mutation | `CreateInterviewRequest` → `InterviewResponse` | `human_resource.interviews.create` |
| `UpdateInterview` | Mengubah entri `interviews` sesuai field mask. | Mutation | `UpdateInterviewRequest` → `InterviewResponse` | `human_resource.interviews.update` |
| `DeleteInterview` | Menghapus entri `interviews` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.interviews.delete` |
| `GetInterviewByID` | Mengambil entri `interviews` berdasarkan ID. | Query | `IDRequest` → `InterviewResponse` | `human_resource.interviews.read` |
| `GetAllInterviews` | Mendaftar entri `interviews` dengan paginasi. | Query | `PageRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `FindInterviewByApplicationID` | Menyaring daftar `interviews` berdasarkan `application_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `FindInterviewByInterviewType` | Menyaring daftar `interviews` berdasarkan `interview_type`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `FindInterviewByLocationReferenceID` | Menyaring daftar `interviews` berdasarkan `location_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `FindInterviewByMeetingReferenceID` | Menyaring daftar `interviews` berdasarkan `meeting_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `FindInterviewByStatus` | Menyaring daftar `interviews` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `FindInterviewByCreatedByPrincipalID` | Menyaring daftar `interviews` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewsResponse` | `human_resource.interviews.read` |
| `CreateInterviewPanelist` | Membuat entri `interview_panelists`. | Mutation | `CreateInterviewPanelistRequest` → `InterviewPanelistResponse` | `human_resource.interview_panelists.create` |
| `UpdateInterviewPanelist` | Mengubah entri `interview_panelists` sesuai field mask. | Mutation | `UpdateInterviewPanelistRequest` → `InterviewPanelistResponse` | `human_resource.interview_panelists.update` |
| `DeleteInterviewPanelist` | Menghapus entri `interview_panelists` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.interview_panelists.delete` |
| `GetInterviewPanelistByID` | Mengambil entri `interview_panelists` berdasarkan ID. | Query | `IDRequest` → `InterviewPanelistResponse` | `human_resource.interview_panelists.read` |
| `GetAllInterviewPanelists` | Mendaftar entri `interview_panelists` dengan paginasi. | Query | `PageRequest` → `InterviewPanelistsResponse` | `human_resource.interview_panelists.read` |
| `FindInterviewPanelistByInterviewID` | Menyaring daftar `interview_panelists` berdasarkan `interview_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewPanelistsResponse` | `human_resource.interview_panelists.read` |
| `FindInterviewPanelistByInterviewerPrincipalID` | Menyaring daftar `interview_panelists` berdasarkan `interviewer_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewPanelistsResponse` | `human_resource.interview_panelists.read` |
| `FindInterviewPanelistByScorecardTemplateReferenceID` | Menyaring daftar `interview_panelists` berdasarkan `scorecard_template_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewPanelistsResponse` | `human_resource.interview_panelists.read` |
| `CreateInterviewFeedback` | Membuat entri `interview_feedback`. | Mutation | `CreateInterviewFeedbackRequest` → `InterviewFeedbackResponse` | `human_resource.interview_feedback.create` |
| `UpdateInterviewFeedback` | Mengubah entri `interview_feedback` sesuai field mask. | Mutation | `UpdateInterviewFeedbackRequest` → `InterviewFeedbackResponse` | `human_resource.interview_feedback.update` |
| `DeleteInterviewFeedback` | Menghapus entri `interview_feedback` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.interview_feedback.delete` |
| `GetInterviewFeedbackByID` | Mengambil entri `interview_feedback` berdasarkan ID. | Query | `IDRequest` → `InterviewFeedbackResponse` | `human_resource.interview_feedback.read` |
| `GetAllInterviewFeedbacks` | Mendaftar entri `interview_feedback` dengan paginasi. | Query | `PageRequest` → `InterviewFeedbacksResponse` | `human_resource.interview_feedback.read` |
| `FindInterviewFeedbackByInterviewID` | Menyaring daftar `interview_feedback` berdasarkan `interview_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewFeedbacksResponse` | `human_resource.interview_feedback.read` |
| `FindInterviewFeedbackByInterviewerPrincipalID` | Menyaring daftar `interview_feedback` berdasarkan `interviewer_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `InterviewFeedbacksResponse` | `human_resource.interview_feedback.read` |
| `ApproveOffer` | Menyetujui draf setelah verifier membuktikan versi kompensasi yang cocok dan penyetuju berbeda dari pembuat. | Mutation | `ApproveOfferRequest` → `OfferResponse` | `human_resource.offers.approve` |
| `IssueOffer` | Menerbitkan penawaran setelah versi dan digest dokumen final terverifikasi serta memindahkan lamaran ke tahap offer. | Mutation | `IssueOfferRequest` → `OfferResponse` | `human_resource.offers.issue` |
| `AcceptOffer` | Mencatat penerimaan yang sudah diverifikasi dari kanal kandidat. | Mutation | `AcceptOfferRequest` → `OfferResponse` | `human_resource.offers.accept` |
| `DeclineOffer` | Mencatat penolakan yang sudah diverifikasi dari kanal kandidat. | Mutation | `DeclineOfferRequest` → `OfferResponse` | `human_resource.offers.decline` |
| `WithdrawOffer` | Menarik penawaran terbit yang belum direspons kandidat dan mempertahankan bukti dokumen. | Mutation | `WithdrawOfferRequest` → `OfferResponse` | `human_resource.offers.withdraw` |
| `CreateOffer` | Membuat entri `offers`. | Mutation | `CreateOfferRequest` → `OfferResponse` | `human_resource.offers.create` |
| `UpdateOffer` | Mengubah entri `offers` sesuai field mask. | Mutation | `UpdateOfferRequest` → `OfferResponse` | `human_resource.offers.update` |
| `DeleteOffer` | Menghapus entri `offers` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.offers.delete` |
| `GetOfferByID` | Mengambil entri `offers` berdasarkan ID. | Query | `IDRequest` → `OfferResponse` | `human_resource.offers.read` |
| `GetAllOffers` | Mendaftar entri `offers` dengan paginasi. | Query | `PageRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByApplicationID` | Menyaring daftar `offers` berdasarkan `application_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByOfferNumber` | Menyaring daftar `offers` berdasarkan `offer_number`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByOfferStatus` | Menyaring daftar `offers` berdasarkan `offer_status`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByProposedJobID` | Menyaring daftar `offers` berdasarkan `proposed_job_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByProposedLegalEntityReferenceID` | Menyaring daftar `offers` berdasarkan `proposed_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByProposedContractType` | Menyaring daftar `offers` berdasarkan `proposed_contract_type`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByProposedCompensationDocumentVersionReferenceID` | Menyaring daftar `offers` berdasarkan `proposed_compensation_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByCurrencyCode` | Menyaring daftar `offers` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByOfferDocumentReferenceID` | Menyaring daftar `offers` berdasarkan `offer_document_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByOfferDocumentVersionReferenceID` | Menyaring daftar `offers` berdasarkan `offer_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `FindOfferByApprovedByPrincipalID` | Menyaring daftar `offers` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OffersResponse` | `human_resource.offers.read` |
| `CreateJobPosting` | Membuat entri `job_postings`. | Mutation | `CreateJobPostingRequest` → `JobPostingResponse` | `human_resource.job_postings.create` |
| `PublishJobPosting` | Menerbitkan entri `job_postings` melalui transisi workflow HRM. | Mutation | `IDRequest` → `JobPostingResponse` | `human_resource.job_postings.publish` |
| `CloseJobPosting` | Menutup entri `job_postings` melalui transisi workflow HRM. | Mutation | `IDRequest` → `JobPostingResponse` | `human_resource.job_postings.close` |
| `WithdrawJobPosting` | Menarik entri `job_postings` melalui transisi workflow HRM. | Mutation | `WithdrawJobPostingRequest` → `JobPostingResponse` | `human_resource.job_postings.withdraw` |
| `UpdateJobPosting` | Mengubah entri `job_postings` sesuai field mask. | Mutation | `UpdateJobPostingRequest` → `JobPostingResponse` | `human_resource.job_postings.update` |
| `DeleteJobPosting` | Menghapus entri `job_postings` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.job_postings.delete` |
| `GetJobPostingByID` | Mengambil entri `job_postings` berdasarkan ID. | Query | `IDRequest` → `JobPostingResponse` | `human_resource.job_postings.read` |
| `GetAllJobPostings` | Mendaftar entri `job_postings` dengan paginasi. | Query | `PageRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByRequisitionID` | Menyaring daftar `job_postings` berdasarkan `requisition_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByChannelCode` | Menyaring daftar `job_postings` berdasarkan `channel_code`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByContentDocumentVersionReferenceID` | Menyaring daftar `job_postings` berdasarkan `content_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByStatus` | Menyaring daftar `job_postings` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByCreatedByPrincipalID` | Menyaring daftar `job_postings` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByWithdrawalReasonCode` | Menyaring daftar `job_postings` berdasarkan `withdrawal_reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `FindJobPostingByWithdrawnByPrincipalID` | Menyaring daftar `job_postings` berdasarkan `withdrawn_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobPostingsResponse` | `human_resource.job_postings.read` |
| `CreateApplicationAssessment` | Membuat entri `application_assessments`. | Mutation | `CreateApplicationAssessmentRequest` → `ApplicationAssessmentResponse` | `human_resource.application_assessments.create` |
| `CompleteApplicationAssessment` | Menyelesaikan entri `application_assessments` melalui transisi workflow HRM. | Mutation | `CompleteApplicationAssessmentRequest` → `ApplicationAssessmentResponse` | `human_resource.application_assessments.complete` |
| `UpdateApplicationAssessment` | Mengubah entri `application_assessments` sesuai field mask. | Mutation | `UpdateApplicationAssessmentRequest` → `ApplicationAssessmentResponse` | `human_resource.application_assessments.update` |
| `DeleteApplicationAssessment` | Menghapus entri `application_assessments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.application_assessments.delete` |
| `GetApplicationAssessmentByID` | Mengambil entri `application_assessments` berdasarkan ID. | Query | `IDRequest` → `ApplicationAssessmentResponse` | `human_resource.application_assessments.read` |
| `GetAllApplicationAssessments` | Mendaftar entri `application_assessments` dengan paginasi. | Query | `PageRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByApplicationID` | Menyaring daftar `application_assessments` berdasarkan `application_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByAssessmentCode` | Menyaring daftar `application_assessments` berdasarkan `assessment_code`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByAttemptNumber` | Menyaring daftar `application_assessments` berdasarkan `attempt_number`, dengan paginasi. | Query | `Int32FinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByRubricDocumentVersionReferenceID` | Menyaring daftar `application_assessments` berdasarkan `rubric_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByStatus` | Menyaring daftar `application_assessments` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByAssessorPrincipalID` | Menyaring daftar `application_assessments` berdasarkan `assessor_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByRecommendationCode` | Menyaring daftar `application_assessments` berdasarkan `recommendation_code`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |
| `FindApplicationAssessmentByEvidenceDocumentVersionReferenceID` | Menyaring daftar `application_assessments` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `ApplicationAssessmentsResponse` | `human_resource.application_assessments.read` |

## Butuh GraphQL

95 kandidat Query dan 47 kandidat Mutation dan 3 RPC nonaktif. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

## Prasyarat panel wawancara

`CreateInterviewPanelist`, `UpdateInterviewPanelist`, dan `DeleteInterviewPanelist` berlaku sebelum wawancara berlangsung dan sebelum waktu mulai terjadwal. Gateway perlu memetakan `FailedPrecondition` jika status atau jadwal berubah sebelum mutation diproses HRM.

`UpdateInterviewFeedback` dan `DeleteInterviewFeedback` tunduk pada urutan versi. HRM menolak perubahan pada versi yang sudah digantikan dengan `FailedPrecondition`.

`CancelInterview` memerlukan MFA dan dijalankan oleh recruiter yang ditugaskan pada lamaran. Jika belum ada recruiter, pembuat wawancara dapat membatalkannya. Wawancara dengan feedback yang sudah tercatat tidak dapat dibatalkan.

`StartInterview` membutuhkan MFA, recruiter yang ditugaskan, lamaran di tahap shortlisted atau interview, waktu mulai yang sudah tercapai, waktu akhir yang belum tercapai, dan minimal satu panelis tercatat. Perpindahan pertama mencatat riwayat tahap dan event mulai wawancara dalam transaksi yang sama.

`CompleteInterview` membutuhkan MFA, recruiter yang ditugaskan, dan versi feedback terakhir dari setiap panelis. Feedback yang mendahului waktu mulai atau bertanggal mendatang tidak memenuhi syarat. Penyelesaian wawancara membekukan panel dan feedback. Tahap lamaran hanya berubah lewat workflow lamaran.

Mutasi feedback hanya dapat dijalankan oleh panelis yang tercatat sebagai pemilik feedback. Pembuatan dan perubahan memerlukan wawancara `in_progress` dan lamaran di tahap `interview`.

`CreateOffer` mencatat pembuat dan waktu server. Respons `Offer` mengembalikan `createdByPrincipalId` dan `createdAt`; keduanya bisa kosong pada data historis. Identitas pembuat tidak berasal dari input GraphQL.

`WithdrawOffer` meminta alasan dan MFA. Hanya recruiter yang ditugaskan dapat menarik penawaran berstatus `issued` tanpa respons kandidat. Response menyertakan waktu, principal, dan alasan penarikan.

`ApproveOffer` membutuhkan MFA dan bukti versi kompensasi yang disetujui. Service produksi saat ini mengembalikan `Unavailable` pada pengajuan yang memenuhi prasyarat lokal sampai adapter DRM dipasang.

`IssueOffer` membutuhkan MFA, recruiter yang ditugaskan, persetujuan independen, tenggat yang masih berlaku, dan bukti versi final serta digest SHA-256 dari DRM. Penawaran aktif lain pada lamaran yang sama menghalangi penerbitan. Jalur produksi mengembalikan `Unavailable` sampai adapter DRM terpasang.

PostgreSQL juga menjaga satu penawaran berstatus `issued` atau `accepted` per lamaran. Penarikan membebaskan slot untuk penerbitan revisi.

`AcceptOffer` dan `DeclineOffer` membawa referensi bukti saja. Recruiter yang ditugaskan memerlukan MFA dan izin terpisah per aksi. Verifier kandidat harus membuktikan keputusan, identitas kandidat, versi dokumen, digest, dan waktu respons. Runtime menolak operasi sampai verifier kanal kandidat tersedia.
Waktu respons dari bukti harus berada sejak waktu penerbitan hingga sebelum tenggat serta tidak boleh melampaui waktu server. Waktu pencatatan recruiter dapat terjadi setelah tenggat apabila kandidat merespons tepat waktu.

`IssueOffer` memeriksa ulang kandidat, status lowongan, pekerjaan, dan badan hukum saat penerbitan, termasuk untuk revisi pada lamaran yang sudah berada di tahap `offer`.

## Referensi individu kandidat

`CreateCandidate` menerima `person_party_reference_id` opsional. HRM memeriksa melalui Enterprise Foundation bahwa referensi menunjuk individu aktif. Identitas itu tetap ketika draf diperbarui. Referensi kontak opsional memerlukan referensi individu. HRM menerima kontak aktif yang belum dihapus dan terikat pada individu tersebut di Enterprise Foundation. Pembaruan draf mempertahankan referensi kontak. Sumber eksternal memerlukan kontrak validasi tersendiri. Untuk tampilan Gateway, gunakan nomor kandidat dan nama yang diperbolehkan bagi operator rekrutmen.

### Butuh GraphQL Mutation

`createCandidate(input: CreateCandidateInput!)` menerima nomor kandidat, nama pilihan opsional, versi pemberitahuan privasi, waktu pengakuan, batas retensi, dan referensi individu/kontak opsional. Pilihan UI mengirim UUID EF dari `_id` pada variabel permintaan. Payload menampilkan nomor kandidat, nama pilihan, dan status. ID hasil dipakai aplikasi sebagai referensi internal.

### RBAC pembuatan kandidat

Sesi employee memerlukan `human_resource.candidates.create` pada scope GLOBAL. IAM memuat definisi permission dari katalog resmi. Penugasan role dan izin dilakukan melalui GraphQL. [Uji HTTPS](../../e2e/gateway/144-hrm-mutation-pembuatan-kandidat-ef.md).

### Butuh GraphQL Query

`candidate(id: ID!)` mengambil detail. `candidates(candidate_number: String, after: String, before: String, first: Int, last: Int)` mengambil halaman kandidat dengan pencarian nomor persis. Keduanya memerlukan `human_resource.candidates.read` berscope GLOBAL. Tampilkan nomor kandidat, nama pilihan, dan status. ID dipakai sebagai referensi navigasi internal. [Uji HTTPS](../../e2e/gateway/145-hrm-query-daftar-dan-detail-kandidat.md).

`updateCandidate(id: ID!, input: UpdateCandidateInput!)` mengoreksi nama pilihan kandidat. Jika operator perlu mengosongkan nama, kirim `preferred_name: null`. Input kosong ditolak. HRM mempertahankan nomor kandidat, referensi individu dan kontak, serta bukti pemberitahuan privasi. Mutation memerlukan `human_resource.candidates.update` dengan scope GLOBAL. [Uji HTTPS](../../e2e/gateway/146-hrm-mutation-koreksi-nama-kandidat.md).

`deleteCandidate(id: ID!)` meminta penghapusan kandidat tanpa lamaran. HRM menolak kandidat yang sudah dirujuk lamaran. Mutation memerlukan `human_resource.candidates.delete`, scope GLOBAL, dan MFA terbaru. [Uji HTTPS](../../e2e/gateway/147-hrm-mutation-penghapusan-kandidat.md).

### Daftar dan detail requisition

`requisition(id: ID!)` memakai `GetRequisitionByID`. `requisitions` memakai `GetAllRequisitions` atau `FindRequisitionByRequisitionNumber` untuk filter nomor persis. Query memakai `human_resource.requisitions.read` dengan scope GLOBAL. Nomor requisition, status, jumlah kebutuhan, kategori pekerja, dan tanggal mulai target dapat ditampilkan. ID job dan unit organisasi menjadi referensi internal untuk mencari nama bisnis dari HRM dan Enterprise Foundation. [Uji HTTPS](../../e2e/gateway/148-hrm-query-daftar-detail-requisition.md).

### Pemilihan hiring manager saat membuat draf

RPC privat `CreateRequisition` menerima `hiring_manager_worker_id` pada request. HRM mengunci record pekerja aktif, mengambil principal yang sudah tertaut, lalu memvalidasi kembali kelayakan hiring manager, job, unit, dan badan hukum dalam transaksi pembuatan draf. Caller yang masih memakai principal langsung dapat memakai kontrak lama. Request yang mengirim kedua referensi ditolak. Gateway memakai ID pekerja untuk membuat draf.

## Penjelasan createRequisition

Membuat requisition berstatus `draft`. Headcount default satu. HRM memvalidasi pekerja aktif dengan akun IAM tertaut, job aktif, dan unit organisasi terhadap badan hukum terkait.

## Butuh GraphQL Mutation

`createRequisition(input: CreateRequisitionInput!)` mengembalikan requisition dan hasil mutation. Detail draf dapat dibaca melalui `requisition`.

## RBAC

Pembuatan memerlukan `human_resource.requisitions.create` dengan grant global bagi akun employee. Pembacaan daftar dan detail memakai `human_resource.requisitions.read`.

## Data untuk UI

- Requisition ditampilkan dengan nomor, status, tipe, kategori kerja, headcount, dan tanggal mulai bila tersedia.
- Hiring manager dipilih lewat nama dan nomor pekerja. Simpan ID pekerja sebagai nilai pilihan. IAM dan HRM mengurus tautan principal secara internal.
- Pilihan job memakai kode dan judul jabatan. Pilihan unit memakai kode dan nama dari Enterprise Foundation. Detail requisition saat ini mengembalikan referensi job dan unit, sehingga UI memerlukan data label dari sumber tersebut. Jika label belum tersedia, tampilkan informasi bahwa label belum tersedia dan simpan ID untuk lookup internal.
- ID dan token `gid://` digunakan untuk nilai pilihan, lookup, atau navigasi internal. Gunakan label bisnis pada tabel, detail, konfirmasi, dan pesan validasi.

| Label di UI | Query dan field | Permission |
| --- | --- | --- |
| Nomor pekerja | `workers` atau `worker` → `worker_number`, `person_party_reference_id` | `human_resource.workers.read` |
| Nama hiring manager | `partiesByIds(ids: ...)` dengan referensi party dari pekerja → `display_name` | `enterprise_foundation.parties.read` pada scope yang berlaku |
| Kode dan judul jabatan | `job(id: ...)` atau daftar pilihan `jobs` → `job_code`, `job_title` | `human_resource.jobs.read` |
| Kode dan nama unit | `organizationUnitsByIds(ids: ...)` → `unit_code`, `name` | `enterprise_foundation.organization_units.read` pada scope yang berlaku |

Cocokkan hasil lookup dengan ID referensi, karena urutan respons batch dapat berbeda dari urutan baris. Gabungkan referensi party dan unit unik per halaman. Jika izin atau data label belum tersedia, gunakan nomor pekerja bila ada atau teks seperti “Unit organisasi tidak tersedia”. Pertahankan kode bisnis sesuai nilai dari pemilik data. [Verifikasi HTTPS label](../../e2e/gateway/149-hrm-mutation-buat-draf-requisition.md).

## Penjelasan submitRequisition

Pembuat draf mengajukan requisition untuk persetujuan. HRM mencatat waktu dan aktor pengajuan dalam transaksi yang sama. Requisition yang sudah diajukan akan menolak pengajuan ulang.

## Butuh GraphQL Mutation

`submitRequisition(id: ID!)` menerima ID dari detail atau hasil pembuatan. Payload mengembalikan requisition berstatus `submitted`. Nomor requisition tetap menjadi label di konfirmasi UI. Status dapat ditampilkan sebagai “Menunggu persetujuan”.

## RBAC

Memerlukan akun employee, grant global `human_resource.requisitions.submit` dan sesi dengan MFA. HRM memastikan aktor manusia yang mengajukan adalah pembuat draf. [Bukti HTTPS](../../e2e/gateway/150-hrm-mutation-submit-requisition-mfa.md).

## Penjelasan rejectRequisition

Peninjau menolak requisition berstatus `submitted` dengan kode alasan. HRM mencatat aktor, waktu, dan alasan penolakan secara atomik. UI menampilkan nomor requisition dan status “Ditolak”. Petakan kode alasan ke label yang dipilih peninjau.

## Butuh GraphQL Mutation

`rejectRequisition(id: ID!, reason_code: String!)` mengembalikan status `rejected` dan `rejection_reason_code`. Kode alasan juga tersedia pada Query daftar dan detail. Kode wajib terisi dan dibatasi 80 byte setelah spasi awal dan akhir dibuang.

## RBAC

Memerlukan akun employee, MFA, dan grant global `human_resource.requisitions.reject`. Peninjau wajib berbeda dari pembuat dan pengaju. Pembacaan alasan mengikuti permission baca requisition.

Persetujuan memakai RPC privat `ApproveRequisition` melalui Gateway. IAM membuktikan mandat unit dan Enterprise Foundation menjadi sumber referensi unit. Finance perlu menyediakan otorisasi anggaran yang mengikat snapshot requisition. Pembukaan memeriksa ulang UUID otorisasi yang disimpan. [Handoff Finance](37-kebutuhan-finance-anggaran-requisition.md).

## Penjelasan deleteRequisition

Menghapus requisition berstatus `draft`. HRM menjaga status di dalam operasi agar requisition yang sudah diajukan tetap tersedia. Konfirmasi UI memakai nomor requisition.

## Butuh GraphQL Mutation

`deleteRequisition(id: ID!)` mengembalikan `deleted` dan status hasil operasi. Setelah berhasil, hapus entri dari daftar UI atau muat ulang daftar.

## RBAC

Memerlukan employee, grant global `human_resource.requisitions.delete`, dan MFA. [Bukti HTTPS](../../e2e/gateway/152-hrm-mutation-hapus-draf-requisition.md).

## Penjelasan updateRequisition

Mengubah headcount, tanggal mulai target, tipe, kategori kerja, jabatan, unit organisasi, lokasi kerja, posisi, atau lini workforce plan pada draf. HRM memvalidasi gabungan referensi setelah perubahan. Hiring manager dapat diganti melalui `hiring_manager_worker_id`. HRM mengambil principal yang sudah ditautkan ke pekerja aktif.

## Butuh GraphQL Mutation

`updateRequisition(id: ID!, input: UpdateRequisitionInput!)` mengembalikan draf terbaru. Field yang tidak dikirim mempertahankan isinya. Nilai null dapat mengosongkan tanggal target, lokasi, posisi, dan lini workforce plan. Jabatan, unit, tipe, kategori kerja, serta headcount harus bernilai bila dikirim. Input referensi menerima UUID atau Global ID dengan jenis yang sesuai. Form menampilkan nomor requisition serta label bisnis yang sudah tersedia dari Query HRM dan Enterprise Foundation.

## RBAC

Memerlukan akun employee dan grant global `human_resource.requisitions.update`. Setelah draf diajukan, HRM menolak perubahan melalui operasi ini.

[Bukti HTTPS edit referensi draf](../../e2e/gateway/163-hrm-mutation-edit-referensi-requisition.md).

[Bukti HTTPS perubahan hiring manager](../../e2e/gateway/164-hrm-mutation-ganti-hiring-manager-draf.md).

Pada detail `requisition(id: ...)` dan refetch `node(id: ...)`, `hiring_manager_worker_id` tersedia bila tautan principal dapat ditentukan ke tepat satu pekerja. Gunakan ID tersebut untuk membaca nomor pekerja dan nama tampil dengan izin HRM serta EF. Daftar dan payload mutation tidak melakukan lookup hiring manager. UI dapat memuat detail setelah membuka draf atau menyimpan perubahan.

[Bukti HTTPS detail hiring manager](../../e2e/gateway/165-hrm-query-pekerja-hiring-manager-requisition.md).

Daftar requisition dapat meminta `jobsByIds` dan `organizationUnitsByIds` sekali per halaman. Cocokkan `_id` untuk menampilkan kode dan judul jabatan serta kode dan nama unit. [Bukti HTTPS](../../e2e/gateway/166-hrm-query-label-jabatan-batch.md).
