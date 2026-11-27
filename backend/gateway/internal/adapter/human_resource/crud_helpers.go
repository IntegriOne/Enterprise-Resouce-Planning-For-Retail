package human_resource

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// authorizeResourceRead menegakkan izin baca (termasuk step-up bila resource sensitif atau scope legal_entity bila didukung).
func (a *Adapter) authorizeResourceRead(ctx context.Context, table, permission string) (context.Context, error) {
	if table == "hr_workers" || table == "hr_workforce_plans" || table == "hr_workforce_plan_lines" {
		if security.IsSensitiveOperation(table, "read") {
			return a.authorizeSensitiveEmployee(ctx, permission)
		}
		return a.authorizeRead(ctx, permission)
	}
	if security.IsSensitiveOperation(table, "read") {
		return a.authorizeSensitiveEmployee(ctx, permission)
	}
	return a.authorizeGlobalEmployee(ctx, permission)
}

// authorizeResourceMutation menegakkan izin mutasi dan MFA sesuai sensitivitas operasi HRM.
func (a *Adapter) authorizeResourceMutation(ctx context.Context, table, action, permission string) (context.Context, error) {
	if security.IsSensitiveOperation(table, action) {
		return a.authorizeSensitiveEmployee(ctx, permission)
	}
	if a == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	session, err := a.authorizer.RequirePermission(ctx, permission, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	if err != nil {
		return nil, err
	}
	if session.Caller.UserKind != "employee" {
		return nil, middleware.ErrForbidden("employee session required", nil)
	}
	return middleware.WithResolvedSession(ctx, session), nil
}

func mapHRMWriteError(err error, resource string) error {
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return middleware.ErrConflict(resource+" conflicts with existing state", err)
	case codes.OK:
		return nil
	default:
		return mapReadError(err, resource)
	}
}

func parsePositiveInt64NodeID(rawID, resource string) (int64, error) {
	raw := rawNodeID(strings.TrimSpace(rawID), resource)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, middleware.ErrInvalidArgument("invalid "+resource+" ID", nil)
	}
	return id, nil
}

func int64NodeID(resource string, id int64) string {
	return fmt.Sprintf("%s%s/%d", nodePrefix, resource, id)
}

func validateFilterString(raw *string, field string) (string, error) {
	if raw == nil {
		return "", middleware.ErrInvalidArgument("missing "+field+" filter", nil)
	}
	v := *raw
	if !utf8.ValidString(v) || strings.ContainsRune(v, 0) || strings.TrimSpace(v) == "" || utf8.RuneCountInString(v) > 200 {
		return "", middleware.ErrInvalidArgument("invalid "+field+" filter", nil)
	}
	return v, nil
}

func validateUUIDFilter(raw *string, resource, field string) (string, error) {
	if raw == nil {
		return "", middleware.ErrInvalidArgument("missing "+field+" filter", nil)
	}
	id, err := parseNodeID(strings.TrimSpace(*raw), resource)
	if err != nil || id == uuid.Nil {
		return "", middleware.ErrInvalidArgument("invalid "+field+" ID", nil)
	}
	return id.String(), nil
}

func normalizeOptionalUUID(raw *string, resource, field string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	id, err := parseNodeID(strings.TrimSpace(*raw), resource)
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid "+field+" ID", nil)
	}
	s := id.String()
	return &s, nil
}

func normalizeRequiredUUID(raw string, resource, field string) (string, error) {
	id, err := parseNodeID(strings.TrimSpace(raw), resource)
	if err != nil || id == uuid.Nil {
		return "", middleware.ErrInvalidArgument("invalid "+field+" ID", nil)
	}
	return id.String(), nil
}

func pickSingleFilter(setFlags ...bool) (int, error) {
	chosen := -1
	for i, isSet := range setFlags {
		if !isSet {
			continue
		}
		if chosen >= 0 {
			return -1, middleware.ErrInvalidArgument("use at most one filter argument", nil)
		}
		chosen = i
	}
	return chosen, nil
}

func mapRelayPage[PB any, Model any](page *common.PageInfo, items []PB, limit, offset int, resource string, mapper func(PB) (*Model, error)) (*Page[Model], error) {
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource "+resource+" page", nil)
	}
	out := &Page[Model]{
		Items:      make([]*Model, 0, len(items)),
		TotalCount: int(page.GetTotal()),
		Limit:      limit,
		Offset:     offset,
	}
	for _, row := range items {
		mapped, err := mapper(row)
		if err != nil {
			return nil, err
		}
		out.Items = append(out.Items, mapped)
	}
	return out, nil
}

func deleteByUUID(
	ctx context.Context,
	a *Adapter,
	rawID, table, permission, resource string,
	call func(context.Context, *common.IDRequest) (*common.EmptyResponse, error),
) (*bool, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, table, "delete", permission)
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, resource, resource)
	if err != nil {
		return nil, err
	}
	if _, err := call(rpcCtx, &common.IDRequest{Id: id}); err != nil {
		return nil, mapHRMWriteError(err, resource)
	}
	ok := true
	return &ok, nil
}

