package source_connections

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

const (
	defaultIAMTimeout          = defaultUpstreamTimeout
	defaultIAMFailureThreshold = defaultCircuitFailureThreshold
	defaultIAMOpenDuration     = defaultCircuitOpenDuration
)

var (
	// ErrIAMCircuitOpen menandakan pemutus sirkuit IAM sedang terbuka akibat kegagalan konektivitas beruntun.
	ErrIAMCircuitOpen = errors.New("identity service circuit breaker is open")
	// ErrIAMUnavailable menandakan klien gRPC IAM belum siap atau sedang tidak dapat dijangkau.
	ErrIAMUnavailable = errors.New("identity client is unavailable")
)

// NewIAMClientSet membentuk kumpulan stub gRPC IAM beserta state pemutus sirkuit dan batas waktu.
func NewIAMClientSet(conn grpc.ClientConnInterface, timeout time.Duration) IAMClientSet {
	return newIAMClientSetWithBreaker(
		conn,
		timeout,
		newCircuitBreaker(defaultIAMFailureThreshold, defaultIAMOpenDuration, ErrIAMCircuitOpen, time.Now),
	)
}

func newIAMClientSetWithBreaker(conn grpc.ClientConnInterface, timeout time.Duration, breaker *circuitBreaker) IAMClientSet {
	if timeout <= 0 {
		timeout = defaultIAMTimeout
	}
	if breaker == nil {
		breaker = newCircuitBreaker(defaultIAMFailureThreshold, defaultIAMOpenDuration, ErrIAMCircuitOpen, time.Now)
	}
	if conn == nil {
		return IAMClientSet{
			state: &iamClientState{
				timeout: timeout,
				initErr: ErrIAMUnavailable,
				breaker: breaker,
			},
		}
	}
	return IAMClientSet{
		Auth:                 iamv1.NewAuthServiceClient(conn),
		Users:                iamv1.NewUserServiceClient(conn),
		Sessions:             iamv1.NewSessionServiceClient(conn),
		Roles:                iamv1.NewRoleServiceClient(conn),
		Permission:           iamv1.NewPermissionServiceClient(conn),
		Groups:               iamv1.NewGroupServiceClient(conn),
		EmployeeProvisioning: iamv1.NewEmployeeProvisioningServiceClient(conn),
		state: &iamClientState{
			timeout: timeout,
			breaker: breaker,
		},
	}
}

func newDegradedIAMClientSet(timeout time.Duration, initErr error) IAMClientSet {
	if timeout <= 0 {
		timeout = defaultIAMTimeout
	}
	if initErr == nil {
		initErr = ErrIAMUnavailable
	}
	return IAMClientSet{
		state: &iamClientState{
			timeout: timeout,
			initErr: initErr,
			breaker: newCircuitBreaker(defaultIAMFailureThreshold, defaultIAMOpenDuration, ErrIAMCircuitOpen, time.Now),
		},
	}
}

// initIAMConnection menyiapkan ClientConn non-blocking untuk IAM tanpa menggagalkan startup Gateway saat IAM sedang mati.
func initIAMConnection(cfg Config, secretCopy []byte) (*grpc.ClientConn, IAMClientSet, time.Duration, error) {
	timeout := cfg.IAMTimeout
	if timeout <= 0 {
		timeout = defaultIAMTimeout
	}
	target := strings.TrimSpace(cfg.IAMTarget)
	if target == "" {
		initErr := errors.New("identity grpc target is empty")
		return nil, newDegradedIAMClientSet(timeout, initErr), timeout, initErr
	}

	conn, err := newResilientGRPCConn(target, secretCopy)
	if err != nil {
		return nil, newDegradedIAMClientSet(timeout, err), timeout, err
	}
	return conn, NewIAMClientSet(conn, timeout), timeout, nil
}

// Managed melaporkan apakah IAMClientSet dibentuk secara resmi melalui Registry/NewIAMClientSet (bukan mock struct-literal).
func (c IAMClientSet) Managed() bool {
	return c.state != nil
}

// InitError mengembalikan kesalahan inisialisasi koneksi IAM bila terjadi degradasi startup.
func (c IAMClientSet) InitError() error {
	if c.state == nil {
		return nil
	}
	return c.state.initErr
}

