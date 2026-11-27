# Dependensi command HRM di runtime produksi

Composition root `internal/service/registry.go` belum memasang port berikut. Bukti dari tes dengan verifier tiruan tidak mengaktifkan operasi pada server yang dibangun melalui composition root tersebut.

| Command | Dependensi yang diperlukan | Dampak |
| --- | --- | --- |
| `ResolveServiceRequest` | DRM membuktikan versi immutable yang terikat ke tiket dan pekerja | Resolusi mengembalikan `Unavailable`. |
| `CompleteOnboardingTask` | DRM membuktikan versi bukti tugas | Penyelesaian tugas mengembalikan `Unavailable`. |
| `PublishJobPosting` | DRM membuktikan versi dan digest konten | Publikasi mengembalikan `Unavailable`. |
| `CompleteApplicationAssessment` | DRM membuktikan versi dan digest bukti asesmen | `Unavailable` hanya bila bukti dokumen disertakan. |
| `ApproveWorkforcePlan` | Finance membuktikan anggaran untuk snapshot rencana dan seluruh baris | Persetujuan mengembalikan `Unavailable`. |
| `ApproveRequisition` | Verifier mandat unit, headcount, dan anggaran | Persetujuan mengembalikan `Unavailable`. |
| `CompleteOffboardingTask` dan `CompleteOffboardingPlan` | Konfirmasi penyelesaian dari pemilik tugas dan keputusan clearance Finance/IAM untuk finalisasi | Command penyelesaian menunggu verifier terkait. |

Rujukan kebutuhan integrasi tersedia pada [catatan DRM](../../notes/document_records_management/1-human_resource.md), [resolusi tiket](../../notes/document_records_management/2-human_resource-hr-service-resolution.md), dan [anggaran rencana tenaga kerja](../../notes/finance_accounting/2-human_resource-workforce-budget.md). Pemeriksaan ini terbatas pada command dan port di atas serta tidak mengubah aturan bisnis.
