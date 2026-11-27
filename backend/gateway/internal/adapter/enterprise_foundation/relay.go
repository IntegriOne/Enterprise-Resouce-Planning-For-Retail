package enterprise_foundation

import (
	"encoding/base64"
	"math"
	"strconv"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
)

const (
	defaultPageLimit    = 20
	maxPageLimit        = 100
	cursorVersionPrefix = "v1:"
	maxPaginationOffset = 1_000_000_000
)

// RelayWindow menyimpan jendela paginasi Relay (cursor & limit) tervalidasi.
type RelayWindow struct {
	StartOffset  int
	BeforeOffset *int
	Limit        int
	IsBackward   bool
	ExplicitZero bool
}

// EncodeCursor menyandikan offset numerik menjadi cursor Relay opaque.
func EncodeCursor(offset int) string {
	if offset < 0 {
		offset = 0
	}
	return base64.RawURLEncoding.EncodeToString([]byte(cursorVersionPrefix + strconv.Itoa(offset)))
}

// DecodeCursor memvalidasi serta menguraikan cursor Relay opaque menjadi offset non-negatif.
func DecodeCursor(cursor string) (int, error) {
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return 0, middleware.ErrInvalidArgument("pagination cursor is empty", nil)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return 0, middleware.ErrInvalidArgument("invalid pagination cursor encoding", err)
	}
	raw := string(decoded)
	if !strings.HasPrefix(raw, cursorVersionPrefix) {
		return 0, middleware.ErrInvalidArgument("unsupported pagination cursor version", nil)
	}
	offset, err := strconv.Atoi(strings.TrimPrefix(raw, cursorVersionPrefix))
	if err != nil || offset < 0 || offset > maxPaginationOffset {
		return 0, middleware.ErrInvalidArgument("invalid pagination cursor offset", err)
	}
	return offset, nil
}

// ParseRelayPage memvalidasi argumen cursor Relay (after, before, first, last).
func ParseRelayPage(after, before *string, first, last *int) (*RelayWindow, error) {
	if first != nil && *first < 0 {
		return nil, middleware.ErrInvalidArgument("first must be non-negative", nil)
	}
	if last != nil && *last < 0 {
		return nil, middleware.ErrInvalidArgument("last must be non-negative", nil)
	}
	if first != nil && last != nil {
		return nil, middleware.ErrInvalidArgument("cannot specify both first and last", nil)
	}

	startOffset := 0
	if after != nil && strings.TrimSpace(*after) != "" {
		idx, err := DecodeCursor(*after)
		if err != nil {
			return nil, err
		}
		startOffset = idx + 1
	}

	var beforeOffset *int
	if before != nil && strings.TrimSpace(*before) != "" {
		idx, err := DecodeCursor(*before)
		if err != nil {
			return nil, err
		}
		beforeOffset = &idx
	}

	win := &RelayWindow{
		StartOffset:  startOffset,
		BeforeOffset: beforeOffset,
		Limit:        defaultPageLimit,
	}

	if last != nil {
		win.IsBackward = true
		win.Limit = clampRequestedLimit(*last)
		win.ExplicitZero = (*last == 0)
		return win, nil
	}
	if first != nil {
		win.Limit = clampRequestedLimit(*first)
		win.ExplicitZero = (*first == 0)
	}
	return win, nil
}

func clampRequestedLimit(requested int) int {
	if requested <= 0 {
		return 0
	}
	if requested > maxPageLimit {
		return maxPageLimit
	}
	return requested
}

