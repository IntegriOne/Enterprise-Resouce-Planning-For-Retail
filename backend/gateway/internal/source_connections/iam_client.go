package source_connections

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// IAMRuntimeBridge menghubungkan IAM gRPC ke SnapshotLoader dan Authorizer Gateway tanpa memaparkan status mentah ke resolver.
type IAMRuntimeBridge struct {
	clients IAMClientSet
	timeout time.Duration
}

// NewIAMRuntimeBridge membentuk jembatan runtime gRPC IAM untuk penyegaran snapshot dan pemeriksaan reauth sensitif.
func NewIAMRuntimeBridge(clients IAMClientSet, timeout time.Duration) *IAMRuntimeBridge {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &IAMRuntimeBridge{
		clients: clients,
		timeout: timeout,
	}
}

// RefreshUserAccessSnapshot meminta IAM menghitung serta memproyeksikan AccessSnapshot terbaru dan mengonversinya ke model authz.
func (b *IAMRuntimeBridge) RefreshUserAccessSnapshot(
	ctx context.Context,
	userID string,
	caller authcontext.CallerContext,
) (*authz.AccessSnapshot, error) {
	if b == nil || b.clients.Sessions == nil {
		return nil, middleware.ErrUnavailable("IAM session client unavailable", nil)
	}
	rpcCtx, cancel := context.WithTimeout(authcontext.WithCaller(ctx, caller), b.timeout)
	defer cancel()

	resp, err := b.clients.Sessions.RefreshUserAccessSnapshot(rpcCtx, &iamv1.RefreshUserAccessSnapshotRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, MapIAMGRPCError(err)
	}
	return MapProtoAccessSnapshot(resp), nil
}

// VerifySensitiveSession memverifikasi secara otoritatif ke IAM bahwa sesi memenuhi MFA dan jendela re-autentikasi 10 menit.
func (b *IAMRuntimeBridge) VerifySensitiveSession(
	ctx context.Context,
	rawToken string,
	requiredPermission string,
	requiredScope authz.ScopeGrant,
	caller authcontext.CallerContext,
) error {
	if b == nil || b.clients.Sessions == nil {
		return middleware.ErrUnavailable("IAM session client unavailable", nil)
	}
	rpcCtx, cancel := context.WithTimeout(authcontext.WithCaller(ctx, caller), b.timeout)
	defer cancel()

	var reqScope *iamv1.Scope
	if strings.TrimSpace(requiredPermission) != "" {
		var scopeID *string
		if strings.TrimSpace(requiredScope.ID) != "" {
			v := strings.TrimSpace(requiredScope.ID)
			scopeID = &v
		}
		reqScope = &iamv1.Scope{Type: ToProtoScopeType(requiredScope.Type), Id: scopeID}
	}

	_, err := b.clients.Sessions.AuthorizeSession(rpcCtx, &iamv1.AuthorizeSessionRequest{
		SessionToken:           rawToken,
		RequireCsrf:            false,
		RequiredPermission:     requiredPermission,
		RequiredScope:          reqScope,
		RequireSensitiveReauth: true,
	})
	if err != nil {
		return MapIAMGRPCError(err)
	}
	return nil
}

