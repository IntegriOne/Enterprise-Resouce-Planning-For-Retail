# GetMyPayslip melalui gRPC dan PostgreSQL

Tes menggunakan verifier Finance dan DRM berupa fixture serta database sementara dengan race detector.

| Kasus | Hasil |
| --- | --- |
| Slip tertaut pada run posted sebelum konfirmasi penerbitan Finance | FailedPrecondition |
| Pekerja membaca slip sendiri setelah penerbitan dengan transfer terjadwal | Metadata slip, referensi versi dokumen, dan status scheduled |
| Pekerja mengirim ID hasil milik pekerja lain yang juga memiliki dokumen tertaut | NotFound |

Tes policy HRM mengizinkan scope self yang cocok dengan principal sesi serta menolak scope self milik principal lain dan grant global untuk operasi ini.

Regresi gRPC menolak timestamp penerbitan di masa depan dengan FailedPrecondition dan menerima penerbitan yang sudah berlaku.