// FetchRelayPage mengeksekusi paginasi Relay forward maupun backward terhadap endpoint gRPC berbasis limit/offset.
func FetchRelayPage[T any, P interface{ GetTotal() int64 }](
	win *RelayWindow,
	fetch func(limit, offset int32) ([]T, P, error),
) ([]T, int, *model.PageInfo, int, error) {
	if win == nil {
		win = &RelayWindow{Limit: defaultPageLimit}
	}

	if win.ExplicitZero {
		_, pageInfo, err := fetch(1, 0)
		if err != nil {
			return nil, 0, nil, 0, err
		}
		total := int(pageInfo.GetTotal())
		return nil, win.StartOffset, &model.PageInfo{
			StartCursor:     nil,
			EndCursor:       nil,
			HasPreviousPage: win.StartOffset > 0 && total > 0,
			HasNextPage:     total > win.StartOffset,
		}, total, nil
	}

	if win.IsBackward {
		var upperBound int
		var knownTotal int
		if win.BeforeOffset != nil {
			upperBound = *win.BeforeOffset
		} else {
			_, headPage, err := fetch(1, 0)
			if err != nil {
				return nil, 0, nil, 0, err
			}
			knownTotal = int(headPage.GetTotal())
			upperBound = knownTotal
		}

		offset := upperBound - win.Limit
		if offset < win.StartOffset {
			offset = win.StartOffset
		}
		limit := upperBound - offset
		if limit <= 0 {
			if win.BeforeOffset != nil {
				_, headPage, err := fetch(1, 0)
				if err != nil {
					return nil, 0, nil, 0, err
				}
				knownTotal = int(headPage.GetTotal())
			}
			return nil, offset, &model.PageInfo{
				StartCursor:     nil,
				EndCursor:       nil,
				HasPreviousPage: offset > 0 && knownTotal > 0,
				HasNextPage:     upperBound < knownTotal,
			}, knownTotal, nil
		}

		limit32, offset32, boundErr := clampRelayBounds(limit, offset)
		if boundErr != nil {
			return nil, 0, nil, 0, boundErr
		}
		items, pageInfo, err := fetch(limit32, offset32)
		if err != nil {
			return nil, 0, nil, 0, err
		}
		total := max(int(pageInfo.GetTotal()), knownTotal)
		if upperBound > total {
			upperBound = total
		}
		if remaining := upperBound - offset; remaining <= 0 {
			items = nil
		} else if len(items) > remaining {
			items = items[:remaining]
		}
		return items, offset, buildRelayPageInfo(offset, len(items), total, win.BeforeOffset != nil && *win.BeforeOffset < total), total, nil
	}

	offset := win.StartOffset
	limit := win.Limit
	clampedByBefore := false
	if win.BeforeOffset != nil {
		if *win.BeforeOffset <= offset {
			_, headPage, err := fetch(1, 0)
			if err != nil {
				return nil, 0, nil, 0, err
			}
			total := int(headPage.GetTotal())
			return nil, offset, &model.PageInfo{
				StartCursor:     nil,
				EndCursor:       nil,
				HasPreviousPage: offset > 0 && total > 0,
				HasNextPage:     total > *win.BeforeOffset,
			}, total, nil
		}
		if offset+limit > *win.BeforeOffset {
			limit = *win.BeforeOffset - offset
			clampedByBefore = true
		}
	}

	limit32, offset32, boundErr := clampRelayBounds(limit, offset)
	if boundErr != nil {
		return nil, 0, nil, 0, boundErr
	}
	items, pageInfo, err := fetch(limit32, offset32)
	if err != nil {
		return nil, 0, nil, 0, err
	}
	total := int(pageInfo.GetTotal())
	if clampedByBefore && len(items) > limit {
		items = items[:limit]
	}
	return items, offset, buildRelayPageInfo(offset, len(items), total, win.BeforeOffset != nil && offset+len(items) < total), total, nil
}

// buildRelayConnection membantu seluruh domain membentuk Edge/Nodes/PageInfo/TotalCount tanpa duplikasi logika paginasi.
func buildRelayConnection[Proto any, GQL any, Edge any, Conn any](
	after, before *string,
	first, last *int,
	fetch func(limit, offset int32) ([]*Proto, *efv1.PageInfo, error),
	mapNode func(*Proto) *GQL,
	makeEdge func(*GQL, string) *Edge,
	makeConn func([]*Edge, []*GQL, *model.PageInfo, int) *Conn,
) (*Conn, error) {
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, err
	}
	items, startOffset, pageInfo, total, err := FetchRelayPage(win, fetch)
	if err != nil {
		return nil, err
	}
	edges := make([]*Edge, 0, len(items))
	nodes := make([]*GQL, 0, len(items))
	for i, item := range items {
		node := mapNode(item)
		if node == nil {
			continue
		}
		edges = append(edges, makeEdge(node, EncodeCursor(startOffset+i)))
		nodes = append(nodes, node)
	}
	return makeConn(edges, nodes, pageInfo, total), nil
}

func buildRelayPageInfo(offset, count, total int, hasNextOverride bool) *model.PageInfo {
	var startCursor, endCursor *string
	if count > 0 {
		sc := EncodeCursor(offset)
		ec := EncodeCursor(offset + count - 1)
		startCursor = &sc
		endCursor = &ec
	}
	hasPrev := offset > 0 && total > 0 && (offset < total || count > 0)
	hasNext := hasNextOverride || (offset+count < total)
	return &model.PageInfo{
		StartCursor:     startCursor,
		EndCursor:       endCursor,
		HasPreviousPage: hasPrev,
		HasNextPage:     hasNext,
	}
}

func clampRelayBounds(limit, offset int) (int32, int32, error) {
	if limit < 0 || limit > maxPageLimit {
		limit = defaultPageLimit
	}
	if offset < 0 || offset > maxPaginationOffset || offset > math.MaxInt32 {
		return 0, 0, middleware.ErrInvalidArgument("pagination offset exceeds safe limit", nil)
	}
	return int32(limit), int32(offset), nil
}
