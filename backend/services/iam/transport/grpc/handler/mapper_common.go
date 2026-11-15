package handler

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
)

func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func toProtoTimestampPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// fromProtoTimestampPtr mengonversi protobuf Timestamp ke *time.Time zona lokal OS.
func fromProtoTimestampPtr(field string, ts *timestamppb.Timestamp) (*time.Time, error) {
	if ts == nil {
		return nil, nil
	}
	if err := ts.CheckValid(); err != nil {
		return nil, iamerrors.NewValidation(field, iamerrors.Code("iam.invalid_timestamp"), "timestamp is invalid")
	}
	localTime := ts.AsTime().Local()
	return &localTime, nil
}

func fromProtoDuration(field string, d *durationpb.Duration) (time.Duration, error) {
	if d == nil {
		return 0, nil
	}
	if err := d.CheckValid(); err != nil {
		return 0, iamerrors.NewValidation(field, iamerrors.Code("iam.invalid_duration"), "duration is invalid")
	}
	return d.AsDuration(), nil
}

func fromProtoPageRequest(p *iamv1.PageRequest) domainrepo.PageRequest {
	if p == nil {
		return domainrepo.PageRequest{}
	}
	return domainrepo.PageRequest{
		Limit:  int(p.GetLimit()),
		Offset: int(p.GetOffset()),
	}
}

func toProtoPageInfo[T any](res domainrepo.PageResult[T]) *iamv1.PageInfo {
	return &iamv1.PageInfo{
		Total:   res.Total,
		Limit:   int32(res.Limit),
		Offset:  int32(res.Offset),
		HasMore: res.HasMore,
	}
}

func toProtoUserKind(k entity.UserKind) iamv1.UserKind {
	switch k {
	case entity.UserKindCustomer:
		return iamv1.UserKind_USER_KIND_CUSTOMER
	case entity.UserKindEmployee:
		return iamv1.UserKind_USER_KIND_EMPLOYEE
	case entity.UserKindVendor:
		return iamv1.UserKind_USER_KIND_VENDOR
	case entity.UserKindSupplier:
		return iamv1.UserKind_USER_KIND_SUPPLIER
	case entity.UserKindSystem:
		return iamv1.UserKind_USER_KIND_SYSTEM
	default:
		return iamv1.UserKind_USER_KIND_UNSPECIFIED
	}
}

func fromProtoUserKind(k iamv1.UserKind) entity.UserKind {
	switch k {
	case iamv1.UserKind_USER_KIND_CUSTOMER:
		return entity.UserKindCustomer
	case iamv1.UserKind_USER_KIND_EMPLOYEE:
		return entity.UserKindEmployee
	case iamv1.UserKind_USER_KIND_VENDOR:
		return entity.UserKindVendor
	case iamv1.UserKind_USER_KIND_SUPPLIER:
		return entity.UserKindSupplier
	case iamv1.UserKind_USER_KIND_SYSTEM:
		return entity.UserKindSystem
	case iamv1.UserKind_USER_KIND_UNSPECIFIED:
		return ""
	default:
		return entity.UserKind("invalid")
	}
}

