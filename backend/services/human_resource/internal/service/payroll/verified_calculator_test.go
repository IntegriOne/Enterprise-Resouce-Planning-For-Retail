package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

type calculationContentFixture struct {
	snapshot, artifact []byte
	artifactCalls      int
	engineCalls        int
}

func (f *calculationContentFixture) ReadCalculationSnapshot(context.Context, string) ([]byte, error) {
	return f.snapshot, nil
}

func (f *calculationContentFixture) ReadRuleArtifactVersion(context.Context, string) ([]byte, error) {
	f.artifactCalls++
	return f.artifact, nil
}

func (f *calculationContentFixture) CalculateVerified(_ context.Context, req CalculationRequest, snapshot, artifact []byte) (CalculationResult, error) {
	f.engineCalls++
	if !bytes.Equal(snapshot, f.snapshot) || !bytes.Equal(artifact, f.artifact) {
		return CalculationResult{}, errors.New("engine inputs do not match provider bytes")
	}
	return CalculationResult{InputSnapshotID: req.InputSnapshotID}, nil
}

func TestVerifiedCalculatorRejectsMismatchedProviderBytes(t *testing.T) {
	t.Parallel()
	snapshot, artifact := []byte("frozen payroll input"), []byte("approved rule artifact")
	inputDigest, ruleDigest := sha256.Sum256(snapshot), sha256.Sum256(artifact)
	request := CalculationRequest{
		InputSnapshotID:       "00000000-0000-4000-8000-000000000001",
		RuleArtifactVersionID: "00000000-0000-4000-8000-000000000002",
		InputDigest:           inputDigest[:], RuleArtifactDigest: ruleDigest[:],
	}
	for _, tc := range []struct {
		name                       string
		snapshot, artifact         []byte
		artifactCalls, engineCalls int
		want                       error
	}{
		{"valid", snapshot, artifact, 1, 1, nil},
		{"changed snapshot", []byte("changed input"), artifact, 0, 0, hrerrors.ErrReferenceConflict},
		{"changed artifact", snapshot, []byte("changed artifact"), 1, 0, hrerrors.ErrReferenceConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &calculationContentFixture{snapshot: tc.snapshot, artifact: tc.artifact}
			calculator, err := NewVerifiedCalculator(f, f, f)
			if err != nil {
				t.Fatal(err)
			}
			result, err := calculator.Calculate(t.Context(), request)
			if !errors.Is(err, tc.want) || f.artifactCalls != tc.artifactCalls || f.engineCalls != tc.engineCalls {
				t.Fatalf("calculation: error=%v artifact calls=%d engine calls=%d", err, f.artifactCalls, f.engineCalls)
			}
			if tc.want == nil && result.InputSnapshotID != request.InputSnapshotID {
				t.Fatal("verified engine result was not returned")
			}
		})
	}
}
