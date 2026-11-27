# Penutupan rencana pengembangan

CompleteDevelopmentPlan menutup rencana aktif melalui mentor aktif yang menyetujuinya. Izin khusus memerlukan step-up. Rencana harus memiliki minimal satu tindakan dan seluruh tindakan harus selesai. Pemeriksaan berjalan dengan lock rencana yang juga diambil oleh mutasi tindakan. Status dan outbox tercatat atomik, dengan metadata approval tetap utuh.

Integrasi gRPC ke PostgreSQL membuktikan penutupan ditolak saat satu tindakan masih pending, berhasil sesudah penyelesaian, menolak pengulangan, dan mencatat tepat satu event tambahan. IAM menggunakan policy tiruan. Penyelesaian tindakan kursus dan bukti eksternal masih memerlukan integrasi pemilik.
