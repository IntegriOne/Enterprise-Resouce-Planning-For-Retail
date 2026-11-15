package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// GroupRepository mendefinisikan operasi persistensi untuk tabel groups.
type GroupRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Group, error)
	SelectByCode(ctx context.Context, code string) (*entity.Group, error)
	SelectList(ctx context.Context, filter GroupFilter) (PageResult[*entity.Group], error)
	Insert(ctx context.Context, group *entity.Group) error
	Update(ctx context.Context, group *entity.Group) error
	Delete(ctx context.Context, id string) error
}

// GroupMemberRepository mendefinisikan operasi keanggotaan pengguna pada tabel group_members.
type GroupMemberRepository interface {
	SelectUsersByGroupID(ctx context.Context, groupID string, page PageRequest) (PageResult[*entity.User], error)
	SelectGroupsByUserID(ctx context.Context, userID string) ([]*entity.Group, error)
	CheckExists(ctx context.Context, groupID, userID string) (bool, error)
	Insert(ctx context.Context, member *entity.GroupMember) error
	Delete(ctx context.Context, groupID, userID string) error
}

// GroupRoleRepository mendefinisikan operasi penugasan peran kolektif grup pada tabel group_roles.
type GroupRoleRepository interface {
	SelectByGroupID(ctx context.Context, groupID string, includeExpired bool, now time.Time) ([]*entity.GroupRole, error)
	CheckExists(ctx context.Context, groupID, roleID string, scope entity.Scope) (bool, error)
	Insert(ctx context.Context, assignment *entity.GroupRole) error
	DeleteByID(ctx context.Context, id string) error
}
