package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type candidateReaderStub struct{ rows []entity.Candidate }

func (s *candidateReaderStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Candidate, error) {
	if len(filters) != 1 || filters[0].Column != "candidate_number" || limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

type candidateApplicationStub struct{ count int64 }

func (s *candidateApplicationStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "candidate_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestCandidateValidator(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)
	ack := now.Add(-time.Hour)
	retention := now.AddDate(0, 1, 0)
	actor := "00000000-0000-4000-8000-000000000001"
	name := "Candidate A"
	candidate := &entity.Candidate{ID: "00000000-0000-4000-8000-000000000002", CandidateNumber: "CAND-1",
		PreferredName: &name, PrivacyNoticeVersion: "v1", NoticeAcknowledgedAt: &ack,
		DataRetentionUntil: &retention, CandidateStatus: "active", CreatedAt: now, CreatedByPrincipalID: &actor}
	reader := &candidateReaderStub{}
	applications := &candidateApplicationStub{}
	v := &candidateValidator{candidates: reader, applications: applications}
	if err := v.ValidateMutation(t.Context(), "hr_candidates", nil, candidate); err != nil {
		t.Fatalf("valid local candidate: %v", err)
	}
	reader.rows = []entity.Candidate{*candidate}
	if err := v.ValidateMutation(t.Context(), "hr_candidates", nil, candidate); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate number: %v", err)
	}
	reader.rows = nil
	changed := *candidate
	changed.NoticeAcknowledgedAt = nil
	if err := v.ValidateMutation(t.Context(), "hr_candidates", nil, &changed); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("missing notice acknowledgement: %v", err)
	}
	changed = *candidate
	shortRetention := now
	changed.DataRetentionUntil = &shortRetention
	if err := v.ValidateMutation(t.Context(), "hr_candidates", nil, &changed); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("invalid retention: %v", err)
	}
	changed = *candidate
	changed.ContactReferenceID = &actor
	if err := v.ValidateMutation(t.Context(), "hr_candidates", nil, &changed); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified Foundation contact: %v", err)
	}
	changed = *candidate
	changed.PrivacyNoticeVersion = "v2"
	if err := v.ValidateMutation(t.Context(), "hr_candidates", candidate, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("notice version changed through CRUD: %v", err)
	}
	changed = *candidate
	corrected := "Corrected name"
	changed.PreferredName = &corrected
	if err := v.ValidateMutation(t.Context(), "hr_candidates", candidate, &changed); err != nil {
		t.Fatalf("name correction: %v", err)
	}
	applications.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_candidates", candidate, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("candidate referenced by application: %v", err)
	}
	applications.count = 0
	if err := v.ValidateMutation(t.Context(), "hr_candidates", candidate, nil); err != nil {
		t.Fatalf("unreferenced candidate delete: %v", err)
	}
}
