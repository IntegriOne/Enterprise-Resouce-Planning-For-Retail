// Package enterprise_foundation mengimplementasikan adapter GraphQL ke gRPC untuk manajemen Enterprise Foundation beserta kontrol RBAC dan isolasi gangguan.
package enterprise_foundation

import (
	"context"
	"errors"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

const defaultAdapterRPCTimeout = 5 * time.Second

// Adapter menghubungkan resolver GraphQL Enterprise Foundation dengan klien gRPC downstream yang terisolasi gangguan.
type Adapter struct {
	clients    *source_connections.EnterpriseFoundationClientSet
	authorizer *middleware.Authorizer
	rpcTimeout time.Duration
}

// NewAdapter membentuk Adapter Enterprise Foundation tanpa menggagalkan startup Gateway apabila koneksi EF belum siap atau dimatikan.
func NewAdapter(
	clients *source_connections.EnterpriseFoundationClientSet,
	authorizer *middleware.Authorizer,
	rpcTimeout time.Duration,
) (*Adapter, error) {
	if authorizer == nil {
		return nil, errors.New("authorizer is required")
	}
	if rpcTimeout <= 0 {
		rpcTimeout = defaultAdapterRPCTimeout
	}
	return &Adapter{
		clients:    clients,
		authorizer: authorizer,
		rpcTimeout: rpcTimeout,
	}, nil
}

// Node melakukan resolusi polimorfik Relay Node untuk seluruh entitas Enterprise Foundation berdasarkan GID ketat.
func (a *Adapter) Node(ctx context.Context, rawGID string) (model.Node, error) {
	if a == nil {
		return nil, middleware.ErrUnavailable("enterprise foundation adapter is unavailable", nil)
	}
	resource, rawID, err := DecodeNodeGID(rawGID)
	if err != nil {
		return nil, err
	}
	switch resource {
	case ResourceAddress:
		return a.Address(ctx, rawID)
	case ResourceLegalEntity:
		return a.LegalEntity(ctx, rawID)
	case ResourceLocation:
		return a.Location(ctx, rawID)
	case ResourceOrganizationUnit:
		return a.OrganizationUnit(ctx, rawID)
	case ResourceParty:
		return a.Party(ctx, rawID)
	case ResourcePartyRole:
		return a.PartyRole(ctx, rawID)
	case ResourcePartyContact:
		return a.PartyContact(ctx, rawID)
	case ResourcePartyAddress:
		return a.PartyAddressLink(ctx, rawID)
	default:
		return nil, middleware.ErrInvalidArgument("unsupported enterprise foundation node GID resource", nil)
	}
}

func callEF[Resp any](
	ctx context.Context,
	a *Adapter,
	sess *middleware.ResolvedSession,
	call func(context.Context, *source_connections.EnterpriseFoundationClientSet) (*Resp, error),
) (*Resp, error) {
	if a == nil {
		return nil, middleware.ErrUnavailable("enterprise foundation service unavailable", nil)
	}
	callCtx := ctx
	if callCtx == nil {
		callCtx = context.Background()
	}
	if sess != nil {
		callCtx = authcontext.WithCaller(callCtx, sess.Caller)
	}
	return source_connections.InvokeEF(callCtx, a.clients, call)
}
