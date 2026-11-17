-- Pastikan shortlist manual memiliki recruiter dan tidak meniru skor terverifikasi.
ALTER TABLE hr_applications DROP CONSTRAINT ck_hr_application_rejected;
ALTER TABLE hr_applications
  ADD CONSTRAINT ck_hr_application_rejected
    CHECK (status <> 'rejected' OR
      (current_stage_code IS NOT NULL AND current_stage_code = 'rejected' AND
        disposition_reason_code IS NOT NULL AND btrim(disposition_reason_code) <> '' AND
        decision_at IS NOT NULL AND decision_at >= applied_at AND
        decision_by_principal_id IS NOT NULL AND hired_worker_id IS NULL));

ALTER TABLE hr_applications
  ADD CONSTRAINT ck_hr_application_shortlisted
    CHECK (status <> 'shortlisted' OR
      (current_stage_code IS NOT NULL AND current_stage_code = 'shortlisted' AND assigned_recruiter_principal_id IS NOT NULL AND
        screening_score IS NULL AND scorecard_version IS NULL AND disposition_reason_code IS NULL AND
        decision_at IS NULL AND decision_by_principal_id IS NULL AND hired_worker_id IS NULL));