func toProtoUserStatus(s entity.UserStatus) iamv1.UserStatus {
	switch s {
	case entity.UserStatusPending:
		return iamv1.UserStatus_USER_STATUS_PENDING
	case entity.UserStatusActive:
		return iamv1.UserStatus_USER_STATUS_ACTIVE
	case entity.UserStatusSuspended:
		return iamv1.UserStatus_USER_STATUS_SUSPENDED
	case entity.UserStatusDisabled:
		return iamv1.UserStatus_USER_STATUS_DISABLED
	default:
		return iamv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func fromProtoUserStatus(s iamv1.UserStatus) entity.UserStatus {
	switch s {
	case iamv1.UserStatus_USER_STATUS_PENDING:
		return entity.UserStatusPending
	case iamv1.UserStatus_USER_STATUS_ACTIVE:
		return entity.UserStatusActive
	case iamv1.UserStatus_USER_STATUS_SUSPENDED:
		return entity.UserStatusSuspended
	case iamv1.UserStatus_USER_STATUS_DISABLED:
		return entity.UserStatusDisabled
	case iamv1.UserStatus_USER_STATUS_UNSPECIFIED:
		return ""
	default:
		return entity.UserStatus("invalid")
	}
}

func toProtoIdentityKind(k entity.IdentityKind) iamv1.IdentityKind {
	switch k {
	case entity.IdentityKindEmail:
		return iamv1.IdentityKind_IDENTITY_KIND_EMAIL
	case entity.IdentityKindUsername:
		return iamv1.IdentityKind_IDENTITY_KIND_USERNAME
	default:
		return iamv1.IdentityKind_IDENTITY_KIND_UNSPECIFIED
	}
}

func fromProtoIdentityKind(k iamv1.IdentityKind) entity.IdentityKind {
	switch k {
	case iamv1.IdentityKind_IDENTITY_KIND_EMAIL:
		return entity.IdentityKindEmail
	case iamv1.IdentityKind_IDENTITY_KIND_USERNAME:
		return entity.IdentityKindUsername
	case iamv1.IdentityKind_IDENTITY_KIND_UNSPECIFIED:
		return ""
	default:
		return entity.IdentityKind("invalid")
	}
}

func toProtoTokenPurpose(p entity.TokenPurpose) iamv1.TokenPurpose {
	switch p {
	case entity.TokenPurposeEmailVerification:
		return iamv1.TokenPurpose_TOKEN_PURPOSE_EMAIL_VERIFICATION
	case entity.TokenPurposePasswordReset:
		return iamv1.TokenPurpose_TOKEN_PURPOSE_PASSWORD_RESET
	case entity.TokenPurposeAccountInvitation:
		return iamv1.TokenPurpose_TOKEN_PURPOSE_ACCOUNT_INVITATION
	case entity.TokenPurposeMFAChallenge:
		return iamv1.TokenPurpose_TOKEN_PURPOSE_MFA_CHALLENGE
	default:
		return iamv1.TokenPurpose_TOKEN_PURPOSE_UNSPECIFIED
	}
}

func toProtoTokenStatus(s entity.TokenStatus) iamv1.TokenStatus {
	switch s {
	case entity.TokenStatusActive:
		return iamv1.TokenStatus_TOKEN_STATUS_ACTIVE
	case entity.TokenStatusConsumed:
		return iamv1.TokenStatus_TOKEN_STATUS_CONSUMED
	case entity.TokenStatusExpired:
		return iamv1.TokenStatus_TOKEN_STATUS_EXPIRED
	case entity.TokenStatusRevoked:
		return iamv1.TokenStatus_TOKEN_STATUS_REVOKED
	default:
		return iamv1.TokenStatus_TOKEN_STATUS_UNSPECIFIED
	}
}

func toProtoSessionEndReason(r entity.SessionEndReason) iamv1.SessionEndReason {
	switch r {
	case entity.SessionEndLogout:
		return iamv1.SessionEndReason_SESSION_END_REASON_LOGOUT
	case entity.SessionEndIdleTimeout:
		return iamv1.SessionEndReason_SESSION_END_REASON_IDLE_TIMEOUT
	case entity.SessionEndAbsoluteTimeout:
		return iamv1.SessionEndReason_SESSION_END_REASON_ABSOLUTE_TIMEOUT
	case entity.SessionEndPasswordChanged:
		return iamv1.SessionEndReason_SESSION_END_REASON_PASSWORD_CHANGED
	case entity.SessionEndAdminRevoked:
		return iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED
	case entity.SessionEndUserDisabled:
		return iamv1.SessionEndReason_SESSION_END_REASON_USER_DISABLED
	default:
		return iamv1.SessionEndReason_SESSION_END_REASON_UNSPECIFIED
	}
}

func fromProtoSessionEndReason(r iamv1.SessionEndReason) entity.SessionEndReason {
	switch r {
	case iamv1.SessionEndReason_SESSION_END_REASON_LOGOUT:
		return entity.SessionEndLogout
	case iamv1.SessionEndReason_SESSION_END_REASON_IDLE_TIMEOUT:
		return entity.SessionEndIdleTimeout
	case iamv1.SessionEndReason_SESSION_END_REASON_ABSOLUTE_TIMEOUT:
		return entity.SessionEndAbsoluteTimeout
	case iamv1.SessionEndReason_SESSION_END_REASON_PASSWORD_CHANGED:
		return entity.SessionEndPasswordChanged
	case iamv1.SessionEndReason_SESSION_END_REASON_ADMIN_REVOKED:
		return entity.SessionEndAdminRevoked
	case iamv1.SessionEndReason_SESSION_END_REASON_USER_DISABLED:
		return entity.SessionEndUserDisabled
	case iamv1.SessionEndReason_SESSION_END_REASON_UNSPECIFIED:
		return ""
	default:
		return entity.SessionEndReason("invalid")
	}
}

func toProtoScopeType(t entity.ScopeType) iamv1.ScopeType {
	switch t {
	case entity.ScopeTypeGlobal:
		return iamv1.ScopeType_SCOPE_TYPE_GLOBAL
	case entity.ScopeTypeLegalEntity:
		return iamv1.ScopeType_SCOPE_TYPE_LEGAL_ENTITY
	case entity.ScopeTypeOrganizationUnit:
		return iamv1.ScopeType_SCOPE_TYPE_ORGANIZATION_UNIT
	case entity.ScopeTypeFacility:
		return iamv1.ScopeType_SCOPE_TYPE_FACILITY
	case entity.ScopeTypeSupplierParty:
		return iamv1.ScopeType_SCOPE_TYPE_SUPPLIER_PARTY
	case entity.ScopeTypeVendorParty:
		return iamv1.ScopeType_SCOPE_TYPE_VENDOR_PARTY
	case entity.ScopeTypeSelf:
		return iamv1.ScopeType_SCOPE_TYPE_SELF
	default:
		return iamv1.ScopeType_SCOPE_TYPE_UNSPECIFIED
	}
}

func fromProtoScopeType(t iamv1.ScopeType) entity.ScopeType {
	switch t {
	case iamv1.ScopeType_SCOPE_TYPE_GLOBAL:
		return entity.ScopeTypeGlobal
	case iamv1.ScopeType_SCOPE_TYPE_LEGAL_ENTITY:
		return entity.ScopeTypeLegalEntity
	case iamv1.ScopeType_SCOPE_TYPE_ORGANIZATION_UNIT:
		return entity.ScopeTypeOrganizationUnit
	case iamv1.ScopeType_SCOPE_TYPE_FACILITY:
		return entity.ScopeTypeFacility
	case iamv1.ScopeType_SCOPE_TYPE_SUPPLIER_PARTY:
		return entity.ScopeTypeSupplierParty
	case iamv1.ScopeType_SCOPE_TYPE_VENDOR_PARTY:
		return entity.ScopeTypeVendorParty
	case iamv1.ScopeType_SCOPE_TYPE_SELF:
		return entity.ScopeTypeSelf
	case iamv1.ScopeType_SCOPE_TYPE_UNSPECIFIED:
		return ""
	default:
		return entity.ScopeType("invalid")
	}
}

func toProtoScope(s entity.Scope) *iamv1.Scope {
	var idCopy *string
	if s.ID != nil {
		v := *s.ID
		idCopy = &v
	}
	return &iamv1.Scope{
		Type: toProtoScopeType(s.Type),
		Id:   idCopy,
	}
}

func fromProtoScopeOptional(s *iamv1.Scope) (entity.Scope, error) {
	if s == nil || s.GetType() == iamv1.ScopeType_SCOPE_TYPE_UNSPECIFIED {
		return entity.GlobalScope(), nil
	}
	return entity.NewScope(fromProtoScopeType(s.GetType()), s.Id)
}
