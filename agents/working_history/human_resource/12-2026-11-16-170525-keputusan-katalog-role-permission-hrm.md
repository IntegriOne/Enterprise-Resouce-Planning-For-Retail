# Keputusan katalog akses HRM

Atas mandat owner, katalog Enterprise Foundation dipakai sebagai pola penamaan dan batas akses HRM. Diputuskan format `human_resource.<resource>.<action>`: 84 izin baca, 78 izin tulis, dan tiga command khusus (165 kode). Enam resource riwayat, ledger, atau hasil kalkulasi hanya memperoleh izin baca. Template role, scope `global`/`legal_entity`/`organization_unit`/`self`, operasi wajib step-up, dan pemisahan tugas payroll, cuti, talent, serta promosi ditetapkan dalam katalog HRM.

Keputusan katalog belum membuat permission atau grant IAM. Policy produksi masih tanpa mapping dan validator integrity belum lengkap, sehingga operasi bisnis tetap fail-closed sampai implementasi dan E2E scoped selesai.
