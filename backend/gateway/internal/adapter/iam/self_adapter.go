package iam

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

const (
	maxAvatarBytes        = 512 * 1024
	maxEncodedAvatarChars = 1024 * 1024
)

// Me mengambil profil pengguna milik principal yang sedang terautentikasi.
func (a *Adapter) Me(ctx context.Context) (*model.User, error) {
	if err := a.ensureAuthorizerDeps(); err != nil {
		return nil, err
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersReadSelf, false)
	if err != nil {
		return nil, err
	}
	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.GetUserByIdResponse, error) {
		return users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{
			UserId: sess.Token.UserID,
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelUser(resp.GetUser()), nil
}

// MyAccessSnapshot mengembalikan snapshot otorisasi efektif terkini milik principal aktif.
func (a *Adapter) MyAccessSnapshot(ctx context.Context) (*model.AccessSnapshot, error) {
	if err := a.ensureAuthDeps(); err != nil {
		return nil, err
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	return toModelAccessSnapshot(sess.Snapshot), nil
}

// MySessions menampilkan daftar sesi aktif maupun riwayat sesi milik principal aktif dalam format Relay Connection.
func (a *Adapter) MySessions(
	ctx context.Context,
	activeOnly *bool,
	after *string,
	before *string,
	first *int,
	last *int,
) (*model.SessionConnection, error) {
	if err := a.ensureAuthorizerDeps(); err != nil {
		return nil, err
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermSessionsReadSelf, false)
	if err != nil {
		return nil, err
	}
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}

	items, startOffset, pageInfo, total, err := FetchRelayPage(win, func(limit, offset int32) ([]*iamv1.Session, *iamv1.PageInfo, error) {
		resp, rpcErr := callIAMSessions(ctx, a, &sess.Caller, func(rpcCtx context.Context, sessions iamv1.SessionServiceClient) (*iamv1.ListUserSessionsResponse, error) {
			return sessions.ListUserSessions(rpcCtx, &iamv1.ListUserSessionsRequest{
				UserId:     sess.Token.UserID,
				ActiveOnly: derefBool(activeOnly, true),
				Page:       &iamv1.PageRequest{Limit: limit, Offset: offset},
			})
		})
		if rpcErr != nil {
			return nil, nil, rpcErr
		}
		return resp.GetItems(), resp.GetPage(), nil
	})
	if err != nil {
		return nil, err
	}
	return buildSessionConnection(items, startOffset, pageInfo, total), nil
}

// UpdateMyProfile memperbarui profil dasar principal aktif tanpa menghapus atau menimpa avatar_url yang sudah ada.
func (a *Adapter) UpdateMyProfile(ctx context.Context, input model.UpdateMyProfileInput) (*model.UpdateMyProfilePayload, error) {
	fail := func(err error) (*model.UpdateMyProfilePayload, error) {
		return &model.UpdateMyProfilePayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersUpdateSelf, false)
	if err != nil {
		return fail(err)
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		return fail(middleware.ErrInvalidArgument("display_name is required", nil))
	}

	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.UpdateUserProfileResponse, error) {
		return users.UpdateUserProfile(rpcCtx, &iamv1.UpdateUserProfileRequest{
			UserId:            sess.Token.UserID,
			DisplayName:       displayName,
			Locale:            derefStr(input.Locale),
			Timezone:          derefStr(input.Timezone),
			PreserveAvatarUrl: true,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.UpdateMyProfilePayload{
		User:          toModelUser(resp.GetUser()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// UploadAvatarUser memvalidasi biner gambar AVIF base64 lalu menyimpan URL HTTPS deterministik ke profil pengguna aktif.
func (a *Adapter) UploadAvatarUser(ctx context.Context, input model.UserAvatarUploadInput) (*model.UserAvatarPayload, error) {
	fail := func(err error) (*model.UserAvatarPayload, error) {
		return &model.UserAvatarPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersUpdateSelf, false)
	if err != nil {
		return fail(err)
	}
	rawBytes, err := decodeAndValidateAVIF(input.Image)
	if err != nil {
		return fail(err)
	}

	currentResp, getErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.GetUserByIdResponse, error) {
		return users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{
			UserId: sess.Token.UserID,
		})
	})
	if getErr != nil {
		return fail(getErr)
	}
	currentUser := currentResp.GetUser()
	if currentUser == nil {
		return fail(middleware.ErrNotFound("authenticated user profile not found", nil))
	}

	digest := sha256.Sum256(rawBytes)
	avatarURL := fmt.Sprintf("%s/%s/%s.avif", a.cfg.AvatarBaseURL, sess.Token.UserID, hex.EncodeToString(digest[:]))

	updateResp, updErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.UpdateUserProfileResponse, error) {
		return users.UpdateUserProfile(rpcCtx, &iamv1.UpdateUserProfileRequest{
			UserId:            sess.Token.UserID,
			DisplayName:       currentUser.GetDisplayName(),
			AvatarUrl:         avatarURL,
			Locale:            currentUser.GetLocale(),
			Timezone:          currentUser.GetTimezone(),
			PreserveAvatarUrl: false,
		})
	})
	if updErr != nil {
		return fail(updErr)
	}
	storedURL := avatarURL
	if updateResp.GetUser() != nil && strings.TrimSpace(updateResp.GetUser().GetAvatarUrl()) != "" {
		storedURL = updateResp.GetUser().GetAvatarUrl()
	}
	return &model.UserAvatarPayload{
		AvatarURL:     ptrStr(storedURL),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// UpdateUserEmail mendaftarkan atau memperbarui alamat email pending sambil menjaga email terverifikasi lama tetap aktif hingga dikonfirmasi.
func (a *Adapter) UpdateUserEmail(ctx context.Context, input model.UserEmailInput) (*model.UserEmailPayload, error) {
	fail := func(err error) (*model.UserEmailPayload, error) {
		return &model.UserEmailPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfSensitive(ctx, authz.PermUsersUpdateSelf)
	if err != nil {
		return fail(err)
	}
	email := strings.TrimSpace(input.Email)
	if email == "" || input.CurrentPassword == "" {
		return fail(middleware.ErrInvalidArgument("email and current_password are required", nil))
	}

	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.UpdateUserEmailResponse, error) {
		return users.UpdateUserEmail(rpcCtx, &iamv1.UpdateUserEmailRequest{
			UserId:          sess.Token.UserID,
			Email:           email,
			CurrentPassword: input.CurrentPassword,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.UserEmailPayload{
		PendingEmail:  ptrStr(resp.GetPendingEmail()),
		IsPending:     ptrBool(resp.GetIsPending()),
		ExpiresAt:     toLocalTimePtr(resp.GetExpiresAt()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// ChangeUsername mengganti atau menambahkan username terverifikasi milik pengguna aktif tanpa verifikasi email tambahan.
func (a *Adapter) ChangeUsername(ctx context.Context, input model.UsernameInput) (*model.UserNamePayload, error) {
	fail := func(err error) (*model.UserNamePayload, error) {
		return &model.UserNamePayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersUpdateSelf, false)
	if err != nil {
		return fail(err)
	}
	username := strings.TrimSpace(input.Username)
	if username == "" {
		return fail(middleware.ErrInvalidArgument("username is required", nil))
	}

	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.UpdateUserUsernameResponse, error) {
		return users.UpdateUserUsername(rpcCtx, &iamv1.UpdateUserUsernameRequest{
			UserId:   sess.Token.UserID,
			Username: username,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.UserNamePayload{
		Username:      ptrStr(resp.GetUsername()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// DeletePendingUserEmail membatalkan alamat email pending yang belum terverifikasi dan mencabut token aktifnya.
func (a *Adapter) DeletePendingUserEmail(ctx context.Context) (*model.DeletePendingUserEmailPayload, error) {
	fail := func(err error) (*model.DeletePendingUserEmailPayload, error) {
		return &model.DeletePendingUserEmailPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersDeleteSelf, false)
	if err != nil {
		return fail(err)
	}
	resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.DeletePendingUserEmailResponse, error) {
		return users.DeletePendingUserEmail(rpcCtx, &iamv1.DeletePendingUserEmailRequest{
			UserId: sess.Token.UserID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.DeletePendingUserEmailPayload{
		Deleted:       ptrBool(resp.GetDeleted()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// ChangeUserPassword mengganti kata sandi principal aktif dengan wajib verifikasi reauth terkini & MFA bila aktif, lalu mencabut seluruh sesi aktif.
func (a *Adapter) ChangeUserPassword(ctx context.Context, input model.ChangeUserPasswordInput) (*model.ChangeUserPasswordPayload, error) {
	fail := func(err error) (*model.ChangeUserPasswordPayload, error) {
		return &model.ChangeUserPasswordPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfSensitive(ctx, authz.PermUsersUpdateSelf)
	if err != nil {
		return fail(err)
	}
	if input.CurrentPassword == "" || input.NewPassword == "" {
		return fail(middleware.ErrInvalidArgument("current_password and new_password are required", nil))
	}

	resp, rpcErr := callIAMAuth(ctx, a, &sess.Caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.ChangePasswordResponse, error) {
		return auth.ChangePassword(rpcCtx, &iamv1.ChangePasswordRequest{
			UserId:          sess.Token.UserID,
			CurrentPassword: input.CurrentPassword,
			NewPassword:     input.NewPassword,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	if resp.GetChanged() {
		a.authManager.ClearSessionCookie(ctx)
		if eff := browserauth.EffectsFromContext(ctx); eff != nil {
			eff.SetHeader("X-CSRF-Token", "")
		}
	}
	return &model.ChangeUserPasswordPayload{
		Changed:           ptrBool(resp.GetChanged()),
		SessionsRevoked:   ptrInt(int(resp.GetSessionsRevoked())),
		PasswordChangedAt: toLocalTimePtr(resp.GetPasswordChangedAt()),
		Errors:            noMutationErrors(),
		WasSuccessful:     true,
	}, nil
}

// VerifyMySessionMfa meningkatkan level sesi aktif dengan TOTP step-up dan menyinkronkan header X-CSRF-Token baru bila diputar.
func (a *Adapter) VerifyMySessionMfa(ctx context.Context, totpCode string) (*model.VerifyMySessionMfaPayload, error) {
	fail := func(err error) (*model.VerifyMySessionMfaPayload, error) {
		return &model.VerifyMySessionMfaPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersUpdateSelf, false)
	if err != nil {
		return fail(err)
	}
	cleanCode := strings.TrimSpace(totpCode)
	if cleanCode == "" {
		return fail(middleware.ErrInvalidArgument("totp_code is required", nil))
	}

	resp, rpcErr := callIAMAuth(ctx, a, &sess.Caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.VerifySessionMfaResponse, error) {
		return auth.VerifySessionMfa(rpcCtx, &iamv1.VerifySessionMfaRequest{
			UserId:    sess.Token.UserID,
			SessionId: sess.Token.SessionID,
			Code:      cleanCode,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	if csrf := strings.TrimSpace(resp.GetCsrfToken()); csrf != "" {
		if eff := browserauth.EffectsFromContext(ctx); eff != nil {
			eff.SetHeader(browserauth.CSRFHeaderName, csrf)
		}
	}
	return &model.VerifyMySessionMfaPayload{
		Verified:      ptrBool(true),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// EnrollMyTotp memulai pendaftaran TOTP MFA mandiri setelah melewati pemeriksaan re-autentikasi sensitif.
func (a *Adapter) EnrollMyTotp(ctx context.Context) (*model.EnrollTotpPayload, error) {
	fail := func(err error) (*model.EnrollTotpPayload, error) {
		return &model.EnrollTotpPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfSensitive(ctx, authz.PermUsersUpdateSelf)
	if err != nil {
		return fail(err)
	}
	if sess == nil || sess.Token == nil {
		return fail(middleware.ErrUnauthenticated("authentication required", nil))
	}

	accountLabel := sess.Token.UserID
	if a.clients.Managed() || a.clients.Users != nil {
		userResp, uErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.GetUserByIdResponse, error) {
			return users.GetUserById(rpcCtx, &iamv1.GetUserByIdRequest{UserId: sess.Token.UserID})
		})
		if uErr != nil {
			if gwErr, ok := middleware.AsGatewayError(uErr); !ok || gwErr.Code != middleware.CodeNotFound {
				return fail(uErr)
			}
		} else if userResp.GetUser() != nil {
			if dn := strings.TrimSpace(userResp.GetUser().GetDisplayName()); dn != "" {
				accountLabel = dn
			}
		}
	}

	resp, rpcErr := callIAMAuth(ctx, a, &sess.Caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.EnrollTotpResponse, error) {
		return auth.EnrollTotp(rpcCtx, &iamv1.EnrollTotpRequest{
			UserId:       sess.Token.UserID,
			AccountLabel: accountLabel,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.EnrollTotpPayload{
		FactorID:      ptrStr(resp.GetFactorId()),
		OtpAuthURI:    ptrStr(resp.GetOtpAuthUri()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// ConfirmMyTotp mengonfirmasi aktivasi faktor TOTP dan mengembalikan kode pemulihan satu kali.
func (a *Adapter) ConfirmMyTotp(ctx context.Context, input model.ConfirmTotpInput) (*model.ConfirmTotpPayload, error) {
	fail := func(err error) (*model.ConfirmTotpPayload, error) {
		return &model.ConfirmTotpPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfSensitive(ctx, authz.PermUsersUpdateSelf)
	if err != nil {
		return fail(err)
	}
	cleanCode := strings.TrimSpace(input.Code)
	if cleanCode == "" {
		return fail(middleware.ErrInvalidArgument("code is required", nil))
	}

	resp, rpcErr := callIAMAuth(ctx, a, &sess.Caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.ConfirmTotpResponse, error) {
		return auth.ConfirmTotp(rpcCtx, &iamv1.ConfirmTotpRequest{
			UserId: sess.Token.UserID,
			Code:   cleanCode,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	enabled := resp.GetFactor() != nil && resp.GetFactor().GetIsEnabled()
	return &model.ConfirmTotpPayload{
		Enabled:       ptrBool(enabled),
		RecoveryCodes: resp.GetRecoveryCodes(),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// DisableMyTotp menonaktifkan MFA TOTP milik principal aktif setelah validasi sesi sensitif.
func (a *Adapter) DisableMyTotp(ctx context.Context) (*model.DisableMyTotpPayload, error) {
	fail := func(err error) (*model.DisableMyTotpPayload, error) {
		return &model.DisableMyTotpPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfSensitive(ctx, authz.PermUsersUpdateSelf)
	if err != nil {
		return fail(err)
	}

	_, rpcErr := callIAMAuth(ctx, a, &sess.Caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.DisableTotpResponse, error) {
		return auth.DisableTotp(rpcCtx, &iamv1.DisableTotpRequest{
			UserId: sess.Token.UserID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.DisableMyTotpPayload{
		Disabled:      ptrBool(true),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// RegenerateMyRecoveryCodes membuat ulang daftar kode pemulihan MFA principal aktif.
func (a *Adapter) RegenerateMyRecoveryCodes(ctx context.Context) (*model.RegenerateMyRecoveryCodesPayload, error) {
	fail := func(err error) (*model.RegenerateMyRecoveryCodesPayload, error) {
		return &model.RegenerateMyRecoveryCodesPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authorizer.RequireSelfSensitive(ctx, authz.PermUsersUpdateSelf)
	if err != nil {
		return fail(err)
	}

	resp, rpcErr := callIAMAuth(ctx, a, &sess.Caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.RegenerateRecoveryCodesResponse, error) {
		return auth.RegenerateRecoveryCodes(rpcCtx, &iamv1.RegenerateRecoveryCodesRequest{
			UserId: sess.Token.UserID,
		})
	})
	if rpcErr != nil {
		return fail(rpcErr)
	}
	return &model.RegenerateMyRecoveryCodesPayload{
		RecoveryCodes: resp.GetRecoveryCodes(),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

func decodeAndValidateAVIF(rawInput string) ([]byte, error) {
	trimmed := strings.TrimSpace(rawInput)
	if trimmed == "" {
		return nil, middleware.ErrInvalidArgument("avatar image payload is required", nil)
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "data:") {
		idx := strings.IndexByte(trimmed, ',')
		if idx < 0 {
			return nil, middleware.ErrInvalidArgument("invalid data URI encoding", nil)
		}
		header := strings.ToLower(strings.TrimSpace(trimmed[:idx]))
		if header != "data:image/avif;base64" {
			return nil, middleware.ErrInvalidArgument("avatar image must be image/avif base64", nil)
		}
		trimmed = strings.TrimSpace(trimmed[idx+1:])
	}
	if len(trimmed) > maxEncodedAvatarChars {
		return nil, middleware.ErrInvalidArgument("avatar image exceeds maximum allowed size of 512KB", nil)
	}

	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(trimmed)
		if err != nil {
			return nil, middleware.ErrInvalidArgument("avatar image is not valid base64", err)
		}
	}
	if len(raw) == 0 {
		return nil, middleware.ErrInvalidArgument("avatar image payload is empty", nil)
	}
	if len(raw) > maxAvatarBytes {
		return nil, middleware.ErrInvalidArgument("avatar image exceeds maximum allowed size of 512KB", nil)
	}
	if !isValidAVIFHeader(raw) {
		return nil, middleware.ErrInvalidArgument("avatar image is not a valid AVIF container", nil)
	}
	return raw, nil
}

func isValidAVIFHeader(data []byte) bool {
	// Box ISOBMFF `ftyp` minimal 16 byte: size(4) + "ftyp"(4) + major_brand(4) + minor_version(4).
	if len(data) < 16 {
		return false
	}
	boxSize := binary.BigEndian.Uint32(data[0:4])
	if !bytes.Equal(data[4:8], []byte("ftyp")) {
		return false
	}
	if boxSize < 16 || int(boxSize) > len(data) || (boxSize-16)%4 != 0 {
		return false
	}
	major := data[8:12]
	if bytes.Equal(major, []byte("avif")) || bytes.Equal(major, []byte("avis")) {
		return true
	}
	for offset := 16; offset+4 <= int(boxSize); offset += 4 {
		brand := data[offset : offset+4]
		if bytes.Equal(brand, []byte("avif")) || bytes.Equal(brand, []byte("avis")) {
			return true
		}
	}
	return false
}
