-- Pembalikan hanya untuk database sekali pakai pada pengujian migrasi.
ALTER TABLE hr_requisitions
  DROP CONSTRAINT ck_hr_requisition_rejected_only,
  DROP CONSTRAINT ck_hr_requisition_rejected,
  DROP COLUMN rejected_by_principal_id,
  DROP COLUMN rejection_reason_code,
  DROP COLUMN rejected_at;
