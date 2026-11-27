package eferrors

// Sentinel error domain untuk entitas master alamat dan koordinat geografis.
var (
	// ErrAddressNotFound menandakan data alamat tidak ditemukan.
	ErrAddressNotFound = newSentinel(KindNotFound, "enterprise_foundation.address.not_found", "address not found")
	// ErrAddressInUse menandakan alamat masih dirujuk oleh badan hukum, lokasi, atau mitra aktif.
	ErrAddressInUse = newSentinel(KindConflict, "enterprise_foundation.address.in_use", "address is still referenced")
	// ErrInvalidAddressLine menandakan baris alamat utama kosong atau melebihi batas panjang skema.
	ErrInvalidAddressLine = newSentinel(KindInvalidInput, "enterprise_foundation.address.invalid_line", "address line must be non-empty and within length bounds")
	// ErrInvalidCity menandakan nama kota kosong atau melebihi batas panjang skema.
	ErrInvalidCity = newSentinel(KindInvalidInput, "enterprise_foundation.address.invalid_city", "city must be non-empty and within length bounds")
	// ErrInvalidCountryCode menandakan kode negara tidak memenuhi format 2 huruf ISO 3166-1 alpha-2.
	ErrInvalidCountryCode = newSentinel(KindInvalidInput, "enterprise_foundation.address.invalid_country_code", "country code must be a 2-letter iso alpha-2 code")
	// ErrInvalidCoordinates menandakan nilai lintang atau bujur berada di luar rentang geografis valid.
	ErrInvalidCoordinates = newSentinel(KindInvalidInput, "enterprise_foundation.address.invalid_coordinates", "coordinates must be within valid latitude and longitude ranges")
	// ErrInvalidAddressUsage menandakan jenis peruntukan alamat tidak dikenali oleh skema.
	ErrInvalidAddressUsage = newSentinel(KindInvalidInput, "enterprise_foundation.address.invalid_usage", "invalid address usage kind")
)
