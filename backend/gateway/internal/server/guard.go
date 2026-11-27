package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

type httpReqContextKey struct{}

func withHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, httpReqContextKey{}, r)
}

func httpRequestFromContext(ctx context.Context) (*http.Request, bool) {
	if ctx == nil {
		return nil, false
	}
	r, ok := ctx.Value(httpReqContextKey{}).(*http.Request)
	return r, ok && r != nil
}

const (
	// maxGraphQLDepth membatasi kedalaman bersarang kueri GraphQL untuk mencegah serangan DoS rekursif.
	maxGraphQLDepth = 10
	// maxGraphQLAliases membatasi jumlah alias field per operasi untuk mencegah amplifikasi batching.
	maxGraphQLAliases = 25
)

// buildOperationGuard membentuk OperationMiddleware untuk menegakkan aturan metode HTTP, Origin, batas struktur AST, rate limit mutasi, dan X-CSRF-Token per operasi GraphQL.
func buildOperationGuard(
	authManager *middleware.AuthManager,
	originValidator *browserauth.OriginValidator,
	presenter graphql.ErrorPresenterFunc,
	enableIntrospection bool,
	mutationRateMax int64,
	mutationRateWindow time.Duration,
) graphql.OperationMiddleware {
	if mutationRateMax <= 0 {
		mutationRateMax = 60
	}
	if mutationRateWindow <= 0 {
		mutationRateWindow = time.Minute
	}
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx == nil {
			return next(ctx)
		}
		if !enableIntrospection {
			opCtx.DisableIntrospection = true
		}
		if opCtx.Operation == nil {
			return next(ctx)
		}

		req, hasReq := httpRequestFromContext(ctx)
		if !hasReq {
			return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("http request context missing", nil))
		}
		if err := validateOperationComplexityShape(opCtx.Doc, opCtx.Operation, maxGraphQLDepth, maxGraphQLAliases); err != nil {
			return singleErrorResponse(ctx, presenter, err)
		}

		switch opCtx.Operation.Operation {
		case ast.Subscription:
			return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("subscriptions are not supported on /api/graphql", nil))

		case ast.Query:
			if req.Method != http.MethodGet && req.Method != http.MethodPost {
				return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("http method not allowed for query", nil))
			}
			// Query mengizinkan panggilan SSR cookie-only pada POST tanpa Origin/CSRF, namun tetap menolak Origin/Sec-Fetch-Site asing bila hadir.
			if err := originValidator.ValidateRequest(req.Header, false); err != nil {
				return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("origin is not allowed", err))
			}
			return next(ctx)

		case ast.Mutation:
			if req.Method == http.MethodGet {
				return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("mutations are not permitted via HTTP GET", nil))
			}
			if req.Method != http.MethodPost {
				return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("mutations require HTTP POST", nil))
			}
			if err := originValidator.ValidateRequest(req.Header, true); err != nil {
				return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("valid same-origin Origin header is required for mutations", err))
			}
			if authManager == nil {
				return singleErrorResponse(ctx, presenter, middleware.ErrUnavailable("authentication manager unavailable", nil))
			}
			meta, _ := middleware.RequestMetadataFromContext(ctx)
			ip := strings.TrimSpace(meta.ClientIP)
			if ip == "" {
				ip = "unknown"
			}
			if err := authManager.CheckRateLimit(ctx, "mutation_ip", ip, mutationRateMax, mutationRateWindow); err != nil {
				return singleErrorResponse(ctx, presenter, err)
			}
			resolved, err := authManager.RequireMutationCSRF(ctx)
			if err != nil {
				return singleErrorResponse(ctx, presenter, err)
			}
			if resolved != nil {
				ctx = middleware.WithResolvedSession(ctx, resolved)
			}
			return next(ctx)

		default:
			return singleErrorResponse(ctx, presenter, middleware.ErrForbidden("unsupported GraphQL operation type", nil))
		}
	}
}

func singleErrorResponse(ctx context.Context, presenter graphql.ErrorPresenterFunc, err error) graphql.ResponseHandler {
	if presenter == nil {
		presenter = graphql.DefaultErrorPresenter
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return graphql.OneShot(&graphql.Response{
		Errors: gqlerror.List{presenter(ctx, err)},
	})
}

func validateOperationComplexityShape(doc *ast.QueryDocument, op *ast.OperationDefinition, maxDepth, maxAliases int) error {
	if op == nil {
		return nil
	}
	aliases := 0
	activeFrags := make(map[string]bool)
	return walkSelectionSet(doc, op.SelectionSet, 0, maxDepth, maxAliases, activeFrags, &aliases)
}

func walkSelectionSet(
	doc *ast.QueryDocument,
	set ast.SelectionSet,
	depth, maxDepth, maxAliases int,
	activeFrags map[string]bool,
	aliases *int,
) error {
	for _, sel := range set {
		switch s := sel.(type) {
		case *ast.Field:
			if s.Alias != "" && s.Alias != s.Name {
				*aliases++
				if *aliases > maxAliases {
					return middleware.ErrInvalidArgument("graphql query exceeds maximum alias count", nil)
				}
			}
			nextDepth := depth + 1
			if nextDepth > maxDepth {
				return middleware.ErrInvalidArgument("graphql query exceeds maximum depth", nil)
			}
			if len(s.SelectionSet) > 0 {
				if err := walkSelectionSet(doc, s.SelectionSet, nextDepth, maxDepth, maxAliases, activeFrags, aliases); err != nil {
					return err
				}
			}
		case *ast.InlineFragment:
			if len(s.SelectionSet) > 0 {
				if err := walkSelectionSet(doc, s.SelectionSet, depth, maxDepth, maxAliases, activeFrags, aliases); err != nil {
					return err
				}
			}
		case *ast.FragmentSpread:
			if activeFrags[s.Name] {
				return middleware.ErrInvalidArgument("graphql fragment cycle is not allowed", nil)
			}
			frag := s.Definition
			if frag == nil && doc != nil {
				frag = doc.Fragments.ForName(s.Name)
			}
			if frag != nil && len(frag.SelectionSet) > 0 {
				activeFrags[s.Name] = true
				err := walkSelectionSet(doc, frag.SelectionSet, depth, maxDepth, maxAliases, activeFrags, aliases)
				delete(activeFrags, s.Name)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
