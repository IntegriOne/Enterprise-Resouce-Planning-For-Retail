-- Simpan pengaju draf penempatan untuk pemisahan tugas saat persetujuan.
ALTER TABLE hr_worker_assignments
  ADD COLUMN requested_by_principal_id uuid,
  ADD COLUMN requested_at timestamptz,
  ADD CONSTRAINT ck_hr_assignment_request_pair
    CHECK ((requested_by_principal_id IS NULL) = (requested_at IS NULL)),
  ADD CONSTRAINT ck_hr_assignment_approval_pair
    CHECK ((approved_by_principal_id IS NULL) = (approved_at IS NULL)),
  ADD CONSTRAINT ck_hr_assignment_approval_provenance
    CHECK (approved_at IS NULL OR
      (requested_by_principal_id IS NOT NULL AND requested_at IS NOT NULL AND
       approved_by_principal_id <> requested_by_principal_id AND approved_at >= requested_at));

CREATE INDEX ix_hr_assignment_requester
  ON hr_worker_assignments (requested_by_principal_id, requested_at);
