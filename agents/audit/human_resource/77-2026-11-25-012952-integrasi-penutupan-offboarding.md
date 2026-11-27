# Integrasi penutupan offboarding

Tes gRPC dengan PostgreSQL disposable kini melewati rangkaian draf, persetujuan, mulai, penyelesaian dan waiver tugas, hingga penutupan rencana. Fixture payroll terminasi melewati constraint, rekonsiliasi baris hasil, dan trigger persetujuan PostgreSQL. Clearance yang ditolak verifier tiruan tidak menutup rencana. Clearance yang diterima menutup rencana, menyimpan referensi payroll dan waktu penutupan, dan menambah tepat satu event outbox. Pembacaan ulang mengonfirmasi status tersimpan. Permintaan ulang ditolak. Jalur tanpa verifier tetap diuji ketika mode fixture payroll tidak diaktifkan.

Tes integrasi Offboarding dengan race detector lulus. Clearance Finance dan IAM di produksi masih memerlukan kontrak otoritatif.
