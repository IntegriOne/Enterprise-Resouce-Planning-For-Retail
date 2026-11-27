# Paginasi validasi keputusan promosi

Validator pencatatan keputusan dan perubahan status rekomendasi kini membaca seluruh riwayat dengan verifikasi count dan batas 10.000 baris. Sebelumnya validator hanya membaca halaman pertama sehingga keputusan pada langkah berikutnya dapat terlewat.

Tes regresi menempatkan keputusan persetujuan pada baris 101 dan memastikan transisi berhasil. Tes security dan gRPC PostgreSQL sementara dengan race detector serta go vet lulus.
