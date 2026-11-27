# Regresi gRPC offboarding pada PostgreSQL

`go test -race -count=1 -tags=integration -v ./transport/grpc/handler/offboarding/test` lulus pada database PostgreSQL disposable. Mode fixture payroll diaktifkan agar kasus clearance dan terminasi ikut diuji.

| Kasus | Hasil |
| --- | --- |
| Penyelesaian tugas sebelum rencana berjalan atau sebelum waktu persetujuan | Ditolak |
| Referensi konfirmasi tugas tidak cocok | Ditolak |
| Tanggal terminasi pekerja tidak cocok | Ditolak |
| Konfirmasi tugas yang cocok dengan fixture | Tugas tersimpan sebagai `completed` dengan waktu dan referensi |
| Penutupan rencana dengan fixture payroll dan verifier clearance | Cakupan integrasi gRPC dan basis data lulus |

Verifier dan otorisasi gRPC memakai fixture. Bukti end-to-end melalui HTTPS untuk pemilik konfirmasi dan clearance masih terbuka.

## Sumber konfirmasi yang harus tersedia

Startup standar belum menyediakan sumber konfirmasi tugas HRM. Integrasi perlu mengikat referensi konfirmasi pada satu tugas, rencana, management pemilik, dan principal yang tercatat. Referensi tugas di fixture hanya dipakai untuk menguji transisi. Katalog atau sumber bukti produksi masih perlu ditetapkan sesuai jenis tugas HRM. Tugas milik IAM atau Finance memerlukan bukti dari service pemilik. [Handoff offboarding](../../grpc/hrm/13-operasi-offboarding.md).
