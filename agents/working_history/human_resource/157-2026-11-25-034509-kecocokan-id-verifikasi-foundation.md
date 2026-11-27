# Kecocokan ID referensi Enterprise Foundation

Adapter HRM memeriksa kesesuaian ID balasan dengan ID permintaan untuk badan hukum, unit organisasi, lokasi, dan party. Balasan berisi record aktif dengan ID yang berbeda menghasilkan konflik referensi sehingga tidak dapat dipakai sebagai bukti untuk transaksi HRM.

Tes bufconn dengan interceptor HMAC dan server fixture membuktikan penolakan pada keempat jenis referensi. Tes adapter dan cmd dengan race detector serta go vet lulus. Server Foundation aktual tidak dipakai dalam tes ini.
