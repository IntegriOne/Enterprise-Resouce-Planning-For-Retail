# Uji HTTPS guard cetak ulang slip

Jalankan `python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py` dari root. Runner membuat stack disposable dengan Gateway, IAM, HRM, dan Enterprise Foundation. Superadmin memberi dan mencabut grant badan hukum melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Grant badan hukum yang valid tanpa MFA | `AUTHORIZATION`, PDF kosong |
| Grant badan hukum yang valid dengan MFA dan ID Worker dari GraphQL | `UNAVAILABLE` karena adapter runtime Finance dan DRM belum terpasang, PDF kosong |

Percobaan awal menemukan Gateway memakai jenis GID yang tidak cocok dengan ID Worker keluarannya. Parser diperbaiki dan regresi Gateway dengan race detector serta runner HTTPS lulus. Uji ini belum membuktikan bahwa PDF terbit bisa dibaca melalui runtime Finance dan DRM.
