-- Simpan keputusan penolakan requisition yang telah diajukan.
ALTER TABLE hr_requisitions
  ADD COLUMN rejected_at timestamptz,
  ADD COLUMN rejection_reason_code varchar(80),
  ADD COLUMN rejected_by_principal_id uuid,
  ADD CONSTRAINT ck_hr_requisition_rejected
    CHECK (status <> 'rejected' OR
      (created_by_principal_id IS NOT NULL AND submitted_at IS NOT NULL AND
       submitted_by_principal_id IS NOT NULL AND rejected_at IS NOT NULL AND
       rejected_at >= submitted_at AND rejection_reason_code IS NOT NULL AND
       btrim(rejection_reason_code) <> '' AND rejected_by_principal_id IS NOT NULL AND
       rejected_by_principal_id <> created_by_principal_id AND
       rejected_by_principal_id <> submitted_by_principal_id AND
       budget_reference_id IS NULL AND approved_at IS NULL AND
       approved_by_principal_id IS NULL AND opened_at IS NULL AND closed_at IS NULL)),
  ADD CONSTRAINT ck_hr_requisition_rejected_only
    CHECK (status = 'rejected' OR
      (rejected_at IS NULL AND rejection_reason_code IS NULL AND rejected_by_principal_id IS NULL));
