package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setupTalentAll(t *testing.T) (
	*mockdb.Harness, *service.Service,
	*mockdb.Store[entity.TalentFramework, string], *mockdb.Store[entity.TalentDimension, string],
	*mockdb.Store[entity.TalentReviewCycle, string], *mockdb.Store[entity.TalentAssessment, string],
	*mockdb.Store[entity.TalentAssessmentScore, string], *mockdb.Store[entity.TalentCalibrationSession, string],
	*mockdb.Store[entity.TalentCalibrationDecision, string], *mockdb.Store[entity.TalentPool, string],
	*mockdb.Store[entity.TalentPoolMembership, string],
) {
	t.Helper()
	h := mockdb.New()
	fw := mockdb.NewStore[entity.TalentFramework, string](h, func(v *entity.TalentFramework) string { return v.ID })
	dim := mockdb.NewStore[entity.TalentDimension, string](h, func(v *entity.TalentDimension) string { return v.ID })
	cyc := mockdb.NewStore[entity.TalentReviewCycle, string](h, func(v *entity.TalentReviewCycle) string { return v.ID })
	ass := mockdb.NewStore[entity.TalentAssessment, string](h, func(v *entity.TalentAssessment) string { return v.ID })
	sco := mockdb.NewStore[entity.TalentAssessmentScore, string](h, func(v *entity.TalentAssessmentScore) string { return v.ID })
	ses := mockdb.NewStore[entity.TalentCalibrationSession, string](h, func(v *entity.TalentCalibrationSession) string { return v.ID })
	dec := mockdb.NewStore[entity.TalentCalibrationDecision, string](h, func(v *entity.TalentCalibrationDecision) string { return v.ID })
	pool := mockdb.NewStore[entity.TalentPool, string](h, func(v *entity.TalentPool) string { return v.ID })
	mem := mockdb.NewStore[entity.TalentPoolMembership, string](h, func(v *entity.TalentPoolMembership) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{
		Runtime: h.Runtime(), TalentFramework: fw, TalentDimension: dim, TalentReviewCycle: cyc,
		TalentAssessment: ass, TalentAssessmentScore: sco, TalentCalibrationSession: ses,
		TalentCalibrationDecision: dec, TalentPool: pool, TalentPoolMembership: mem,
	})
	if err != nil {
		t.Fatal(err)
	}
	return h, svc, fw, dim, cyc, ass, sco, ses, dec, pool, mem
}

