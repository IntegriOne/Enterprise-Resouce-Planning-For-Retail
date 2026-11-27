# Evaluasi scope mandiri untuk slip gaji

Pengenal kode izin bersama kini mengenali `human_resource.payslips.read_self` sebagai izin mandiri. Snapshot menolak target global dan target principal lain, termasuk saat pemanggil memegang grant global. Gateway dan gRPC HRM tetap meminta grant `SELF` eksplisit. Tes authz, domain dan sesi IAM, adapter Gateway, serta policy HRM lulus. Grant produksi dan uji slip dengan bukti Finance serta DRM masih terbuka.
