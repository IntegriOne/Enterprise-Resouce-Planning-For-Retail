package entity

import "time"

// PayrollInputSnapshot menyimpan byte input dan cakupan yang dibekukan oleh HRM.
type PayrollInputSnapshot struct {
	ID               string    `gorm:"column:id;primaryKey"`
	PeriodID         string    `gorm:"column:payroll_period_id"`
	GroupID          string    `gorm:"column:payroll_group_id"`
	EmployerID       string    `gorm:"column:employing_legal_entity_reference_id"`
	JurisdictionCode string    `gorm:"column:jurisdiction_code"`
	CurrencyCode     string    `gorm:"column:currency_code"`
	Content          []byte    `gorm:"column:content"`
	Digest           []byte    `gorm:"column:digest"`
	CreatedBy        string    `gorm:"column:created_by_principal_id"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime:false"`
}

func (PayrollInputSnapshot) TableName() string { return "hr_payroll_input_snapshots" }
