package resolvers

import (
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	efadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/enterprise_foundation"
	hradapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/human_resource"
)

func hrConnection[T, E any](
	after, before *string, first, last *int,
	fetch func(limit, offset int) (*hradapter.Page[T], error),
	makeEdge func(*T, string) *E,
) ([]*E, []*T, *model.PageInfo, int, error) {
	window, err := efadapter.ParseRelayPage(after, before, first, last)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	nodes, offset, info, total, err := efadapter.FetchRelayPage(window,
		func(limit, offset int32) ([]*T, *hradapter.Page[T], error) {
			page, err := fetch(int(limit), int(offset))
			if err != nil {
				return nil, nil, err
			}
			return page.Items, page, nil
		})
	if err != nil {
		return nil, nil, nil, 0, err
	}
	edges := make([]*E, len(nodes))
	for i, node := range nodes {
		edges[i] = makeEdge(node, efadapter.EncodeCursor(offset+i))
	}
	if nodes == nil {
		nodes = []*T{}
	}
	return edges, nodes, info, total, nil
}
