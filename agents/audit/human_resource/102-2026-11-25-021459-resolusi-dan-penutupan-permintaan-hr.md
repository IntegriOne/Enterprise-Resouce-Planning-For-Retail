# Resolusi dan penutupan permintaan HR

ResolveServiceRequest hanya menerima PIC yang memiliki izin step-up dan membawa referensi versi dokumen yang disahkan verifier DRM. Resolusi mencatat waktu server, bukti dan status atomik dengan outbox. CloseServiceRequest hanya menerima pengusul berizin sesudah resolusi dan mempertahankan metadata resolusi.

Integrasi gRPC ke PostgreSQL membuktikan dokumen yang ditolak tidak mengubah permintaan, resolusi sah tercatat, penutupan oleh PIC yang bukan pengusul ditolak, pengusul dapat menutup, dan penutupan ulang ditolak. Policy IAM dan DRM memakai stub. Bootstrap masih memerlukan verifier DRM otoritatif sebelum resolusi dapat dipakai di runtime produksi.
