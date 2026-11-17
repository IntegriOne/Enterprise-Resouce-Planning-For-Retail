DROP INDEX IF EXISTS ix_hr_assignment_requester;

ALTER TABLE hr_worker_assignments
  DROP CONSTRAINT ck_hr_assignment_approval_provenance,
  DROP CONSTRAINT ck_hr_assignment_approval_pair,
  DROP CONSTRAINT ck_hr_assignment_request_pair,
  DROP COLUMN requested_at,
  DROP COLUMN requested_by_principal_id;
