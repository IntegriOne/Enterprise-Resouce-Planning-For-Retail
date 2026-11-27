// Package location mengimplementasikan repositori PostgreSQL berbasis GORM untuk tabel locations.
package location

import "time"

type locationModel struct {
	ID                    string     `gorm:"column:id;type:uuid;primaryKey"`
	LegalEntityID         string     `gorm:"column:legal_entity_id;type:uuid"`
	ParentLocationID      *string    `gorm:"column:parent_location_id;type:uuid"`
	AddressID             *string    `gorm:"column:address_id;type:uuid"`
	LocationCode          string     `gorm:"column:location_code"`
	Name                  string     `gorm:"column:name"`
	Kind                  string     `gorm:"column:location_kind"`
	Phone                 *string    `gorm:"column:phone"`
	Timezone              string     `gorm:"column:timezone"`
	OperationalAttributes string     `gorm:"column:operational_attributes;type:jsonb"`
	IsActive              bool       `gorm:"column:is_active"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
	DeletedAt             *time.Time `gorm:"column:deleted_at"`
}

func (locationModel) TableName() string {
	return "locations"
}
