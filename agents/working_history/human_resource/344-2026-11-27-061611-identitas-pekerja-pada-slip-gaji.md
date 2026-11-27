# Identitas yang tercetak pada slip gaji

HRM mengambil nomor pekerja yang stabil ketika menerbitkan PDF. Renderer menampilkan nomor tersebut dan mempertahankan referensi teknis pada metadata integrasi DRM. Tes renderer dengan DejaVu Sans dan pdftotext membuktikan nomor mudah dibaca, referensi internal tidak tercetak, dan PDF tetap deterministik. Regresi gRPC payroll dengan PostgreSQL terisolasi dan race detector lulus.