// MapProtoAccessSnapshot mengonversi keluaran gRPC RefreshUserAccessSnapshot menjadi authz.AccessSnapshot yang siap dipakai Gateway.
func MapProtoAccessSnapshot(resp *iamv1.RefreshUserAccessSnapshotResponse) *authz.AccessSnapshot {
	if resp == nil || resp.GetSnapshot() == nil {
		return nil
	}
	src := resp.GetSnapshot()
	now := time.Now()
	if src.GetCheckedAt() != nil && src.GetCheckedAt().IsValid() {
		now = src.GetCheckedAt().AsTime().Local()
	}

	roles := make([]authz.RoleGrant, 0, len(src.GetRoles()))
	for _, r := range src.GetRoles() {
		roles = append(roles, authz.RoleGrant{
			ID:                  r.GetId(),
			Code:                r.GetCode(),
			Name:                r.GetName(),
			Priority:            r.GetRedirectPriority(),
			DefaultRedirectPath: r.GetDefaultRedirectPath(),
		})
	}

	perms := make([]authz.PermissionGrant, 0, len(src.GetPermissions()))
	for _, p := range src.GetPermissions() {
		scope := authz.ScopeGrant{Type: authz.ScopeGlobal}
		if p.GetScope() != nil {
			scope = authz.ScopeGrant{
				Type: FromProtoScopeTypeString(p.GetScope().GetType()),
				ID:   p.GetScope().GetId(),
			}
		}
		perms = append(perms, authz.PermissionGrant{
			Code:        p.GetPermissionCode(),
			IsSensitive: p.GetIsSensitive(),
			Scope:       scope,
		})
	}

	return &authz.AccessSnapshot{
		PrincipalID:          src.GetUserId(),
		UserKind:             FromProtoUserKindString(src.GetUserKind()),
		Status:               FromProtoUserStatusString(src.GetUserStatus()),
		MFAEnabled:           src.GetMfaEnabled(),
		AuthEpoch:            src.GetAuthEpoch(),
		PermRevision:         src.GetPermRevision(),
		RecoveryEpoch:        src.GetRecoveryEpoch(),
		Roles:                roles,
		Permissions:          perms,
		CanAccessIdentity:    resp.GetCanAccessIdentity(),
		DefaultRedirectPath:  resp.GetDefaultRedirectPath(),
		AllowedRedirectPaths: append([]string(nil), resp.GetAllowedRedirectPaths()...),
		SyncedAtUnix:         now.Unix(),
		FreshUntilUnix:       now.Add(authz.DefaultSnapshotMaxAge).Unix(),
	}
}

// MapIAMGRPCError menerjemahkan status gRPC IAM dan detail ErrorInfo menjadi GatewayError aman tanpa membocorkan internal backend.
func MapIAMGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := middleware.AsGatewayError(err); ok {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return middleware.ErrUnavailable("upstream identity request timed out", err)
	}

	st, ok := status.FromError(err)
	if !ok {
		return middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", err)
	}

	reason := extractErrorReason(st)
	switch reason {
	case "iam.credential.invalid", "iam.totp.invalid_code", "iam.recovery_code.invalid":
		return middleware.ErrInvalidCredentials(safeMessage(st.Message(), "invalid credentials"), err)
	case "iam.credential.locked":
		return middleware.ErrTooManyRequests("account temporarily locked due to failed attempts", err)
	case "iam.session.not_found", "iam.session.expired", "iam.session.idle_expired", "iam.session.absolute_expired", "iam.session.revoked":
		return middleware.ErrUnauthenticated("session expired or invalid", err)
	case "iam.session.snapshot_stale":
		return middleware.ErrUnavailable("authorization snapshot stale", err)
	case "iam.session.reauth_required", "iam.session.csrf_mismatch", "iam.user.inactive", "iam.user.suspended", "iam.user.disabled", "iam.identity.not_verified", "iam.credential.password_rotation_required":
		return middleware.ErrForbidden(safeMessage(st.Message(), "forbidden"), err)
	case "iam.identity.already_exists", "iam.identity.same_email", "iam.identity.same_username", "iam.identity.limit_exceeded", "iam.identity.email_already_verified", "iam.role.already_exists", "iam.group.already_exists":
		return middleware.ErrConflict(safeMessage(st.Message(), "request conflicts with current state"), err)
	case "iam.identity.cannot_remove_last", "iam.identity.cannot_remove_primary", "iam.user.deleted", "iam.role.system_protected", "iam.role.system_immutable":
		return middleware.ErrUnprocessableEntity(safeMessage(st.Message(), "request could not be processed"), err)
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return middleware.ErrInvalidArgument(safeMessage(st.Message(), "invalid input"), err)
	case codes.NotFound:
		return middleware.ErrNotFound(safeMessage(st.Message(), "resource not found"), err)
	case codes.AlreadyExists, codes.FailedPrecondition:
		return middleware.ErrConflict(safeMessage(st.Message(), "request conflicts with current state"), err)
	case codes.Unauthenticated:
		return middleware.ErrUnauthenticated("authentication required", err)
	case codes.PermissionDenied:
		return middleware.ErrForbidden(safeMessage(st.Message(), "forbidden"), err)
	case codes.ResourceExhausted:
		return middleware.ErrTooManyRequests("too many requests, please retry later", err)
	case codes.Unavailable, codes.DeadlineExceeded:
		return middleware.ErrUnavailable("identity service unavailable", err)
	default:
		return middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", err)
	}
}