// Ready melaporkan apakah seluruh stub gRPC IAM aktif tanpa kesalahan inisialisasi.
func (c IAMClientSet) Ready() bool {
	return c.InitError() == nil &&
		c.Auth != nil &&
		c.Users != nil &&
		c.Sessions != nil &&
		c.Roles != nil &&
		c.Permission != nil &&
		c.Groups != nil
}

// InvokeIAM mengeksekusi satu panggilan unary IAM secara terisolasi lengkap dengan timeout, pemutus sirkuit, pemulihan panic, dan validasi respons nil.
func InvokeIAM[Resp any](
	ctx context.Context,
	clients IAMClientSet,
	stubAvailable bool,
	timeout time.Duration,
	call func(context.Context, IAMClientSet) (*Resp, error),
) (resp *Resp, err error) {
	if initErr := clients.InitError(); initErr != nil {
		return nil, middleware.ErrUnavailable("identity service unavailable", initErr)
	}
	if !stubAvailable {
		return nil, middleware.ErrUnavailable("identity service unavailable", ErrIAMUnavailable)
	}
	if call == nil {
		return nil, middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", errors.New("identity rpc callback is nil"))
	}

	effectiveTimeout := timeout
	var breaker *circuitBreaker
	if clients.state != nil {
		if effectiveTimeout <= 0 {
			effectiveTimeout = clients.state.timeout
		}
		breaker = clients.state.breaker
	}

	var isProbe bool
	if breaker != nil {
		var cbErr error
		isProbe, cbErr = breaker.beforeCall()
		if cbErr != nil {
			return nil, middleware.ErrUnavailable("identity service unavailable", cbErr)
		}
	}

	rpcCtx, cancel := withBoundedTimeout(ctx, effectiveTimeout)
	defer cancel()

	defer func() {
		if recovered := recover(); recovered != nil {
			if breaker != nil {
				breaker.afterCall(isProbe, true, false)
			}
			resp = nil
			err = middleware.ErrUnavailable(
				"identity service unavailable",
				fmt.Errorf("recovered identity panic: %v", recovered),
			)
		}
	}()

	res, rpcErr := call(rpcCtx, clients)
	if breaker != nil {
		breaker.afterCall(isProbe, isUpstreamUnavailableFailure(ctx, rpcErr), isLocalCallerCancellation(ctx, rpcErr))
	}
	if rpcErr != nil {
		return nil, MapIAMGRPCError(rpcErr)
	}
	if res == nil {
		return nil, middleware.NewGatewayError(
			middleware.CodeInternalError,
			"internal server error",
			errors.New("identity service returned nil response"),
		)
	}
	return res, nil
}

// IAMRuntimeBridge menghubungkan IAM gRPC ke SnapshotLoader dan Authorizer Gateway tanpa memaparkan status mentah ke resolver.
type IAMRuntimeBridge struct {
	clients IAMClientSet
	timeout time.Duration
}

