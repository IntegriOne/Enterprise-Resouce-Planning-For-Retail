# Uji HTTPS prasyarat tanda tangan kontrak dan aktivasi pekerja

Superadmin fixture memakai sesi TOTP, CSRF valid, dan role IAM sementara dengan grant pencatatan tanda tangan serta aktivasi pekerja. Pekerja preboarding dan draf kontrak sudah tersimpan pada database disposable. Tanggal masuk dan tanggal efektif kontrak sudah tercapai.

| Kasus | Hasil |
| --- | --- |
| `hrmRecordSignedEmploymentContract` dengan ID draf yang valid dan UUID versi sintetis ketika verifier DRM belum terpasang | `UNAVAILABLE` |
| `hrmActivateWorker` tanpa kontrak yang telah ditandatangani | `CONFLICT` |
| Detail kontrak melalui HTTPS dan pemeriksaan PostgreSQL | Status `draft`, waktu tanda tangan kosong |
| Detail pekerja melalui HTTPS dan pemeriksaan PostgreSQL | Status `preboarding` |

Database IAM, Foundation, HRM, dan Redis sementara telah dibersihkan setelah runtime development kembali sehat. [Uji draf kontrak](101-hrm-query-mutation-draf-kontrak-mfa.md) · [Handoff verifier DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md).
