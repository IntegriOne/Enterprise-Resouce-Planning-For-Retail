# Benturan jadwal wawancara dalam satu lamaran

Tes gRPC dan PostgreSQL sementara mereproduksi dua putaran dengan jadwal beririsan pada satu lamaran. Validator kini menolak irisan dengan wawancara planned atau scheduled saat lamaran masih terkunci. Jadwal yang bersebelahan diperbolehkan oleh perbandingan batas waktu eksklusif. Pengecekan menjaga pesan AlreadyExists pada duplikasi putaran yang sama. Integrasi Recruitment dengan race detector, vet modul, dan diff check lulus. Cakupan benturan terbatas pada satu lamaran.
