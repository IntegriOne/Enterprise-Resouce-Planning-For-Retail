// Package address mengimplementasikan repositori PostgreSQL berbasis GORM untuk tabel addresses.
package address

import "time"

type addressModel struct {
	ID               string    `gorm:"column:id;type:uuid;primaryKey"`
	Label            *string   `gorm:"column:label"`
	AddressLine1     string    `gorm:"column:address_line_1"`
	AddressLine2     *string   `gorm:"column:address_line_2"`
	City             string    `gorm:"column:city"`
	Province         *string   `gorm:"column:province"`
	PostalCode       *string   `gorm:"column:postal_code"`
	CountryCode      string    `gorm:"column:country_code"`
	FormattedAddress *string   `gorm:"column:formatted_address"`
	Latitude         *float64  `gorm:"column:latitude"`
	Longitude        *float64  `gorm:"column:longitude"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (addressModel) TableName() string {
	return "addresses"
}
