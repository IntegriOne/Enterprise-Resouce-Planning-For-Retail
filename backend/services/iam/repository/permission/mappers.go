package permission

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func permissionToDomain(m *permissionModel) (*entity.Permission, error) {
	if m == nil {
		return nil, iamerrors.ErrPermissionNotFound
	}
	return &entity.Permission{
		ID:          strings.ToLower(strings.TrimSpace(m.ID)),
		Code:        m.Code,
		Resource:    m.Resource,
		Action:      m.Action,
		Description: pgstore.DerefString(m.Description),
		IsSensitive: m.IsSensitive,
		CreatedAt:   pgstore.NormalizeLocalTime(m.CreatedAt),
	}, nil
}

func permissionToModel(p *entity.Permission) (permissionModel, error) {
	if p == nil {
		return permissionModel{}, iamerrors.NewValidation("permission", iamerrors.Code("iam.permission.nil"), "permission must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", p.ID)
	if err != nil {
		return permissionModel{}, err
	}
	code, err := entity.ValidatePermissionCode(p.Code)
	if err != nil {
		return permissionModel{}, err
	}
	resource := entity.NormalizeCode(p.Resource)
	if resource == "" {
		return permissionModel{}, iamerrors.NewValidation("resource", iamerrors.Code("iam.permission.invalid_resource"), "permission resource must not be empty")
	}
	action := entity.NormalizeCode(p.Action)
	if action == "" {
		return permissionModel{}, iamerrors.NewValidation("action", iamerrors.Code("iam.permission.invalid_action"), "permission action must not be empty")
	}
	return permissionModel{
		ID:          id,
		Code:        code,
		Resource:    resource,
		Action:      action,
		Description: pgstore.NullableStringPtr(p.Description),
		IsSensitive: p.IsSensitive,
		CreatedAt:   pgstore.MustLocalNow(p.CreatedAt),
	}, nil
}
