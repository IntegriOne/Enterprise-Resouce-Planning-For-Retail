# Validasi tenggat permintaan layanan HR

Tes regresi mereproduksi create dan update yang menerima tenggat sebelum waktu pembukaan tiket. Validator kini menolak nilai baru tersebut dengan kesalahan validasi. Tenggat historis dengan nilai yang sama masih bisa dipertahankan agar tiket lama dapat melanjutkan workflow. Tes race pada paket security, vet modul, dan diff check lulus.
