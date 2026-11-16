# Struktur folder pra-development HRM

Disiapkan folder lapisan domain, repository, services, contract, transport gRPC, konfigurasi, entrypoint, integrasi dan pengujian. Sebanyak 22 subdomain bisnis menggabungkan HRM, recruitment dan Talent Management dalam satu service. Inbox/outbox mempunyai direktori persistensi dan pengiriman tersendiri.

README memetakan seluruh 86 tabel DBML tepat sekali, menjelaskan hubungan worker–IAM, transaksi promosi, batas Foundation/Finance/DRM dan urutan development. Folder kosong memakai `.gitkeep` agar tersimpan di Git.

Verifikasi: struktur direktori, kelengkapan/keunikan pemetaan tabel, tautan relatif dan `git diff --check`. Tidak ada perubahan Go, DBML, SQL, workspace, Gateway atau deployment. Build/unit test/E2E tidak dijalankan karena perubahan hanya folder dan dokumentasi. Runtime dan kontrak API HRM masih belum diimplementasikan.
