# Validator workforce planning awal

Composition root HRM kini membuat Integrity dari repository aktif dan gagal startup bila dependency validator inti hilang. Validator `job_families` dan `jobs` memeriksa kode unik, mengunci job family aktif, serta menolak penghapusan/deaktivasi saat referensi inti masih ada. Sebanyak 76 resource berkode mutasi lain tetap fail-closed; enam resource riwayat/hasil hanya baca. Tes unit/race/vet serta build dan health Docker lulus. Belum ada E2E positif, scope parsial, relay outbox, atau validator lengkap seluruh resource mutabel.
