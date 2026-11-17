-- Pembalikan hanya untuk database sekali pakai pada pengujian migrasi.
ALTER TABLE hr_requisitions
  DROP CONSTRAINT ck_hr_requisition_approval_separation,
  DROP CONSTRAINT ck_hr_requisition_submission_pair,
  DROP COLUMN submitted_by_principal_id,
  DROP COLUMN submitted_at,
  DROP COLUMN created_by_principal_id;
