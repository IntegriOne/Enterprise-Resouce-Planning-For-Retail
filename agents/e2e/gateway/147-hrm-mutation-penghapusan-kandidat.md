# Penghapusan kandidat

Pengujian HTTPS memakai stack terisolasi. Superadmin memasang role dan izin lewat GraphQL. Kandidat dibuat melalui GraphQL. Lamaran dan requisition disiapkan sebagai fixture PostgreSQL untuk kandidat yang harus dipertahankan.

| Skenario | Hasil |
| --- | --- |
| Hapus kandidat dengan izin tetapi tanpa MFA | Ditolak |
| Hapus kandidat yang memiliki lamaran setelah MFA | `CONFLICT` |
| Hapus kandidat tanpa lamaran setelah MFA | `deleted: true` |
| Pemeriksaan PostgreSQL | Hanya kandidat uji tanpa lamaran terhapus. Kandidat lain dan lamarannya tetap tersimpan. |

Runner selesai dengan status 0 dan menghapus database uji.

IAM dan kedua replika Gateway development sehat setelah rollout. Uji baca pekerja pertama menerima `UNAVAILABLE` dengan status gRPC `DeadlineExceeded`. Percobaan berikutnya berhasil pada empat permintaan HTTPS.