// NewIAMRuntimeBridge membentuk jembatan runtime gRPC IAM untuk penyegaran snapshot dan pemeriksaan reauth sensitif.
func NewIAMRuntimeBridge(clients IAMClientSet, timeout time.Duration) *IAMRuntimeBridge {
	if timeout <= 0 {
		timeout = defaultIAMTimeout
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
	if b == nil {
		return nil, middleware.ErrUnavailable("identity service unavailable", ErrIAMUnavailable)
	}
	callCtx := authcontext.WithCaller(ctx, caller)
	resp, err := InvokeIAM(callCtx, b.clients, b.clients.Sessions != nil, b.timeout, func(rpcCtx context.Context, c IAMClientSet) (*iamv1.RefreshUserAccessSnapshotResponse, error) {
		return c.Sessions.RefreshUserAccessSnapshot(rpcCtx, &iamv1.RefreshUserAccessSnapshotRequest{
			UserId: userID,
		})
	})
	if err != nil {
		return nil, err
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
	if b == nil {
		return middleware.ErrUnavailable("identity service unavailable", ErrIAMUnavailable)
	}
	var reqScope *iamv1.Scope
	if strings.TrimSpace(requiredPermission) != "" {
		var scopeID *string
		if strings.TrimSpace(requiredScope.ID) != "" {
			v := strings.TrimSpace(requiredScope.ID)
			scopeID = &v
		}
		reqScope = &iamv1.Scope{Type: ToProtoScopeType(requiredScope.Type), Id: scopeID}
	}

	callCtx := authcontext.WithCaller(ctx, caller)
	_, err := InvokeIAM(callCtx, b.clients, b.clients.Sessions != nil, b.timeout, func(rpcCtx context.Context, c IAMClientSet) (*iamv1.AuthorizeSessionResponse, error) {
		return c.Sessions.AuthorizeSession(rpcCtx, &iamv1.AuthorizeSessionRequest{
			SessionToken:           rawToken,
			RequireCsrf:            false,
			RequiredPermission:     requiredPermission,
			RequiredScope:          reqScope,
			RequireSensitiveReauth: true,
		})
	})
	return err
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
	if errors.Is(err, ErrIAMCircuitOpen) || errors.Is(err, ErrIAMUnavailable) {
		return middleware.ErrUnavailable("identity service unavailable", err)
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
	case "iam.service_unavailable", "iam.scope.owner_verification_unavailable":
		return middleware.ErrUnavailable(safeMessage(st.Message(), "identity service unavailable"), err)
	case "iam.credential.invalid", "iam.totp.invalid_code", "iam.recovery_code.invalid":
		return middleware.ErrInvalidCredentials(safeMessage(st.Message(), "invalid credentials"), err)
	case "iam.credential.locked":
		return middleware.ErrTooManyRequests("account temporarily locked due to failed attempts", err)
	case "iam.session.expired", "iam.session.idle_expired", "iam.session.absolute_expired", "iam.session.revoked":
		return middleware.ErrUnauthenticated("session expired or invalid", err)
	case "iam.session.snapshot_stale":
		return middleware.ErrUnavailable("authorization snapshot stale", err)
	case "iam.session.reauth_required", "iam.rbac.sensitive_reauth_required", "iam.session.csrf_mismatch", "iam.user.inactive", "iam.user.suspended", "iam.user.disabled", "iam.identity.not_verified", "iam.credential.password_rotation_required", "iam.user.system_cannot_login", "iam.user.system_interactive_forbidden", "iam.rbac.role_user_kind_mismatch":
		return middleware.ErrForbidden(safeMessage(st.Message(), "forbidden"), err)
	case "iam.identity.already_exists", "iam.identity.same_email", "iam.identity.same_username", "iam.identity.limit_exceeded", "iam.identity.email_already_verified", "iam.role.already_exists", "iam.role.in_use", "iam.group.already_exists":
		return middleware.ErrConflict(safeMessage(st.Message(), "request conflicts with current state"), err)
	case "iam.identity.cannot_remove_last", "iam.identity.cannot_remove_primary", "iam.user.deleted", "iam.role.system_protected", "iam.role.system_immutable":
		return middleware.ErrUnprocessableEntity(safeMessage(st.Message(), "request could not be processed"), err)
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return middleware.ErrInvalidArgument(safeMessage(st.Message(), "invalid input"), err)
	case codes.NotFound:
		return middleware.ErrNotFound(safeMessage(st.Message(), "resource not found"), err)
	case codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
		return middleware.ErrConflict(safeMessage(st.Message(), "request conflicts with current state"), err)
	case codes.Unauthenticated:
		// Kegagalan otentikasi konteks internal S2S antar Gateway-IAM tidak boleh menghapus sesi browser pengguna.
		if isInternalCallerUnauthenticatedMessage(st.Message()) {
			return middleware.ErrUnavailable("identity service unavailable", err)
		}
		return middleware.ErrUnauthenticated("authentication required", err)
	case codes.PermissionDenied:
		return middleware.ErrForbidden(safeMessage(st.Message(), "forbidden"), err)
	case codes.ResourceExhausted:
		return middleware.ErrTooManyRequests("too many requests, please retry later", err)
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
		return middleware.ErrUnavailable("identity service unavailable", err)
	default:
		return middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", err)
	}
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
	return extractErrorField(st)
}

func safeMessage(msg, fallback string) string {
	return sanitizeUpstreamMessage(msg, fallback, "iam:9001", ":9001", "postgres", "redis")
}

func isInternalCallerUnauthenticatedMessage(msg string) bool {
	lower := strings.ToLower(strings.TrimSpace(msg))
	return lower == "internal caller verification is unavailable" ||
		lower == "invalid internal caller context" ||
		lower == "missing internal caller context"
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
