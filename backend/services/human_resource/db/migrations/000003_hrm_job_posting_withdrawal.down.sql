-- Pembalikan hanya untuk database sekali pakai pada pengujian migrasi.
ALTER TABLE hr_job_postings
  DROP CONSTRAINT ck_hr_posting_withdrawn_only,
  DROP CONSTRAINT ck_hr_posting_withdrawn,
  DROP COLUMN withdrawn_by_principal_id,
  DROP COLUMN withdrawal_reason_code,
  DROP COLUMN withdrawn_at;
