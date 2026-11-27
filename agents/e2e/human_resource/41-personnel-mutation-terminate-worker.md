# TerminateWorker dan penutupan offboarding

Tes gRPC `bufconn` dengan PostgreSQL sementara dan race detector menjalankan `TerminateWorker` pada rencana berjalan. Respons memuat status terminated, tanggal efektif dan kode alasan dari rencana. Panggilan ulang ditolak dengan FailedPrecondition. Setelah kondisi payroll final dan checklist dipenuhi, `CompleteOffboardingPlan` berhasil memakai terminasi yang dicatat oleh RPC. Finance, IAM, dan verifier pemilik memakai fixture.

Tes menambahkan penempatan primer yang sedang berjalan, penempatan sekunder mendatang, dan dua kontrak bertanda tangan. Timestamp persetujuan penempatan yang melampaui waktu server menyebabkan FailedPrecondition dan penempatan tetap terbuka. Setelah timestamp persetujuan diselaraskan pada fixture, RPC menutup penempatan dan kontrak berjalan tepat pada hari kerja terakhir, lalu membatalkan penempatan dan kontrak mendatang. Query penempatan menggunakan gRPC dan membuktikan waktu pembatalan tersedia bersama jadwal dan persetujuan asli. Bukti tanda tangan kontrak tetap tersimpan.

Tes dengan transfer tertunda membuktikan bahwa approve setelah terminasi mengembalikan FailedPrecondition. Petugas berwenang kemudian dapat mencatat reject melalui RPC yang sama. Hasil berstatus rejected dengan referensi penempatan terapan kosong.
