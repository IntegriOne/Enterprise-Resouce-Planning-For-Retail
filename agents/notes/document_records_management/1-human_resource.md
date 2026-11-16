# Kebutuhan verifikasi versi bukti onboarding dari HRM

HRM memiliki `hr_onboarding_tasks` dan command `CompleteOnboardingTask`. Referensi `completion_evidence_reference_id` harus menunjuk **versi dokumen DRM yang tepat**, immutable, tersedia untuk tujuan onboarding, dan boleh diakses oleh principal pelaksana. ID dari request atau keberadaan UUID saja tidak cukup.

Port HRM `EvidenceVerifier.VerifyOnboardingEvidence` mengirim `VersionID`, `TaskID`, dan `PrincipalID`; adapter DRM produksi belum tersedia. Kontrak owner perlu mengembalikan ID versi yang sama, status verifikasi dan sifat immutable, atau error terpisah untuk bukti tidak sah versus dependency tidak tersedia. Panggilan privat harus membawa workload identity/audience, deadline, purpose onboarding, scope pemilik bukti, serta hasil berversi atau masa berlaku yang jelas. Jika status berubah atau pemeriksaan gagal, HRM menolak sebelum menyimpan task/outbox. Tidak ada SQL/FK lintas database atau grant IAM otomatis.

Selama kontrak dan runtime DRM belum tersedia, registry HRM tidak memasang verifier sehingga RPC completion mengembalikan `Unavailable`. Tes HRM dengan verifier tiruan hanya membuktikan transaksi dan invariant HRM, bukan validasi dokumen DRM nyata.
