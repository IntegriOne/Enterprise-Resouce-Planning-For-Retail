# LinkWorkerPrincipal melalui gRPC dan PostgreSQL

Tes dengan race detector memakai verifier fixture. Bukti yang berbeda ditolak dengan FailedPrecondition. Bukti yang mengikat principal, worker, person party, dan pelaku diterima. Permintaan ulang ditolak dan satu event lifecycle tercatat. Verifier produksi belum dipasang pada startup HRM.

Tes melalui registry tanpa verifier membuktikan permintaan valid dari pelaku yang diizinkan menerima Unavailable dan referensi principal pekerja tetap kosong. Pemeriksaan otorisasi dan kelayakan pekerja mendahului pemeriksaan verifier.

Pengaitan principal yang sudah terikat pada pekerja aktif lain mengembalikan AlreadyExists meski fixture verifier menerima bukti. Tautan pada pekerja kedua tetap kosong, tanpa event lifecycle atau outbox tambahan.