func applyStringMask(mask *fieldmaskpb.FieldMask, path string, in graphql.Omittable[*string], target **string, nullable bool) error {
	val, set := in.ValueOK()
	if !set {
		return nil
	}
	if val == nil && !nullable {
		return middleware.ErrInvalidArgument("required "+path+" cannot be null", nil)
	}
	*target = val
	mask.Paths = append(mask.Paths, path)
	return nil
}

func applyBoolMask(mask *fieldmaskpb.FieldMask, path string, in graphql.Omittable[*bool], target **bool, nullable bool) error {
	val, set := in.ValueOK()
	if !set {
		return nil
	}
	if val == nil && !nullable {
		return middleware.ErrInvalidArgument("required "+path+" cannot be null", nil)
	}
	*target = val
	mask.Paths = append(mask.Paths, path)
	return nil
}

func applyInt32Mask(mask *fieldmaskpb.FieldMask, path string, in graphql.Omittable[*int], target **int32, nullable bool) error {
	val, set := in.ValueOK()
	if !set {
		return nil
	}
	if val == nil {
		if !nullable {
			return middleware.ErrInvalidArgument("required "+path+" cannot be null", nil)
		}
		*target = nil
		mask.Paths = append(mask.Paths, path)
		return nil
	}
	v := int32(*val)
	*target = &v
	mask.Paths = append(mask.Paths, path)
	return nil
}

func applyInt64Mask(mask *fieldmaskpb.FieldMask, path string, in graphql.Omittable[*string], target **int64, nullable bool) error {
	val, set := in.ValueOK()
	if !set {
		return nil
	}
	if val == nil {
		if !nullable {
			return middleware.ErrInvalidArgument("required "+path+" cannot be null", nil)
		}
		*target = nil
		mask.Paths = append(mask.Paths, path)
		return nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(*val), 10, 64)
	if err != nil {
		return middleware.ErrInvalidArgument("invalid "+path, nil)
	}
	*target = &parsed
	mask.Paths = append(mask.Paths, path)
	return nil
}

func applyTimestampMask(mask *fieldmaskpb.FieldMask, path string, in graphql.Omittable[*time.Time], target **timestamppb.Timestamp, nullable bool) error {
	val, set := in.ValueOK()
	if !set {
		return nil
	}
	if val == nil {
		if !nullable {
			return middleware.ErrInvalidArgument("required "+path+" cannot be null", nil)
		}
		*target = nil
		mask.Paths = append(mask.Paths, path)
		return nil
	}
	*target = timestamppb.New(*val)
	mask.Paths = append(mask.Paths, path)
	return nil
}

func applyBytesMask(mask *fieldmaskpb.FieldMask, path string, in graphql.Omittable[*string], target *[]byte, nullable bool) error {
	val, set := in.ValueOK()
	if !set {
		return nil
	}
	if val == nil {
		if !nullable {
			return middleware.ErrInvalidArgument("required "+path+" cannot be null", nil)
		}
		*target = nil
		mask.Paths = append(mask.Paths, path)
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(*val))
	if err != nil {
		return middleware.ErrInvalidArgument("invalid base64 for "+path, nil)
	}
	*target = decoded
	mask.Paths = append(mask.Paths, path)
	return nil
}

func pbTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func pbIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	x := int(*v)
	return &x
}

func pbInt64StrPtr(v *int64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatInt(*v, 10)
	return &s
}

func pbJSONString(b []byte) *string {
	if len(b) == 0 {
		return nil
	}
	s := string(b)
	return &s
}

func pbBase64String(b []byte) *string {
	if len(b) == 0 {
		return nil
	}
	s := base64.StdEncoding.EncodeToString(b)
	return &s
}

func pbBase64StringValue(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func optionalTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func optionalInt32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	x := int32(*v)
	return &x
}

func optionalInt64FromStr(v *string, field string) (*int64, error) {
	if v == nil {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(*v), 10, 64)
	if err != nil {
		return nil, middleware.ErrInvalidArgument("invalid "+field, nil)
	}
	return &parsed, nil
}

func optionalJSONBytes(v *string, field string) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	raw := strings.TrimSpace(*v)
	if !json.Valid([]byte(raw)) {
		return nil, middleware.ErrInvalidArgument("invalid JSON for "+field, nil)
	}
	return []byte(raw), nil
}

func requiredJSONBytes(v string, field string) ([]byte, error) {
	raw := strings.TrimSpace(v)
	if !json.Valid([]byte(raw)) {
		return nil, middleware.ErrInvalidArgument("invalid JSON for "+field, nil)
	}
	return []byte(raw), nil
}

func optionalBase64Bytes(v *string, field string) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(*v))
	if err != nil {
		return nil, middleware.ErrInvalidArgument("invalid base64 for "+field, nil)
	}
	return b, nil
}
