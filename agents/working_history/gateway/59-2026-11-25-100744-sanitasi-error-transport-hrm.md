# Sanitasi error transport HRM

Gateway tetap memetakan kegagalan autentikasi antarservice ke UNAVAILABLE untuk browser. Penyebab yang dapat dicatat pada log kini hanya berisi kode status gRPC, sehingga pesan transport mentah tidak ikut tercatat. Regresi adapter, race test dan vet Gateway lulus.
