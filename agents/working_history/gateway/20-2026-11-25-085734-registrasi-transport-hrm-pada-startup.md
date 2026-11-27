# Registrasi transport HRM pada startup

Startup Gateway membaca konfigurasi mTLS HRM opsional, membuka koneksi nonblocking dan menutupnya saat shutdown. Kegagalan konfigurasi HRM dicatat secara terpisah dan tidak menghentikan IAM. Tes source connections dan config dengan race detector serta go vet lulus. Deployment sertifikat kedua replika dan schema/resolver masih perlu diselesaikan.
