# Referensi workforce pada requisition

Requisition kini dapat merujuk baris recruit dari rencana workforce yang masih approved dan memiliki referensi anggaran. Validator memastikan badan hukum manajer, unit, jabatan, kategori pekerja, lokasi jika ditentukan, dan tanggal kebutuhan sesuai. Rencana dikunci dan baris dibaca ulang sebelum pemeriksaan status serta referensi. Penggantian versi menghalangi penggunaan versi lama pada pengajuan dan pembukaan baru.

Tes validator membuktikan penolakan rencana superseded, beda badan hukum dan tanggal di luar periode. Integrasi gRPC ke PostgreSQL membuktikan draf, submit, approval, dan pembukaan dengan baris workforce yang sah. Verifier anggaran requisition masih tiruan dan tetap diperlukan untuk kapasitas headcount, otorisasi anggaran dan mandat. Tes race modul dan go vet lulus.
