# Katalog compensation band melalui gRPC

Validator produksi `hr_compensation_bands` dipasang pada composition root HRM. Versi band dijaga unik; approval tanpa workflow ditolak; band yang telah dipakai catatan kompensasi dibekukan. Tes unit, race, vet, build, gRPC `bufconn` ke PostgreSQL sementara lima kali, dan build serta health Docker lulus. Cakupan validator menjadi 21/78 resource mutasi. Grant IAM, approval resmi, dan mutasi compensation record masih belum selesai.
