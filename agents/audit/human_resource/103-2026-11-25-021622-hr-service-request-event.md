# Audit event permintaan layanan HR

Operasi generik pembuatan event menerima tipe bebas dan keputusan persetujuan. Tes regresi mereproduksi penerimaan event penugasan, resolusi, penutupan, tipe kustom, dan komentar dengan keputusan.

Validator kini membatasi operasi generik ke event `commented` tanpa keputusan. Pemeriksaan aktor terautentikasi, status permintaan, dan idempotensi tetap berlaku. Kebijakan tipe event ini memakai tipe komentar yang sudah dicakup tes.
