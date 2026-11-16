package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/career"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestCareerTarget(t *testing.T) {
	t.Parallel()
	approver, approvedAt := f.ID(3), f.Date()
	path := entity.CareerPath{ID: f.ID(1), PathCode: "RETAIL", PathVersion: 1, PathName: "Retail", JobFamilyID: f.ID(2),
		ValidFrom: f.Date(), Status: "approved", CreatedByPrincipalID: f.ID(10), CreatedAt: f.Date(),
		ApprovedByPrincipalID: &approver, ApprovedAt: &approvedAt}
	transition := entity.CareerPathTransition{ID: f.ID(4), CareerPathID: path.ID, FromJobID: f.ID(5), ToJobID: f.ID(6), MovementType: "promotion", RequirementsDocumentVersionReferenceID: f.ID(7), RequirementsDigest: []byte{1}}
	aspiration := entity.WorkerCareerAspiration{ID: f.ID(8), WorkerID: f.ID(9), TargetJobID: transition.ToJobID, CareerPathTransitionID: &transition.ID, PreferenceRank: 1, RecordedAt: f.Date(), RecordedByPrincipalID: f.ID(3), MobilityPreference: "unspecified"}
	if err := service.ValidateAspiration(&aspiration, &transition, &path); err != nil {
		t.Fatal(err)
	}
	aspiration.TargetJobID = transition.FromJobID
	if err := service.ValidateAspiration(&aspiration, &transition, &path); err == nil {
		t.Fatal("wrong career target accepted")
	}
}
