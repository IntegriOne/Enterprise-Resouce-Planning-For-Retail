# Ikat slip ke penerbitan Finance

Snapshot payroll menyimpan ID penerbitan Finance bersama referensi versi DRM. Replay dengan ID lain ditolak meski verifier menerima permintaan. PostgreSQL menjaga referensi penerbitan immutable setelah dokumen tertaut. Run lama dengan referensi dokumen tanpa ID penerbitan tetap dapat dibaca melalui verifier. Regresi PostgreSQL dengan race detector, verifikasi skema lengkap, dan go vet lulus.
