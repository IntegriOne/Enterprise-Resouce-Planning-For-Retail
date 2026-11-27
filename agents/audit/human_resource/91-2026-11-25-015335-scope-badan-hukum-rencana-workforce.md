# Scope badan hukum rencana workforce

Policy HRM kini menerima grant baca legal_entity untuk rencana Workforce Planning. Policy memetakan grant ke filter legal_entity_reference_id pada query Get, list, finder dan total sebelum pagination. Pembacaan satu objek dengan atribut badan hukum di luar scope ditolak. Pembatasan step-up, pemeriksaan freshness, dan grant global tetap mengikuti policy pusat.

Tes policy memeriksa filter, objek dalam dan luar badan hukum, penolakan grant submit parsial, serta larangan scope parsial pada baris rencana yang membutuhkan relasi melalui parent. Tes memakai snapshot tiruan. Verifikasi gRPC dengan snapshot IAM nyata masih terbuka.
