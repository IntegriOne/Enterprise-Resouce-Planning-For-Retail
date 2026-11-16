package repository

// Filter berasal dari service/policy tepercaya; nama kolom tetap diverifikasi repository.
// Beberapa filter digabung AND. Nilai kosong ditolak, bukan diartikan sebagai akses global.
type Filter struct {
	Column string
	Values []any
}
