# Transport mTLS Gateway ke HRM

Regresi sebelum wiring mencatat 68 paket HRM lulus. Transport HRM di Gateway memakai TLS 1.3, CA dan sertifikat klien eksplisit, HMAC caller, hash sesi employee terverifikasi dan deadline. Tes race source connections dan config serta vet lulus. Registrasi startup, handshake lintas kontainer, schema, dan resolver masih terbuka.
