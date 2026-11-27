package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CalculationSnapshot memuat isi immutable dan cakupan yang disimpan oleh penerbit snapshot.
type CalculationSnapshot = entity.PayrollInputSnapshot

type CalculationSnapshotReader interface {
	ReadCalculationSnapshot(ctx context.Context, id string) (CalculationSnapshot, error)
}

type RuleArtifactReader interface {
	ReadRuleArtifactVersion(ctx context.Context, versionID string) ([]byte, error)
}

type CalculationEngine interface {
	CalculateVerified(ctx context.Context, request CalculationRequest, snapshot, ruleArtifact []byte) (CalculationResult, error)
}

type verifiedCalculator struct {
	snapshots CalculationSnapshotReader
	rules     RuleArtifactReader
	engine    CalculationEngine
}

// NewVerifiedCalculator mengikat mesin kalkulasi ke byte dari penyedia snapshot dan versi aturan.
func NewVerifiedCalculator(snapshots CalculationSnapshotReader, rules RuleArtifactReader, engine CalculationEngine) (Calculator, error) {
	for _, dependency := range []any{snapshots, rules, engine} {
		if err := operation.RequireDependency(dependency); err != nil {
			return nil, err
		}
	}
	return verifiedCalculator{snapshots: snapshots, rules: rules, engine: engine}, nil
}

func (c verifiedCalculator) Calculate(ctx context.Context, request CalculationRequest) (CalculationResult, error) {
	if entity.ValidateUUID("input_snapshot_id", request.InputSnapshotID) != nil ||
		entity.ValidateUUID("rule_artifact_version_id", request.RuleArtifactVersionID) != nil ||
		len(request.InputDigest) != sha256.Size || len(request.RuleArtifactDigest) != sha256.Size {
		return CalculationResult{}, hrerrors.ErrReferenceConflict
	}
	snapshot, err := c.snapshots.ReadCalculationSnapshot(ctx, request.InputSnapshotID)
	if err != nil {
		return CalculationResult{}, err
	}
	if !strings.EqualFold(snapshot.ID, request.InputSnapshotID) ||
		entity.ValidateUUID("payroll_period_id", snapshot.PeriodID) != nil || !strings.EqualFold(snapshot.PeriodID, request.PeriodID) ||
		entity.ValidateUUID("payroll_group_id", snapshot.GroupID) != nil || !strings.EqualFold(snapshot.GroupID, request.GroupID) ||
		entity.ValidateUUID("employing_legal_entity_reference_id", snapshot.EmployerID) != nil || !strings.EqualFold(snapshot.EmployerID, request.EmployerID) {
		return CalculationResult{}, hrerrors.ErrReferenceConflict
	}
	content := bytes.Clone(snapshot.Content)
	inputDigest := sha256.Sum256(content)
	if len(content) == 0 || !bytes.Equal(inputDigest[:], request.InputDigest) {
		return CalculationResult{}, hrerrors.ErrReferenceConflict
	}
	artifact, err := c.rules.ReadRuleArtifactVersion(ctx, request.RuleArtifactVersionID)
	if err != nil {
		return CalculationResult{}, err
	}
	artifact = bytes.Clone(artifact)
	ruleDigest := sha256.Sum256(artifact)
	if len(artifact) == 0 || !bytes.Equal(ruleDigest[:], request.RuleArtifactDigest) {
		return CalculationResult{}, hrerrors.ErrReferenceConflict
	}
	if err := ctx.Err(); err != nil {
		return CalculationResult{}, err
	}
	request.InputDigest = bytes.Clone(request.InputDigest)
	request.RuleArtifactDigest = bytes.Clone(request.RuleArtifactDigest)
	return c.engine.CalculateVerified(ctx, request, content, artifact)
}
