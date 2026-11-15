package resolvers

import (
	"github.com/integrione/erp-retail/backend/gateway/graph/generated"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	iamadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/iam"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

// Resolver menyimpan dependensi adapter domain yang melayani eksekusi GraphQL Gateway.
type Resolver struct {
	IAMAdapter *iamadapter.Adapter
}

func (r *Resolver) iam() (*iamadapter.Adapter, error) {
	if r == nil || r.IAMAdapter == nil {
		return nil, middleware.ErrUnavailable("IAM adapter is not initialized", nil)
	}
	return r.IAMAdapter, nil
}

// NewComplexityRoot membentuk aturan penghitungan kompleksitas kueri Relay terpaginasi untuk batas FixedComplexityLimit.
func NewComplexityRoot() generated.ComplexityRoot {
	var c generated.ComplexityRoot

	paginated := func(childComplexity int, first, last *int) int {
		limit := 20
		if first != nil && *first > 0 {
			limit = *first
		} else if last != nil && *last > 0 {
			limit = *last
		}
		if limit > 100 {
			// Menolak permintaan limit halaman berlebih langsung pada lapis kompleksitas GraphQL.
			return limit * 10
		}
		factor := (limit + 3) / 4
		if factor < 1 {
			factor = 1
		}
		if childComplexity < 1 {
			childComplexity = 1
		}
		return factor * childComplexity
	}

	c.Query.Users = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *model.UserKind, _ *model.UserStatus, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserIdentities = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserRoles = func(childComplexity int, _ string, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserEffectiveRoles = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserGroups = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserSessions = func(childComplexity int, _ string, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.RolePermissions = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Permissions = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Groups = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.GroupMembers = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.GroupRoles = func(childComplexity int, _ string, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.MySessions = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}

	c.User.Identities = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.User.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.User.EffectiveRoles = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.User.Groups = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.User.Sessions = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Role.Permissions = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Group.Members = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Group.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}

	return c
}
