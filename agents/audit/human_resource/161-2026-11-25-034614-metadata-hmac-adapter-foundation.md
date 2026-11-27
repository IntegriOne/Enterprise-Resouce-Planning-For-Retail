# Metadata HMAC pada koneksi adapter Foundation

Jalur Dial sebelumnya menandatangani panggilan di Client dan interceptor koneksi, sehingga proof dan request ID dikirim dua kali. Penandatanganan dipusatkan pada adapter yang juga dipakai oleh NewClient. Pengujian memakai Dial melalui TCP loopback ke server fixture yang memverifikasi HMAC dan mensyaratkan tepat satu proof dan satu request ID.

Tes adapter dan cmd dengan race detector serta go vet lulus. Kontrak Foundation saat ini hanya mengekspor status aktif unit dan lokasi, tanpa periode efektivitas keduanya. Pengujian ini tidak memakai data atau runtime Foundation produksi.
