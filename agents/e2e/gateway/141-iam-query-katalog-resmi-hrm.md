# Katalog resmi HRM melalui HTTPS

IAM development dibangun dari checkout dan lulus healthcheck. Superadmin login, membaca katalog berdasarkan resource melalui GraphQL, mencocokkan kode dan sensitivitas seluruh 63 izin HRM dari manifest, lalu logout. Runner keluar dengan status 0. Skenario ini hanya membaca katalog. Tidak ada assignment baru.

Tes sinkronisasi IAM dan tes pencocokan manifest dengan policy gRPC juga lulus. Status lima izin yang masih menunggu dependency dicatat pada [matriks izin](../../grpc/hrm/28-izin-iam-untuk-wiring-gateway.md).
