# Aktivasi lokal HRM pada kedua replika

Sertifikat klien Gateway khusus HRM diterbitkan dari CA development. Kunci dan kunci HMAC HRM disimpan pada bind mount lokal terabaikan Git yang terbaca UID 10001. Health gRPC HRM menerima sertifikat baru. Build dan startup kedua replika Gateway dengan HRM aktif lulus. Query anonim menghasilkan UNAUTHENTICATED pada kedua port Gateway serta HTTPS Caddy. Query positif dengan grant IAM tetap terbuka.
