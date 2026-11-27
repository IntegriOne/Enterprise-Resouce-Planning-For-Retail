# Generator validasi CHECK HRM

Generator memahami btrim, coalesce dengan fallback teks, dan cast tanggal dari timestamp UTC yang dipakai DBML. Regenerasi terisolasi menghasilkan 87 model dan 153 CHECK. Validasi hasil regenerasi dipakai pada Personnel dan Offboarding untuk metadata pembatalan. File domain lainnya yang masih berbeda dari keluaran generator belum disinkronkan. Go vet, tes domain dan service dengan race detector, serta regresi gRPC PostgreSQL Offboarding lulus.