func extractErrorReason(st *status.Status) string {
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			return strings.TrimSpace(info.GetReason())
		}
	}
	return ""
}

// ExtractIAMErrorReason mengekstrak kode alasan ErrorInfo dari status gRPC IAM bila tersedia.
func ExtractIAMErrorReason(err error) string {
	st, ok := status.FromError(err)
	if !ok {
		return ""
	}
	return extractErrorReason(st)
}

// ExtractIAMErrorField mengekstrak nama field pertama pada detail BadRequest dari status gRPC IAM bila tersedia.
func ExtractIAMErrorField(err error) string {
	st, ok := status.FromError(err)
	if !ok {
		return ""
	}
	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok {
			for _, fv := range badReq.GetFieldViolations() {
				if field := strings.TrimSpace(fv.GetField()); field != "" {
					return strings.ToLower(field)
				}
			}
		}
	}
	return ""
}

func safeMessage(msg, fallback string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

// ToProtoScopeType mengubah string scope menjadi enum protobuf ScopeType IAM.
func ToProtoScopeType(raw string) iamv1.ScopeType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case authz.ScopeGlobal:
		return iamv1.ScopeType_SCOPE_TYPE_GLOBAL
	case "legal_entity":
		return iamv1.ScopeType_SCOPE_TYPE_LEGAL_ENTITY
	case "organization_unit":
		return iamv1.ScopeType_SCOPE_TYPE_ORGANIZATION_UNIT
	case "facility":
		return iamv1.ScopeType_SCOPE_TYPE_FACILITY
	case "supplier_party":
		return iamv1.ScopeType_SCOPE_TYPE_SUPPLIER_PARTY
	case "vendor_party":
		return iamv1.ScopeType_SCOPE_TYPE_VENDOR_PARTY
	case "self":
		return iamv1.ScopeType_SCOPE_TYPE_SELF
	default:
		return iamv1.ScopeType_SCOPE_TYPE_UNSPECIFIED
	}
}

// FromProtoScopeTypeString mengubah ScopeType protobuf ke string kanonik authz.
func FromProtoScopeTypeString(t iamv1.ScopeType) string {
	switch t {
	case iamv1.ScopeType_SCOPE_TYPE_GLOBAL:
		return authz.ScopeGlobal
	case iamv1.ScopeType_SCOPE_TYPE_LEGAL_ENTITY:
		return "legal_entity"
	case iamv1.ScopeType_SCOPE_TYPE_ORGANIZATION_UNIT:
		return "organization_unit"
	case iamv1.ScopeType_SCOPE_TYPE_FACILITY:
		return "facility"
	case iamv1.ScopeType_SCOPE_TYPE_SUPPLIER_PARTY:
		return "supplier_party"
	case iamv1.ScopeType_SCOPE_TYPE_VENDOR_PARTY:
		return "vendor_party"
	case iamv1.ScopeType_SCOPE_TYPE_SELF:
		return "self"
	default:
		return ""
	}
}

// FromProtoUserKindString mengonversi enum protobuf UserKind menjadi string kanonik.
func FromProtoUserKindString(k iamv1.UserKind) string {
	switch k {
	case iamv1.UserKind_USER_KIND_CUSTOMER:
		return "customer"
	case iamv1.UserKind_USER_KIND_EMPLOYEE:
		return "employee"
	case iamv1.UserKind_USER_KIND_VENDOR:
		return "vendor"
	case iamv1.UserKind_USER_KIND_SUPPLIER:
		return "supplier"
	case iamv1.UserKind_USER_KIND_SYSTEM:
		return "system"
	default:
		return ""
	}
}

// FromProtoUserStatusString mengonversi enum protobuf UserStatus menjadi string kanonik.
func FromProtoUserStatusString(s iamv1.UserStatus) string {
	switch s {
	case iamv1.UserStatus_USER_STATUS_PENDING:
		return "pending"
	case iamv1.UserStatus_USER_STATUS_ACTIVE:
		return authz.StatusActive
	case iamv1.UserStatus_USER_STATUS_SUSPENDED:
		return "suspended"
	case iamv1.UserStatus_USER_STATUS_DISABLED:
		return "disabled"
	default:
		return ""
	}
}
