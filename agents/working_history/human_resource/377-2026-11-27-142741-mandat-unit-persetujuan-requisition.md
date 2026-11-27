# Mandat unit untuk persetujuan requisition

Policy HRM kini memerlukan grant IAM khusus unit saat approve. Pemeriksaan awal terjadi sebelum pembacaan, kemudian unit diperiksa terhadap record yang terkunci dan sebelum commit. Grant global, unit lain, dan sesi tanpa MFA ditolak. Race test policy dan rekrutmen serta go vet lulus. Pengecekan ulang mandat saat pembukaan dan verifier anggaran masih terbuka.
