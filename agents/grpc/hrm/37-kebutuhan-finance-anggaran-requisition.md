# List operasi

| Operasi HRM | Kebutuhan dari Finance |
| --- | --- |
| `ApproveRequisition` | Verifikasi otorisasi anggaran untuk requisition dan rincian kebutuhan terkunci. Kembalikan UUID otorisasi. |
| `OpenRequisition` | Periksa ulang otorisasi dengan UUID yang sudah disimpan saat approval. Kembalikan UUID yang sama jika otorisasi tetap berlaku dan cocok. |

## Penjelasan ApproveRequisition

HRM memeriksa mandat IAM peninjau untuk unit yang tepat sebelum meminta keputusan Finance. HRM mengirim ID requisition, tipe kebutuhan, unit, jabatan, jumlah kebutuhan, kategori pekerja, tanggal target, posisi, lokasi kerja, lini rencana tenaga kerja, serta principal yang terikat pada proses. Referensi opsional yang kosong tetap kosong dalam snapshot. Approval pertama mengirim `ExpectedBudgetReferenceID` kosong. Finance perlu mengambil keputusan dari catatan anggarannya sendiri dan mengembalikan UUID stabil yang terikat pada requisition dan snapshot tersebut.

## Penjelasan OpenRequisition

HRM membaca ulang mandat IAM milik peninjau tercatat. Setelah mandat lolos, HRM mengirim ulang snapshot beserta `ExpectedBudgetReferenceID`. Finance perlu membuktikan bahwa referensi itu masih aktif dan tetap berlaku untuk snapshot yang sama. Referensi pengganti tidak membuka requisition lama secara otomatis. HRM membandingkan UUID hasil verifikasi dengan UUID yang tersimpan sebelum menyimpan perubahan status.

## Butuh GraphQL Mutation

Gateway sudah menyediakan `approveRequisition(id: ID!)` dan `openRequisition(id: ID!)`. Integrasi Finance berlangsung melalui kontrak privat antarlayanan. UI mengirim ID requisition, sedangkan UUID otorisasi berasal dari verifier Finance.

## RBAC

Approval memakai `human_resource.requisitions.approve` berscope `ORGANIZATION_UNIT` sesuai unit record dan membutuhkan MFA terkini. Pembukaan memerlukan `human_resource.requisitions.open` GLOBAL dengan MFA. HRM memeriksa ulang mandat peninjau tercatat. Pemeriksa Finance perlu mengautentikasi workload HRM sebelum menerima permintaan.

## Kondisi yang perlu disinkronkan

- Finance menetapkan cara menemukan otorisasi awal, identitas workload yang dipercaya, versi kontrak, serta kriteria status aktif, pencabutan, dan kedaluwarsa.
- Adapter HRM perlu membedakan otorisasi yang hilang, dicabut, kedaluwarsa, atau tidak sesuai snapshot dari gangguan layanan. Status requisition dan outbox tetap utuh pada kegagalan.
- Pengulangan request untuk otorisasi dan snapshot yang sama harus mempertahankan UUID. Verifikasi tidak membuat pembayaran ataupun memindahkan dana.
- Uji integrasi perlu mencakup persetujuan dengan bukti sah, pembukaan dengan UUID yang sama, penolakan UUID pengganti dan snapshot berbeda, serta kegagalan komunikasi.

Kode HRM sudah mengirim referensi anggaran yang diharapkan pada pembukaan. Adapter produksi Finance masih menunggu sinkronisasi. [Kontrak HRM](../../../backend/services/human_resource/internal/service/recruitment/requisition_approval.go) · [Handoff approval dan IAM](35-kebutuhan-persetujuan-requisition.md).
