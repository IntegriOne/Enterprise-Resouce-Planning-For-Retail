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

type engSurveyStub struct {
	rows []entity.EngagementSurvey
}

func (s *engSurveyStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.EngagementSurvey, error) {
	var out []entity.EngagementSurvey
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "legal_entity_reference_id":
				if row.LegalEntityReferenceID != f.Values[0] {
					match = false
				}
			case "survey_code":
				if row.SurveyCode != f.Values[0] {
					match = false
				}
			case "survey_version":
				if row.SurveyVersion != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

type engResultCounterStub struct {
	count int64
}

func (s *engResultCounterStub) CountStoredBySurveyID(context.Context, string) (int64, error) {
	return s.count, nil
}

func TestEngagementSurveyValidator(t *testing.T) {
	t.Parallel()
	leID := "00000000-0000-4000-8000-000000000501"
	surveys := &engSurveyStub{}
	results := &engResultCounterStub{}
	validator := &engagementSurveyValidator{surveys: surveys, results: results, foundation: &erFoundationStub{}}

	now := time.Now().UTC()
	survey := &entity.EngagementSurvey{
		ID: "surv-1", SurveyCode: "ENG-27", SurveyVersion: 1, LegalEntityReferenceID: leID, TopicCode: "engagement",
		OpensAt: now, ClosesAt: now.Add(24 * time.Hour), MinimumReportingPopulation: 5,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_engagement_surveys", nil, survey); err != nil {
		t.Fatalf("create valid engagement survey: %v", err)
	}
	lowPop := *survey
	lowPop.MinimumReportingPopulation = 4
	if err := validator.ValidateMutation(t.Context(), "hr_engagement_surveys", nil, &lowPop); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("anonymity threshold < 5 must fail: %v", err)
	}
	results.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_engagement_surveys", survey, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete survey with results must fail: %v", err)
	}
}
