# Kalkulasi payroll gRPC

`CalculatePayrollRun` menjalankan adapter kalkulator, memeriksa ikatan hasil pada run, rule set, dan snapshot input, serta menyimpan hasil dan komponen secara atomik. Tes gRPC–PostgreSQL dengan kalkulator tiruan mencakup izin penginisiasi, penolakan digest, hasil tersimpan, dan pengulangan. Go test dan go vet lulus. Mesin kalkulasi produksi masih diperlukan.
