package human_resource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreatePayrollRun membuat draf dari referensi snapshot dan aturan yang diperiksa HRM.
func (a *Adapter) CreatePayrollRun(ctx context.Context, input model.CreatePayrollRunInput) (*model.PayrollRun, error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, "human_resource.payroll_runs.create")
	if err != nil {
		return nil, err
	}
	if input.RunNumber < 1 || input.RunNumber > math.MaxInt32 {
		return nil, middleware.ErrInvalidArgument("invalid payroll run number", nil)
	}
	if len(input.InputDigestSha256) != 2*sha256.Size {
		return nil, middleware.ErrInvalidArgument("input digest must be a hexadecimal SHA-256 digest", nil)
	}
	digest, err := hex.DecodeString(input.InputDigestSha256)
	if err != nil {
		return nil, middleware.ErrInvalidArgument("input digest must be a hexadecimal SHA-256 digest", nil)
	}
	if !validIdempotencyKey(input.IdempotencyKey) {
		return nil, middleware.ErrInvalidArgument("invalid payroll idempotency key", nil)
	}
	runNumber := int32(input.RunNumber)
	response, err := a.payroll.CreatePayrollRun(rpcCtx, &payroll.CreatePayrollRunRequest{
		Input: &payroll.PayrollRunInput{
			PayrollPeriodId: optionalRawNodeID(&input.PayrollPeriodID, "payroll-period"), PayrollGroupId: optionalRawNodeID(&input.PayrollGroupID, "payroll-group"),
			EmployingLegalEntityReferenceId: optionalRawNodeID(&input.EmployingLegalEntityReferenceID, "legal-entity"),
			CurrencyCode:                    &input.CurrencyCode, RuleSetId: optionalRawNodeID(&input.RuleSetID, "payroll-rule-set"), OriginalRunId: optionalRawNodeID(input.OriginalRunID, "payroll-run"),
			SettlementDirection: input.SettlementDirection, RunNumber: &runNumber, RunType: &input.RunType,
			JurisdictionCode: &input.JurisdictionCode, RuleEngineName: &input.RuleEngineName,
			RuleEngineVersion: &input.RuleEngineVersion, InputSnapshotReferenceId: &input.InputSnapshotReferenceID,
			InputDigest: digest, IdempotencyKey: &input.IdempotencyKey,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists:
		return nil, middleware.ErrConflict("payroll draft already exists", err)
	case codes.FailedPrecondition:
		return nil, middleware.ErrConflict("payroll draft does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "payroll run")
	}
	run := response.GetItem()
	if run.GetStatus() != "draft" || run.GetRunNumber() != runNumber {
		return nil, middleware.ErrUnavailable("invalid human resource payroll draft response", nil)
	}
	return toModelPayrollRun(run)
}

func validIdempotencyKey(key string) bool {
	return utf8.ValidString(key) && !strings.ContainsRune(key, 0) &&
		strings.TrimSpace(key) != "" && utf8.RuneCountInString(key) <= 160
}
