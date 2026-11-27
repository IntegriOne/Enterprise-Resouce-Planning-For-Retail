# Pengikatan snapshot pada penerbitan slip

HRM menyertakan ID snapshot input dan kedua digest saat meminta verifikasi penerbitan. Regresi PostgreSQL dengan verifier fixture menolak bukti yang mengacu pada ID atau digest input berbeda sebelum penyimpanan PDF. Race detector lulus.
