package group

import (
	"context"
	"errors"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// Service mengelola entitas grup pengguna, keanggotaan anggota grup, dan pewarisan peran grup.
type Service struct {
	tx              domainrepo.Transactor
	groups          domainrepo.GroupRepository
	members         domainrepo.GroupMemberRepository
	gRoles          domainrepo.GroupRoleRepository
	roles           domainrepo.RoleRepository
	users           domainrepo.UserRepository
	securityRuntime domainrepo.SecurityRuntimeRepository
	clock           domainservice.Clock
	idGen           domainservice.IDGenerator
	outboxCoord     *redisstore.Coordinator
	projector       domainservice.RuntimeSecurityProjector
}

// NewService membentuk Service grup dengan validasi fail-closed atas seluruh ketergantungan wajib.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Transactor == nil {
		return nil, iamerrors.NewValidation("transactor", iamerrors.Code("iam.service.nil_transactor"), "transactor is required")
	}
	if deps.Groups == nil {
		return nil, iamerrors.NewValidation("groups", iamerrors.Code("iam.service.nil_group_repo"), "group repository is required")
	}
	if deps.GroupMembers == nil {
		return nil, iamerrors.NewValidation("group_members", iamerrors.Code("iam.service.nil_group_member_repo"), "group member repository is required")
	}
	if deps.GroupRoles == nil {
		return nil, iamerrors.NewValidation("group_roles", iamerrors.Code("iam.service.nil_group_role_repo"), "group role repository is required")
	}
	if deps.Roles == nil {
		return nil, iamerrors.NewValidation("roles", iamerrors.Code("iam.service.nil_role_repo"), "role repository is required")
	}
	if deps.Users == nil {
		return nil, iamerrors.NewValidation("users", iamerrors.Code("iam.service.nil_user_repo"), "user repository is required")
	}
	if deps.SecurityRuntime == nil {
		return nil, iamerrors.NewValidation("security_runtime", iamerrors.Code("iam.service.nil_security_runtime_repo"), "security runtime repository is required")
	}
	if deps.OutboxCoord == nil {
		return nil, iamerrors.NewValidation("outbox_coordinator", iamerrors.Code("iam.service.nil_outbox_coordinator"), "outbox coordinator is required")
	}
	if deps.IDGen == nil {
		return nil, iamerrors.NewValidation("id_gen", iamerrors.Code("iam.service.nil_id_generator"), "id generator is required")
	}
	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	return &Service{
		tx:              deps.Transactor,
		groups:          deps.Groups,
		members:         deps.GroupMembers,
		gRoles:          deps.GroupRoles,
		roles:           deps.Roles,
		users:           deps.Users,
		securityRuntime: deps.SecurityRuntime,
		clock:           clock,
		idGen:           deps.IDGen,
		outboxCoord:     deps.OutboxCoord,
		projector:       deps.Projector,
	}, nil
}

// GetByID mengambil detail grup berdasarkan UUID-nya.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", id); err != nil {
		return nil, err
	}
	return s.groups.SelectByID(ctx, id)
}

// GetByCode mengambil detail grup berdasarkan kode unik bernormalisasi.
func (s *Service) GetByCode(ctx context.Context, code string) (*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normalized, err := entity.ValidateEntityCode(code, entity.MaxGroupCodeLength, iamerrors.ErrInvalidGroupCode)
	if err != nil {
		return nil, err
	}
	return s.groups.SelectByCode(ctx, normalized)
}

// List mengambil daftar grup berhalaman berdasarkan kata kunci dan status aktif.
func (s *Service) List(ctx context.Context, filter domainrepo.GroupFilter) (domainrepo.PageResult[*entity.Group], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Group]{}, err
	}
	return s.groups.SelectList(ctx, filter.Normalize())
}

