# Persetujuan rencana pengembangan

ApproveDevelopmentPlan mengaktifkan draf melalui mentor aktif yang principal IAM-nya cocok dengan actor dan memiliki izin step-up. Validator mengunci pekerja, mentor, dan jabatan target jika ada. Tanggal target tidak boleh berlalu. Hanya status, pelaku dan waktu approval berubah, bersama outbox atomik. Posisi dan assignment tidak berubah. Jalur ini memakai mentor yang tercatat pada rencana. Persetujuan manajer tanpa penugasan mentor dan penyelesaian tindakan setelah aktivasi masih terbuka.

Tes integrasi Career gRPC ke PostgreSQL membuktikan penolakan actor lain, persetujuan oleh mentor, dan perlindungan dari penghapusan. Policy memakai stub IAM.
