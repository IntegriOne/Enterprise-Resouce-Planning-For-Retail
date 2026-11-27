// Package service menyediakan kebijakan domain murni dan port layanan abstrak Enterprise Foundation tanpa ketergantungan ORM atau transport.
package service

import "time"

// Clock menyediakan sumber waktu terabstraksi agar logika domain tetap deterministik saat diuji.
type Clock interface {
	Now() time.Time
}

// SystemClock mengimplementasikan Clock menggunakan waktu lokal sistem operasi.
type SystemClock struct{}

// Now mengembalikan waktu lokal sistem operasi saat ini melalui time.Now().
func (SystemClock) Now() time.Time {
	return time.Now()
}

// IDGenerator menghasilkan pengenal UUID baru yang unik untuk entitas domain maupun event outbox.
type IDGenerator interface {
	NewUUID() (string, error)
}
