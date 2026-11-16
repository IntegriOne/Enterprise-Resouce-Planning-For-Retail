# Riwayat kerja HRM 51

Command dan RPC `CompleteOnboardingPlan` menutup plan berjalan setelah seluruh tugas wajib tercatat selesai. Service mengunci plan, memeriksa tugas dalam transaksi, mengisi actor/waktu server dan outbox atomik. Validator menjaga kolom immutable; izin `human_resource.onboarding_plans.complete` bersifat sensitif. Tes unit, mock, race, dan gRPC→PostgreSQL disposable lulus; build dan restart Docker HRM sehat. Fixture tugas selesai bukan verifikasi DRM. `CompleteOnboardingTask`, bukti versi DRM/PIC, grant IAM, dan E2E Gateway masih terbuka. Detail: audit HRM nomor 35.
