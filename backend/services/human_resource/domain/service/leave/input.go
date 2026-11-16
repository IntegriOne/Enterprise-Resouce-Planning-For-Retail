package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// BalanceInput membawa ledger lengkap serta bukti adjustment yang sudah diputuskan.
type BalanceInput struct {
	Entitlement  *entity.LeaveEntitlement
	Entries      []entity.LeaveLedgerEntry
	Adjustments  []entity.LeaveAdjustment
	CalculatedAt time.Time
}
