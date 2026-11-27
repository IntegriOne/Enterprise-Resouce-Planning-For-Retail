# Verifikasi tautan akun HRM

Commit `d6d36eb` di `origin/main` menambahkan client mTLS yang mencocokkan pekerja, principal, dan badan hukum. Client membedakan tautan yang ditolak dari ketidaktersediaan HRM. Service provisioning lokal kini mewajibkan konfirmasi langsung sebelum memperbarui status. Unit test membuktikan bahwa verifier yang tidak tersedia atau konflik tidak mengubah status. Tes race dan vet lulus. Runtime provisioning dan uji HTTPS dari main belum selesai.
