# Draf review kinerja mandiri HRM

CRUD draf review `self` diaktifkan setelah memeriksa worker aktif, link principal IAM, dan siklus terbuka. Reviewer ditetapkan server; role lain, rating, submit/finalisasi, dan perubahan review yang telah direferensikan tetap ditolak. Read/list/create/update memerlukan step-up. Tes unit, lima pengulangan gRPC–PostgreSQL dengan race detector, fmt/vet/build, dan Docker health lulus. Validator kini 25/78 resource; IAM/Gateway E2E serta workflow review lengkap belum selesai.