// Create membentuk grup baru dengan kode unik tervalidasi dalam kondisi aktif.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.group.id_generation_failed"), "failed to generate group id")
	}
	grp, err := entity.NewGroup(id, in.Code, in.Name, in.Description, s.clock.Now())
	if err != nil {
		return nil, err
	}

	existing, err := s.groups.SelectByCode(ctx, grp.Code)
	if err == nil && existing != nil {
		return nil, iamerrors.WithField(iamerrors.ErrGroupAlreadyExists, "code", "group code already exists")
	}
	if err != nil && !errors.Is(err, iamerrors.ErrGroupNotFound) {
		return nil, err
	}

	if err := s.groups.Insert(ctx, grp); err != nil {
		return nil, err
	}
	return grp, nil
}

// UpdateDetails memperbarui nama dan deskripsi grup pengguna.
func (s *Service) UpdateDetails(ctx context.Context, in UpdateInput) (*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", in.ID); err != nil {
		return nil, err
	}

	var updated *entity.Group
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		grp, err := s.groups.SelectByID(txCtx, in.ID)
		if err != nil {
			return err
		}
		if err := grp.UpdateDetails(in.Name, in.Description, s.clock.Now()); err != nil {
			return err
		}
		if err := s.groups.Update(txCtx, grp); err != nil {
			return err
		}
		updated = grp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Activate mengaktifkan kembali grup sehingga keanggotaan dan pewarisan perannya berlaku.
func (s *Service) Activate(ctx context.Context, id string) (*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", id); err != nil {
		return nil, err
	}

	var updated *entity.Group
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		grp, err := s.groups.SelectByID(txCtx, id)
		if err != nil {
			return err
		}
		grp.Activate(s.clock.Now())
		if err := s.groups.Update(txCtx, grp); err != nil {
			return err
		}
		if err := s.bumpAndRecordGroupSecurityInTx(txCtx, grp.ID); err != nil {
			return err
		}
		updated = grp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Deactivate menonaktifkan grup agar pewarisan peran melalui grup berhenti dievaluasi.
func (s *Service) Deactivate(ctx context.Context, id string) (*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", id); err != nil {
		return nil, err
	}

	var updated *entity.Group
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		grp, err := s.groups.SelectByID(txCtx, id)
		if err != nil {
			return err
		}
		grp.Deactivate(s.clock.Now())
		if err := s.groups.Update(txCtx, grp); err != nil {
			return err
		}
		if err := s.bumpAndRecordGroupSecurityInTx(txCtx, grp.ID); err != nil {
			return err
		}
		updated = grp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete menghapus grup pengguna dari repositori berdasarkan UUID-nya.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("group_id", id); err != nil {
		return err
	}
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.bumpAndRecordGroupSecurityInTx(txCtx, id); err != nil {
			return err
		}
		return s.groups.Delete(txCtx, id)
	})
}

// AddMember memasukkan pengguna ke dalam grup aktif secara atomik tanpa duplikasi.
func (s *Service) AddMember(ctx context.Context, groupID, userID string) (*entity.GroupMember, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", groupID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}

	var created *entity.GroupMember
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		grp, err := s.groups.SelectByID(txCtx, groupID)
		if err != nil {
			return err
		}
		if !grp.IsActive {
			return iamerrors.ErrGroupInactive
		}

		u, err := s.users.SelectByID(txCtx, userID)
		if err != nil {
			return err
		}
		if u.IsDeleted() {
			return iamerrors.ErrUserDeleted
		}

		exists, err := s.members.CheckExists(txCtx, groupID, userID)
		if err != nil {
			return err
		}
		if exists {
			return iamerrors.ErrGroupMemberExists
		}

		member, err := entity.NewGroupMember(groupID, userID, s.clock.Now())
		if err != nil {
			return err
		}
		if caller, ok := authcontext.FromContext(txCtx); ok && caller.Authenticated && caller.PrincipalID != "" {
			if err := member.SetAddedBy(caller.PrincipalID); err != nil {
				return err
			}
		}
		if err := s.members.Insert(txCtx, member); err != nil {
			return err
		}
		if err := s.bumpAndRecordUserPermissionsInTx(txCtx, userID); err != nil {
			return err
		}
		created = member
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// RemoveMember menghapus keanggotaan pengguna dari grup.
func (s *Service) RemoveMember(ctx context.Context, groupID, userID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("group_id", groupID); err != nil {
		return err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return err
	}
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.members.Delete(txCtx, groupID, userID); err != nil {
			return err
		}
		return s.bumpAndRecordUserPermissionsInTx(txCtx, userID)
	})
}

