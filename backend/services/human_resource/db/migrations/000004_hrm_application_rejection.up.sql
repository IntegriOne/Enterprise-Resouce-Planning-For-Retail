-- Wajibkan alasan dan actor keputusan untuk lamaran yang ditolak.
ALTER TABLE hr_applications
  ADD CONSTRAINT ck_hr_application_rejected
    CHECK (status <> 'rejected' OR
      (current_stage_code = 'rejected' AND disposition_reason_code IS NOT NULL AND
        btrim(disposition_reason_code) <> '' AND decision_at IS NOT NULL AND
        decision_at >= applied_at AND decision_by_principal_id IS NOT NULL AND
        hired_worker_id IS NULL));
