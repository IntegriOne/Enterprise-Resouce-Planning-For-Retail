package human_resource

import (
	"math"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
)

func validPagination(limit, offset int) bool {
	return limit >= 1 && limit <= 100 && offset >= 0 && offset <= math.MaxInt32
}

// validPage memeriksa batas respons. HRM membaca item dan total secara terpisah.
func validPage(page *common.PageInfo, limit, offset, count int) bool {
	return page != nil && page.GetLimit() == int32(limit) && page.GetOffset() == int32(offset) &&
		page.GetTotal() >= 0 && page.GetTotal() <= math.MaxInt32 &&
		count >= 0 && count <= limit
}

// validPublishedPayslipPage memeriksa kontrak verifikasi slip terbit dengan total yang konsisten.
func validPublishedPayslipPage(page *common.PageInfo, limit, offset, count int) bool {
	return validPage(page, limit, offset, count) &&
		int64(count) == min(int64(limit), max(0, page.GetTotal()-int64(offset)))
}

// Page membawa hasil RPC limit/offset untuk penyusunan koneksi GraphQL.
type Page[T any] struct {
	Items      []*T
	TotalCount int
	Limit      int
	Offset     int
}

func (p *Page[T]) GetTotal() int64 {
	return int64(p.TotalCount)
}
