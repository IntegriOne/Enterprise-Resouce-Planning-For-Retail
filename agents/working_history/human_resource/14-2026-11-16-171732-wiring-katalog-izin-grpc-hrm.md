# Wiring katalog izin gRPC HRM

Peta permission eksplisit untuk 84 resource dan tiga command khusus dihubungkan ke policy runtime. Aksi `list` memakai kode `read`; enam resource riwayat/hasil hanya baca. Delete, posting cuti, approval payroll, finalisasi talent dan akses sensitif tetap memerlukan MFA walau metadata grant IAM salah. Tes katalog/policy, generator proto, vet, tes race, serta build dan health Docker lulus. Scope non-global, validator mutasi, relay outbox, grant IAM dan E2E positif belum selesai; operasi tersebut tetap fail-closed.
