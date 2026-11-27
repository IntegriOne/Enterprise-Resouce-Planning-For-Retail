# Klasifikasi kegagalan verifier onboarding dan offboarding

Tes onboarding mereproduksi error transport verifier yang diteruskan mentah dan dipetakan ke status gRPC Internal. Tiga command penyelesaian kini mengklasifikasikan error pemilik melalui aturan yang sama. Penolakan akses tetap menjadi PermissionDenied, referensi tidak ditemukan atau bertentangan menjadi FailedPrecondition, dan kegagalan dependency lain menjadi Unavailable. Pembatalan dan tenggat permintaan tetap mengikuti context transaksi. Tes onboarding memastikan kegagalan verifier tidak menyelesaikan tugas atau menambah outbox.
