-- Simpan identitas pembuat dan pengusul tanpa mengubah baseline terpasang.
ALTER TABLE hr_requisitions
  ADD COLUMN created_by_principal_id uuid,
  ADD COLUMN submitted_at timestamptz,
  ADD COLUMN submitted_by_principal_id uuid,
  ADD CONSTRAINT ck_hr_requisition_submission_pair
    CHECK ((submitted_at IS NULL) = (submitted_by_principal_id IS NULL)),
  ADD CONSTRAINT ck_hr_requisition_approval_separation
    CHECK ((approved_by_principal_id IS NULL OR created_by_principal_id IS NULL OR
      approved_by_principal_id <> created_by_principal_id) AND
      (approved_by_principal_id IS NULL OR submitted_by_principal_id IS NULL OR
      approved_by_principal_id <> submitted_by_principal_id));
