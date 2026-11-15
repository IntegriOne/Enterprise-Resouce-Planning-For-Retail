package iam

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
	cursorPrefix     = "v1:"
)

// RelayWindow menyimpan parameter jendela paginasi Relay yang telah divalidasi dari argumen after/before/first/last.
type RelayWindow struct {
	StartOffset  int
	BeforeOffset *int
	Limit        int
	IsBackward   bool
	ExplicitZero bool
}

// ParseRelayPage memvalidasi argumen cursor Relay (after, before, first, last) secara ketat.
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
		if *last == 0 {
			win.Limit = 0
			win.ExplicitZero = true
		} else if *last > maxPageLimit {
			win.Limit = maxPageLimit
		} else {
			win.Limit = *last
		}
		return win, nil
	}

	if first != nil {
		if *first == 0 {
			win.Limit = 0
			win.ExplicitZero = true
		} else if *first > maxPageLimit {
			win.Limit = maxPageLimit
		} else {
			win.Limit = *first
		}
	}
	return win, nil
}

// EncodeCursor mengkodekan indeks posisi absolut menjadi opaque cursor Relay berbasis base64url.
func EncodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(cursorPrefix + strconv.Itoa(offset)))
}

// DecodeCursor mengurai opaque cursor Relay menjadi indeks posisi non-negatif.
func DecodeCursor(cursor string) (int, error) {
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return 0, middleware.ErrInvalidArgument("empty pagination cursor", nil)
	}
	raw, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return 0, middleware.ErrInvalidArgument("invalid pagination cursor encoding", err)
	}
	decoded := string(raw)
	if !strings.HasPrefix(decoded, cursorPrefix) {
		return 0, middleware.ErrInvalidArgument("invalid pagination cursor format", nil)
	}
	idx, err := strconv.Atoi(strings.TrimPrefix(decoded, cursorPrefix))
	if err != nil || idx < 0 || idx > maxPaginationOffset {
		return 0, middleware.ErrInvalidArgument("invalid pagination cursor index", err)
	}
	return idx, nil
}

// FetchRelayPage mengeksekusi paginasi Relay forward maupun backward ke RPC backend berbasis limit/offset.
func FetchRelayPage[T any](
	win *RelayWindow,
	fetch func(limit, offset int32) ([]T, *iamv1.PageInfo, error),
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

	// Paginasi mundur (`last`) membutuhkan penentuan ujung atas; bila `before` absen, ambil total dahulu.
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
		total := int(pageInfo.GetTotal())
		if knownTotal > total {
			total = knownTotal
		}
		if upperBound > total {
			upperBound = total
		}
		remaining := upperBound - offset
		if remaining <= 0 {
			items = nil
		} else if len(items) > remaining {
			items = items[:remaining]
		}
		return items, offset, buildRelayPageInfo(offset, len(items), total, win.BeforeOffset != nil && *win.BeforeOffset < total), total, nil
	}

	// Paginasi maju (`first` atau default).
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

// SliceRelayWindow mempaginasi slice penuh di memori sesuai spesifikasi cursor Relay (after, before, first, last).
func SliceRelayWindow[T any](items []T, after, before *string, first, last *int) ([]T, int, *model.PageInfo, int, error) {
	win, err := ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, 0, nil, 0, err
	}
	total := len(items)
	start := win.StartOffset
	if start > total {
		start = total
	}
	end := total
	if win.BeforeOffset != nil && *win.BeforeOffset < end {
		end = *win.BeforeOffset
	}
	if end < start {
		end = start
	}

	if win.ExplicitZero {
		return nil, start, &model.PageInfo{
			StartCursor:     nil,
			EndCursor:       nil,
			HasPreviousPage: start > 0 && total > 0,
			HasNextPage:     end < total || start < total,
		}, total, nil
	}

	if win.IsBackward {
		if end-start > win.Limit {
			start = end - win.Limit
		}
	} else if end-start > win.Limit {
		end = start + win.Limit
	}

	sliced := items[start:end]
	return sliced, start, buildRelayPageInfo(start, len(sliced), total, false), total, nil
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
