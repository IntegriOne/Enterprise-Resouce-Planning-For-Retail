# Selaraskan generator Payroll dan Succession

Aturan pembatalan payroll independen dan pasangan penerbitan–dokumen dimasukkan ke DBML agar regenerasi menjaga validasi yang berlaku. Baseline PostgreSQL turut memeriksa urutan waktu dan pemisahan tugas pembatalan payroll. Setelah regenerasi, 49 file yang dihasilkan cocok byte demi byte dengan checkout. Go vet, tes modul dengan race detector, regresi gRPC PostgreSQL Payroll dan Succession, dan verifikasi skema lengkap lulus.
