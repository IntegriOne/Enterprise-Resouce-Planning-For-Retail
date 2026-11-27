// Package organizationunit mengimplementasikan repositori PostgreSQL berbasis GORM untuk tabel organization_units.
package organizationunit

import "time"

type organizationUnitModel struct {
	ID                string     `gorm:"column:id;type:uuid;primaryKey"`
	LegalEntityID     string     `gorm:"column:legal_entity_id;type:uuid"`
	ParentUnitID      *string    `gorm:"column:parent_unit_id;type:uuid"`
	PrimaryLocationID *string    `gorm:"column:primary_location_id;type:uuid"`
	UnitCode          string     `gorm:"column:unit_code"`
	Name              string     `gorm:"column:name"`
	Kind              string     `gorm:"column:unit_kind"`
	CostCenterCode    *string    `gorm:"column:cost_center_code"`
	HeadPersonID      *string    `gorm:"column:head_person_id;type:uuid"`
	IsActive          bool       `gorm:"column:is_active"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
}

func (organizationUnitModel) TableName() string {
	return "organization_units"
}
