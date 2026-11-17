-- Simpan alasan, actor, dan waktu penarikan posting secara atomik.
ALTER TABLE hr_job_postings
  ADD COLUMN withdrawn_at timestamptz,
  ADD COLUMN withdrawal_reason_code varchar(80),
  ADD COLUMN withdrawn_by_principal_id uuid,
  ADD CONSTRAINT ck_hr_posting_withdrawn
    CHECK (status <> 'withdrawn' OR
      (published_at IS NOT NULL AND withdrawn_at IS NOT NULL AND
        withdrawn_at > published_at AND closes_at IS NULL AND
        withdrawal_reason_code IS NOT NULL AND btrim(withdrawal_reason_code) <> '' AND
        withdrawn_by_principal_id IS NOT NULL)),
  ADD CONSTRAINT ck_hr_posting_withdrawn_only
    CHECK (status = 'withdrawn' OR
      (withdrawn_at IS NULL AND withdrawal_reason_code IS NULL AND withdrawn_by_principal_id IS NULL));
