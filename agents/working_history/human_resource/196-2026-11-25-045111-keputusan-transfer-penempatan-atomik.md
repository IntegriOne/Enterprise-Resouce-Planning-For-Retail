# Keputusan transfer penempatan atomik

RPC privat Personnel menerima persetujuan dan penolakan dengan izin transfer dan step-up. Persetujuan independen menutup penempatan sumber, memvalidasi kapasitas dan referensi Foundation, memasang pengganti, serta menyimpan lifecycle dan outbox dalam satu transaksi. Integrasi gRPC–PostgreSQL dengan race detector membuktikan rollback, penerapan, retry, dan penolakan. Go test dan vet modul lulus. Handoff Gateway mencakup 987 RPC.
