# Penugasan permintaan HR

AssignServiceRequest memungkinkan petugas manusia dengan izin step-up mengambil permintaan open yang belum ditugaskan. ID petugas berasal dari actor terverifikasi. Penugasan mengubah status ke in_progress dan mencatat outbox atomik. Create dan Update generik menolak pengisian atau pergantian PIC. Penugasan ke petugas lain memerlukan workflow dan pemeriksaan wewenang tambahan.

Integrasi gRPC ke PostgreSQL membuktikan penugasan lewat CRUD ditolak, penugasan melalui RPC mencatat actor, pengulangan ditolak, dan tepat satu event ditambahkan. Policy IAM memakai stub. Jalur resolusi berbukti dan penutupan masih terbuka.
