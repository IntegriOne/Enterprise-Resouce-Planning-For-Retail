package handler

import (
	"context"
	"fmt"
	"strings"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// UserHandler menangani permintaan gRPC siklus hidup akun pengguna, profil, snapshot akses, dan identitas login.
type UserHandler struct {
	iamv1.UnimplementedUserServiceServer
	svc       *user.Service
	errMapper *ErrorMapper
}

// NewUserHandler membentuk handler gRPC untuk UserService secara fail-closed.
func NewUserHandler(svc *user.Service, errMapper *ErrorMapper) (*UserHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("user service cannot be nil")
	}
	if errMapper == nil {
		errMapper = NewErrorMapper(nil)
	}
	return &UserHandler{
		svc:       svc,
		errMapper: errMapper,
	}, nil
}

// CreateUser mendaftarkan akun pengguna baru beserta identitas awal dan kredensial sandinya.
func (h *UserHandler) CreateUser(ctx context.Context, req *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	u, err := h.svc.CreateUser(ctx, user.CreateUserInput{
		Kind:                   fromProtoUserKind(req.GetKind()),
		Status:                 fromProtoUserStatus(req.GetStatus()),
		DisplayName:            req.GetDisplayName(),
		AvatarURL:              req.GetAvatarUrl(),
		Locale:                 req.GetLocale(),
		Timezone:               req.GetTimezone(),
		Email:                  req.GetEmail(),
		Username:               req.GetUsername(),
		Password:               req.GetPassword(),
		RequirePasswordChange:  req.GetRequirePasswordChange(),
		MarkIdentitiesVerified: req.GetMarkIdentitiesVerified(),
		ProvisioningRequestID:  req.GetProvisioningRequestId(),
		WorkerID:               req.GetWorkerId(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CreateUserResponse{User: toProtoUser(u)}, nil
}

// GetUserById mengambil entitas pengguna berdasarkan UUID-nya.
func (h *UserHandler) GetUserById(ctx context.Context, req *iamv1.GetUserByIdRequest) (*iamv1.GetUserByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	u, err := h.svc.GetByID(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetUserByIdResponse{User: toProtoUser(u)}, nil
}

// GetUserAccessSnapshot mengambil ringkasan peran dan izin efektif pengguna saat ini.
func (h *UserHandler) GetUserAccessSnapshot(ctx context.Context, req *iamv1.GetUserAccessSnapshotRequest) (*iamv1.GetUserAccessSnapshotResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	snap, err := h.svc.GetAccessSnapshot(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.GetUserAccessSnapshotResponse{Snapshot: toProtoAccessSnapshot(snap)}, nil
}

// ListUsers mengembalikan daftar pengguna terpaginasi berdasarkan filter opsional.
func (h *UserHandler) ListUsers(ctx context.Context, req *iamv1.ListUsersRequest) (*iamv1.ListUsersResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}

	var kindPtr *entity.UserKind
	if req.Kind != nil && req.GetKind() != iamv1.UserKind_USER_KIND_UNSPECIFIED {
		k := fromProtoUserKind(req.GetKind())
		if !k.IsValid() {
			return nil, h.errMapper.ToStatus(iamerrors.WithField(iamerrors.ErrInvalidUserKind, "kind", ""))
		}
		kindPtr = &k
	}

	var statusPtr *entity.UserStatus
	if req.Status != nil && req.GetStatus() != iamv1.UserStatus_USER_STATUS_UNSPECIFIED {
		s := fromProtoUserStatus(req.GetStatus())
		if !s.IsValid() {
			return nil, h.errMapper.ToStatus(iamerrors.WithField(iamerrors.ErrInvalidUserStatus, "status", ""))
		}
		statusPtr = &s
	}

	res, err := h.svc.List(ctx, domainrepo.UserFilter{
		Query:          req.GetQuery(),
		Kind:           kindPtr,
		Status:         statusPtr,
		IncludeDeleted: req.GetIncludeDeleted(),
		Page:           fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListUsersResponse{
		Items: toProtoUsers(res.Items),
		Page:  toProtoPageInfo(res),
	}, nil
}

// UpdateUserProfile memperbarui metadata profil pengguna.
func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *iamv1.UpdateUserProfileRequest) (*iamv1.UpdateUserProfileResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	u, err := h.svc.UpdateProfile(ctx, user.UpdateProfileInput{
		UserID:            req.GetUserId(),
		DisplayName:       req.GetDisplayName(),
		AvatarURL:         req.GetAvatarUrl(),
		PreserveAvatarURL: req.GetPreserveAvatarUrl(),
		Locale:            req.GetLocale(),
		Timezone:          req.GetTimezone(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.UpdateUserProfileResponse{User: toProtoUser(u)}, nil
}

// UpdateUserEmail menyimpan alamat email baru sebagai pending email sekunder sambil tetap memakai email utama lama hingga diverifikasi.
func (h *UserHandler) UpdateUserEmail(ctx context.Context, req *iamv1.UpdateUserEmailRequest) (*iamv1.UpdateUserEmailResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	res, err := h.svc.UpdateEmail(ctx, user.UpdateEmailInput{
		UserID:          req.GetUserId(),
		NewEmail:        req.GetEmail(),
		CurrentPassword: req.GetCurrentPassword(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.UpdateUserEmailResponse{
		PendingEmail:      res.PendingIdentity.Value,
		IsPending:         true,
		ExpiresAt:         toProtoTimestamp(res.ExpiresAt),
		Identity:          toProtoUserIdentity(res.PendingIdentity),
		VerificationToken: res.VerificationToken,
	}, nil
}

func (h *UserHandler) UpdateUserUsername(ctx context.Context, req *iamv1.UpdateUserUsernameRequest) (*iamv1.UpdateUserUsernameResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	ident, err := h.svc.UpdateUsername(ctx, user.UpdateUsernameInput{
		UserID:   req.GetUserId(),
		Username: req.GetUsername(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.UpdateUserUsernameResponse{
		Username: ident.Value,
		Identity: toProtoUserIdentity(ident),
	}, nil
}

// DeletePendingUserEmail membatalkan pengubahan email tertunda dan mencabut token verifikasi email aktif terkait.
func (h *UserHandler) DeletePendingUserEmail(ctx context.Context, req *iamv1.DeletePendingUserEmailRequest) (*iamv1.DeletePendingUserEmailResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	deleted, err := h.svc.DeletePendingEmail(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DeletePendingUserEmailResponse{Deleted: deleted}, nil
}

// ResendValidationEmailToken menerbitkan tepat 1 token verifikasi email aktif terbaru baik via user_id (login) maupun email (registrasi pending).
func (h *UserHandler) ResendValidationEmailToken(ctx context.Context, req *iamv1.ResendValidationEmailTokenRequest) (*iamv1.ResendValidationEmailTokenResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if strings.TrimSpace(req.GetUserId()) != "" {
		if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
			return nil, h.errMapper.ToStatus(err)
		}
	}
	res, err := h.svc.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
		UserID: req.GetUserId(),
		Email:  req.GetEmail(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ResendValidationEmailTokenResponse{
		IsResend:          res.IsResend,
		PendingEmail:      res.PendingEmail,
		VerificationToken: res.VerificationToken,
		ExpiresAt:         toProtoTimestampPtr(res.ExpiresAt),
	}, nil
}

// ActivateUser mengubah status pengguna menjadi aktif.
func (h *UserHandler) ActivateUser(ctx context.Context, req *iamv1.ActivateUserRequest) (*iamv1.ActivateUserResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	u, err := h.svc.Activate(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ActivateUserResponse{User: toProtoUser(u)}, nil
}

// SuspendUser menangguhkan akun pengguna dan mencabut sesi aktifnya.
func (h *UserHandler) SuspendUser(ctx context.Context, req *iamv1.SuspendUserRequest) (*iamv1.SuspendUserResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	u, err := h.svc.Suspend(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.SuspendUserResponse{User: toProtoUser(u)}, nil
}

// DisableUser menonaktifkan akun pengguna secara permanen/administratif dan mencabut sesi.
func (h *UserHandler) DisableUser(ctx context.Context, req *iamv1.DisableUserRequest) (*iamv1.DisableUserResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	u, err := h.svc.Disable(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DisableUserResponse{User: toProtoUser(u)}, nil
}

// SoftDeleteUser menandai akun pengguna terhapus dan mengakhiri seluruh sesinya.
func (h *UserHandler) SoftDeleteUser(ctx context.Context, req *iamv1.SoftDeleteUserRequest) (*iamv1.SoftDeleteUserResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.SoftDelete(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.SoftDeleteUserResponse{}, nil
}

// ListUserIdentities mengambil seluruh identitas login yang terhubung ke pengguna.
func (h *UserHandler) ListUserIdentities(ctx context.Context, req *iamv1.ListUserIdentitiesRequest) (*iamv1.ListUserIdentitiesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	idents, err := h.svc.ListIdentities(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListUserIdentitiesResponse{Identities: toProtoUserIdentities(idents)}, nil
}

// AddUserIdentity menambahkan identitas login baru untuk pengguna.
func (h *UserHandler) AddUserIdentity(ctx context.Context, req *iamv1.AddUserIdentityRequest) (*iamv1.AddUserIdentityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	ident, err := h.svc.AddIdentity(ctx, user.AddIdentityInput{
		UserID:    req.GetUserId(),
		Kind:      fromProtoIdentityKind(req.GetKind()),
		Value:     req.GetValue(),
		IsPrimary: req.GetIsPrimary(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AddUserIdentityResponse{Identity: toProtoUserIdentity(ident)}, nil
}

// RequestIdentityVerification menerbitkan token sekali pakai untuk memverifikasi kepemilikan identitas.
func (h *UserHandler) RequestIdentityVerification(ctx context.Context, req *iamv1.RequestIdentityVerificationRequest) (*iamv1.RequestIdentityVerificationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.RequestIdentityVerification(ctx, req.GetIdentityId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RequestIdentityVerificationResponse{IssuedToken: toProtoIssuedToken(res)}, nil
}

// ConfirmIdentityVerification menandai identitas terverifikasi dengan token valid.
func (h *UserHandler) ConfirmIdentityVerification(ctx context.Context, req *iamv1.ConfirmIdentityVerificationRequest) (*iamv1.ConfirmIdentityVerificationResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	ident, err := h.svc.ConfirmIdentityVerification(ctx, req.GetIdentityId(), req.GetRawToken(), req.GetActivatePendingUser())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ConfirmIdentityVerificationResponse{Identity: toProtoUserIdentity(ident)}, nil
}

// SetPrimaryUserIdentity menetapkan identitas tertentu sebagai identitas utama untuk tipenya.
func (h *UserHandler) SetPrimaryUserIdentity(ctx context.Context, req *iamv1.SetPrimaryUserIdentityRequest) (*iamv1.SetPrimaryUserIdentityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	ident, err := h.svc.SetPrimaryIdentity(ctx, req.GetUserId(), req.GetIdentityId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.SetPrimaryUserIdentityResponse{Identity: toProtoUserIdentity(ident)}, nil
}

// RemoveUserIdentity mencabut identitas login pengguna jika syarat minimum identitas terpenuhi.
func (h *UserHandler) RemoveUserIdentity(ctx context.Context, req *iamv1.RemoveUserIdentityRequest) (*iamv1.RemoveUserIdentityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RemoveIdentity(ctx, req.GetUserId(), req.GetIdentityId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RemoveUserIdentityResponse{}, nil
}

// CheckUsernameAvailability memeriksa ketersediaan username tanpa mengekspos PII atau keberadaan identitas pengguna.
func (h *UserHandler) CheckUsernameAvailability(ctx context.Context, req *iamv1.CheckUsernameAvailabilityRequest) (*iamv1.CheckUsernameAvailabilityResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.CheckUsernameAvailability(ctx, req.GetUsername())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	if res == nil {
		return nil, h.errMapper.ToStatus(iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.identity.empty_username_check_result"), "username availability check returned empty result"))
	}
	return &iamv1.CheckUsernameAvailabilityResponse{
		Available:          res.Available,
		NormalizedUsername: res.NormalizedUsername,
		Reason:             res.Reason,
	}, nil
}

// ConfirmEmailVerificationByToken memverifikasi tautan verifikasi email sekali pakai dan mengaktifkan akun pending.
func (h *UserHandler) ConfirmEmailVerificationByToken(ctx context.Context, req *iamv1.ConfirmEmailVerificationByTokenRequest) (*iamv1.ConfirmEmailVerificationByTokenResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.ConfirmEmailVerificationByToken(ctx, req.GetRawToken())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ConfirmEmailVerificationByTokenResponse{
		Verified: res.Status == user.VerifyEmailStatusVerified,
		Status:   toProtoVerifyEmailStatus(res.Status),
		Message:  res.Message,
		User:     toProtoUser(res.User),
		Identity: toProtoUserIdentity(res.Identity),
	}, nil
}

// enforceSelfCaller menolak panggilan terautentikasi yang menargetkan principal selain pemanggil itu sendiri.
func enforceSelfCaller(ctx context.Context, targetUserID string) error {
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return iamerrors.ErrUnauthorized
	}
	callerID := strings.TrimSpace(caller.PrincipalID)
	targetID := strings.TrimSpace(targetUserID)
	if callerID == "" || targetID == "" || !strings.EqualFold(callerID, targetID) {
		return iamerrors.WithField(iamerrors.ErrPermissionDenied, "user_id", "cannot access another principal self resource")
	}
	return nil
}

// enforceSelfSessionCaller menolak mutasi sesi mandiri terautentikasi jika user_id atau session_id bukan milik pemanggil.
func enforceSelfSessionCaller(ctx context.Context, targetUserID, sessionID string) error {
	if err := enforceSelfCaller(ctx, targetUserID); err != nil {
		return err
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return iamerrors.ErrUnauthorized
	}
	callerSessID := strings.TrimSpace(caller.SessionID)
	targetSessID := strings.TrimSpace(sessionID)
	if callerSessID == "" || targetSessID == "" || !strings.EqualFold(callerSessID, targetSessID) {
		return iamerrors.WithField(iamerrors.ErrPermissionDenied, "session_id", "cannot operate on another session in self flow")
	}
	return nil
}
