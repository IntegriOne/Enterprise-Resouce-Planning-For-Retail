package human_resource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FreezePayrollInput membekukan input payroll periode terbuka dan mengembalikan metadata snapshot tanpa isi sensitif.
func (a *Adapter) FreezePayrollInput(ctx context.Context, input model.FreezePayrollInputInput) (*model.FrozenPayrollInput, error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, "human_resource.payroll_runs.create")
	if err != nil {
		return nil, err
	}
	periodID, err := parseNodeID(input.PayrollPeriodID, "payroll-period")
	if err != nil || periodID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll period ID", nil)
	}
	idemKey, err := uuid.Parse(strings.TrimSpace(input.IdempotencyKey))
	if err != nil || idemKey == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid idempotency key", nil)
	}
	response, err := a.payroll.FreezePayrollInput(rpcCtx, &payroll.FreezePayrollInputRequest{
		PayrollPeriodId: periodID.String(),
		IdempotencyKey:  idemKey.String(),
	})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("payroll period cannot be frozen", err)
	}
	if err != nil {
		return nil, mapReadError(err, "payroll input snapshot")
	}
	item := response.GetItem()
	for _, raw := range []string{
		item.GetId(),
		item.GetPayrollPeriodId(),
		item.GetPayrollGroupId(),
		item.GetEmployingLegalEntityReferenceId(),
	} {
		if id, parseErr := uuid.Parse(raw); parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll input snapshot response", nil)
		}
	}
	if !strings.EqualFold(item.GetPayrollPeriodId(), periodID.String()) ||
		strings.TrimSpace(item.GetJurisdictionCode()) == "" ||
		strings.TrimSpace(item.GetCurrencyCode()) == "" ||
		len(item.GetDigest()) != sha256.Size ||
		item.GetFrozenAt().CheckValid() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource payroll input snapshot response", nil)
	}
	return &model.FrozenPayrollInput{
		InputSnapshotReferenceID:        item.GetId(),
		PayrollPeriodID:                 item.GetPayrollPeriodId(),
		PayrollGroupID:                  item.GetPayrollGroupId(),
		EmployingLegalEntityReferenceID: item.GetEmployingLegalEntityReferenceId(),
		JurisdictionCode:                item.GetJurisdictionCode(),
		CurrencyCode:                    item.GetCurrencyCode(),
		InputDigestSha256:               hex.EncodeToString(item.GetDigest()),
		FrozenAt:                        item.GetFrozenAt().AsTime(),
	}, nil
}

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
