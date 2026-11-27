# Bootstrap superadmin dan grant EF

Commit `08fdec6` terverifikasi pada `origin/main`. Runtime IAM menyelaraskan bootstrap superadmin, registrasi katalog dan verifikasi grant oleh EF. Lint, vet, race suite, build binary dan image lulus.

Stack HTTPS disposable dari main lulus untuk login/MFA superadmin, grant lewat GraphQL, provisioning employee dan replay, penautan IAM–HRM, aktivasi, terminasi otomatis, pencabutan sesi, dan clearance mTLS. Finance/DRM tidak tersedia pada stack ini.
