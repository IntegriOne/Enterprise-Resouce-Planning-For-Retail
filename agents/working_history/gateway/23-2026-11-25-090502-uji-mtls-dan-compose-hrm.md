# Uji mTLS dan Compose HRM

Tes TCP membuktikan TLS 1.3 dua arah, validasi nama server, penerusan hash sesi, dan verifikasi HMAC caller. Kedua replika Gateway mendapat opt-in HRM, target, dan bind mount privat read-only. Direktori lokal kredensial Gateway dan HRM dikecualikan dari build context. Konfigurasi Compose default/opt-in, build image Gateway, race test transport dan go vet lulus. Replika berjalan masih perlu diuji dengan deployment sertifikat yang sesuai.
