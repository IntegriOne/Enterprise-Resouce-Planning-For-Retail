# Pengakhiran benefit lokal

EndBenefitEnrollment memakai izin dan step-up tersendiri. Actor manusia dapat mengakhiri enrollment lokal yang telah disetujui. Tanggal akhir harus berada dalam periode yang sah dan tidak melampaui hari UTC berjalan. Command menjaga bukti persetujuan dan menyimpan status ended serta event outbox secara atomik. Enrollment yang sudah berakhir tetap tersimpan.

Integrasi gRPC ke PostgreSQL lulus untuk penolakan enrollment pending dan tanggal mendatang, pengakhiran yang sah, retensi waktu persetujuan, serta penolakan pengakhiran ulang dan penghapusan. Policy IAM memakai stub. Kontrak provider dan kontribusi tetap memerlukan bukti pemilik.