func TestTalentCycleAssessmentCalibrationAndPoolWorkflows(t *testing.T) {
	t.Parallel()
	h, svc, fw, dim, cyc, ass, sco, ses, dec, pool, mem := setupTalentAll(t)
	framework := fixture.Framework()
	dimension := fixture.Dimension()
	fw.Rows[framework.ID] = framework
	dim.Rows[dimension.ID] = dimension

	cycle := fixture.Cycle()
	cycle.Status = "draft"
	cyc.Rows[cycle.ID] = cycle

	if _, err := svc.StartTalentReviewCycleCalibration(t.Context(), cycle.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("calibration before launch = %v, want ErrWorkflowRequired", err)
	}
	launched, err := svc.LaunchTalentReviewCycle(t.Context(), cycle.ID)
	if err != nil || launched.Status != "open" {
		t.Fatalf("launch cycle: %+v, err=%v", launched, err)
	}

	// Assessor mengajukan lalu finalisasi asesmen oleh panelis berbeda.
	assessment := fixture.Assessment()
	assessment.AssessedByPrincipalID = h.Actor.PrincipalID
	ass.Rows[assessment.ID] = assessment
	score := fixture.Score()
	sco.Rows[score.ID] = score

	submitted, err := svc.SubmitTalentAssessment(t.Context(), assessment.ID)
	if err != nil || submitted.Status != "submitted" {
		t.Fatalf("submit assessment: %+v, err=%v", submitted, err)
	}
	calibrator := fixture.ID(99)
	h.Actor = operation.Actor{PrincipalID: calibrator, SessionID: fixture.ID(90), Kind: "human"}
	finalized, err := svc.FinalizeTalentAssessment(t.Context(), assessment.ID)
	if err != nil || finalized.Status != "finalized" {
		t.Fatalf("finalize submitted assessment: %+v, err=%v", finalized, err)
	}

	// Penarikan asesmen yang sudah final wajib ditolak.
	if _, err := svc.WithdrawTalentAssessment(t.Context(), assessment.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("withdraw finalized = %v, want ErrWorkflowRequired", err)
	}

	// Buka sesi kalibrasi dan catat keputusan kalibrasi.
	session := entity.TalentCalibrationSession{
		ID: fixture.ID(21), TalentReviewCycleID: cycle.ID, SessionCode: "CAL-1",
		ChairPrincipalID: calibrator, ScheduledAt: fixture.Date(), Status: "planned",
	}
	ses.Rows[session.ID] = session
	if _, err := svc.OpenTalentCalibrationSession(t.Context(), session.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("open session before cycle calibration = %v", err)
	}
	if _, err := svc.StartTalentReviewCycleCalibration(t.Context(), cycle.ID); err != nil {
		t.Fatalf("start calibration: %v", err)
	}
	openedSession, err := svc.OpenTalentCalibrationSession(t.Context(), session.ID)
	if err != nil || openedSession.Status != "open" {
		t.Fatalf("open session: %+v, err=%v", openedSession, err)
	}
	recorded, err := svc.RecordTalentCalibrationDecision(t.Context(), service.RecordTalentCalibrationDecisionInput{
		SessionID: session.ID, AssessmentID: assessment.ID, DimensionID: dimension.ID,
		CalibratedScore: 5, RationaleDocumentVersionReferenceID: fixture.ID(97), RationaleDigest: []byte{1},
	})
	if err != nil || recorded.OriginalScore != 4 || recorded.CalibratedScore != 5 || len(dec.Rows) != 1 {
		t.Fatalf("record calibration decision: %+v, err=%v", recorded, err)
	}

	// Nominasi, aktivasi, dan exit keanggotaan talent pool.
	talentPool := entity.TalentPool{
		ID: fixture.ID(31), PoolCode: "HIPO", PoolVersion: 1, PoolName: "HiPo",
		LegalEntityReferenceID: fixture.ID(91), EligibilityDocumentVersionReferenceID: fixture.ID(97),
		EligibilityDigest: []byte{1}, ValidFrom: fixture.Date().AddDate(0, -1, 0), OwnerPrincipalID: calibrator,
	}
	pool.Rows[talentPool.ID] = talentPool
	nominee := entity.TalentPoolMembership{
		ID: fixture.ID(32), TalentPoolID: talentPool.ID, WorkerID: assessment.WorkerID, AssessmentID: assessment.ID,
		NominatedByPrincipalID: fixture.ID(96), Status: "nominated", ValidFrom: fixture.Date().AddDate(0, -1, 0),
	}
	mem.Rows[nominee.ID] = nominee
	approved, err := svc.ApproveTalentPoolMembership(t.Context(), nominee.ID)
	if err != nil || approved.Status != "active" || approved.ApprovedByPrincipalID == nil {
		t.Fatalf("approve pool membership: %+v, err=%v", approved, err)
	}
	nowExited, err := svc.ExitTalentPoolMembership(t.Context(), nominee.ID, "PROMOTED")
	if err != nil || nowExited.Status != "exited" || nowExited.ValidTo == nil {
		t.Fatalf("exit pool membership: %+v, err=%v", nowExited, err)
	}

	// Tutup siklus ketika seluruh sesi sudah selesai atau dibatalkan.
	closedSession := ses.Rows[session.ID]
	closedSession.Status = "cancelled"
	ses.Rows[session.ID] = closedSession
	closedCycle, err := svc.CloseTalentReviewCycle(t.Context(), cycle.ID)
	if err != nil || closedCycle.Status != "closed" || closedCycle.ClosedAt == nil {
		t.Fatalf("close cycle: %+v, err=%v", closedCycle, err)
	}
	_ = time.UTC
}
