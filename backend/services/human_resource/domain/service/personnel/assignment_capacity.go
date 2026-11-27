package service

import (
	"math/big"
	"sort"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// CheckAssignmentCapacity memeriksa FTE penempatan yang disetujui sepanjang periode inklusif.
// ID kandidat dikecualikan agar persetujuan atau penggantian penempatan tidak menghitungnya dua kali.
func CheckAssignmentCapacity(candidate *entity.WorkerAssignment, existing []entity.WorkerAssignment, effectiveDate time.Time, endDate *time.Time, limit *big.Rat) error {
	total, err := candidate.FTEPercent.Rat()
	if err != nil {
		return hrerrors.ErrInvalidInput
	}
	// Akumulasi perubahan pada hari yang sama sebelum memeriksa kapasitas.
	deltas := make(map[time.Time]*big.Rat)
	addDelta := func(day time.Time, delta *big.Rat) {
		year, month, date := day.Date()
		key := time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
		if deltas[key] == nil {
			deltas[key] = new(big.Rat)
		}
		deltas[key].Add(deltas[key], delta)
	}
	for _, row := range existing {
		if row.CancelledAt != nil || row.ApprovedAt == nil || row.ID == candidate.ID ||
			(row.ValidTo != nil && entity.CompareDates(*row.ValidTo, effectiveDate) < 0) ||
			(endDate != nil && entity.CompareDates(row.ValidFrom, *endDate) > 0) {
			continue
		}
		fte, err := row.FTEPercent.Rat()
		if err != nil {
			return hrerrors.ErrReferenceConflict
		}
		if entity.CompareDates(row.ValidFrom, effectiveDate) <= 0 {
			total.Add(total, fte)
		} else {
			addDelta(row.ValidFrom, fte)
		}
		if row.ValidTo != nil {
			addDelta(row.ValidTo.AddDate(0, 0, 1), new(big.Rat).Neg(fte))
		}
	}
	if total.Cmp(limit) > 0 {
		return hrerrors.ErrReferenceConflict
	}
	dates := make([]time.Time, 0, len(deltas))
	for date := range deltas {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	for _, date := range dates {
		if endDate != nil && entity.CompareDates(date, *endDate) > 0 {
			break
		}
		total.Add(total, deltas[date])
		if total.Cmp(limit) > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
