package service

import (
	"bytes"
	"context"
	"crypto/sha256"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type CalculationSnapshotReader interface {
	ReadCalculationSnapshot(ctx context.Context, id string) ([]byte, error)
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
	snapshot = bytes.Clone(snapshot)
	inputDigest := sha256.Sum256(snapshot)
	if len(snapshot) == 0 || !bytes.Equal(inputDigest[:], request.InputDigest) {
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
	return c.engine.CalculateVerified(ctx, request, snapshot, artifact)
}
