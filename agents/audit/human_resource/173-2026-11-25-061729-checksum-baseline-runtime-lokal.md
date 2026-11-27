# Checksum baseline runtime HRM lokal

Kontainer HRM lokal sehat. Perbandingan baca saja antara checksum versi 1 yang tercatat pada database development `human_resource` dan SHA-256 baseline `000001_hrm.up.sql` pada checkout terbaru menunjukkan nilai berbeda.

Migration runner menolak mismatch pada startup. Dengan demikian health check kontainer yang sedang berjalan membuktikan kesehatan image aktif, sedangkan penerapan baseline terbaru sudah dibuktikan pada PostgreSQL sementara. Upgrade database development perlu mempertahankan data atau memakai database pengganti yang diinisialisasi dari baseline terbaru. [Migration runner](../../../backend/services/human_resource/db/migrate.go)

Binary `./cmd` dari checkout terbaru dikompilasi dan dijalankan pada kontainer sementara dengan database PostgreSQL khusus. Startup memasang baseline, menyambung ke Redis TLS, lalu menjawab `SERVING` melalui mTLS. Checksum yang tercatat cocok dengan baseline checkout. Kontainer dan database sementara dibersihkan setelah tes. Kontainer development dan database `human_resource` yang sudah berjalan tetap dipertahankan.
