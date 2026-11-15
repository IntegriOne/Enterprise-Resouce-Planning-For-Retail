// Package permission mengimplementasikan repositori PostgreSQL GORM untuk katalog izin atomik pada tabel permissions.
package permission

import "time"

type permissionModel struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey"`
	Code        string    `gorm:"column:code"`
	Resource    string    `gorm:"column:resource"`
	Action      string    `gorm:"column:action"`
	Description *string   `gorm:"column:description"`
	IsSensitive bool      `gorm:"column:is_sensitive"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (permissionModel) TableName() string {
	return "permissions"
}