// ListMembers mengambil daftar pengguna anggota suatu grup secara berhalaman.
func (s *Service) ListMembers(ctx context.Context, groupID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.User], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}
	if err := entity.ValidateUUID("group_id", groupID); err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}
	if _, err := s.groups.SelectByID(ctx, groupID); err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}
	return s.members.SelectUsersByGroupID(ctx, groupID, page.Normalize())
}

// ListUserGroups mengambil daftar grup tempat pengguna menjadi anggota.
func (s *Service) ListUserGroups(ctx context.Context, userID string) ([]*entity.Group, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return s.members.SelectGroupsByUserID(ctx, userID)
}

// AssignRole menugaskan peran aktif berscope ke grup pengguna secara transaksional.
func (s *Service) AssignRole(ctx context.Context, in AssignGroupRoleInput) (*entity.GroupRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", in.GroupID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", in.RoleID); err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if in.ExpiresAt != nil && !in.ExpiresAt.After(now) {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "expires_at", "expires_at must be in the future")
	}

	scope, err := entity.NewScope(in.ScopeType, in.ScopeID)
	if err != nil {
		return nil, err
	}
	assignmentID, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.group_role.id_generation_failed"), "failed to generate group role id")
	}

	var created *entity.GroupRole
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		grp, err := s.groups.SelectByID(txCtx, in.GroupID)
		if err != nil {
			return err
		}
		if !grp.IsActive {
			return iamerrors.ErrGroupInactive
		}

		r, err := s.roles.SelectByID(txCtx, in.RoleID)
		if err != nil {
			return err
		}
		if !r.IsActive {
			return iamerrors.ErrRoleInactive
		}

		exists, err := s.gRoles.CheckExists(txCtx, in.GroupID, in.RoleID, scope)
		if err != nil {
			return err
		}
		if exists {
			return iamerrors.ErrGroupRoleExists
		}

		var grantedBy *string
		if caller, ok := authcontext.FromContext(txCtx); ok && caller.Authenticated && caller.PrincipalID != "" {
			if err := entity.ValidateUUID("principal_id", caller.PrincipalID); err != nil {
				return err
			}
			actor := caller.PrincipalID
			grantedBy = &actor
		}
		assignment, err := entity.NewGroupRole(assignmentID, in.GroupID, in.RoleID, scope, grantedBy, in.ExpiresAt, now)
		if err != nil {
			return err
		}
		if err := s.gRoles.Insert(txCtx, assignment); err != nil {
			return err
		}
		if err := s.bumpAndRecordGroupSecurityInTx(txCtx, in.GroupID); err != nil {
			return err
		}
		created = assignment
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// RevokeRole mencabut penugasan peran pada grup berdasarkan ID penugasannya.
func (s *Service) RevokeRole(ctx context.Context, assignmentID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("group_role_id", assignmentID); err != nil {
		return err
	}
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		groupID, err := s.securityRuntime.ResolveGroupRoleGroupID(txCtx, assignmentID)
		if err != nil {
			return err
		}
		if err := s.gRoles.DeleteByID(txCtx, assignmentID); err != nil {
			return err
		}
		return s.bumpAndRecordGroupSecurityInTx(txCtx, groupID)
	})
}

// ListGroupRoles mengambil daftar penugasan peran pada suatu grup.
func (s *Service) ListGroupRoles(ctx context.Context, groupID string, includeExpired bool) ([]*entity.GroupRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("group_id", groupID); err != nil {
		return nil, err
	}
	if _, err := s.groups.SelectByID(ctx, groupID); err != nil {
		return nil, err
	}
	return s.gRoles.SelectByGroupID(ctx, groupID, includeExpired, s.clock.Now())
}

func (s *Service) ensureReady() error {
	if s == nil || s.groups == nil || s.tx == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.group.service_unavailable"), "group service is not initialized")
	}
	return nil
}
