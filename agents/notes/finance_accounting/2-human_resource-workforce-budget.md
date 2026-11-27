# Bukti anggaran untuk persetujuan rencana workforce

HRM memerlukan operasi privat Finance untuk memverifikasi otorisasi anggaran rencana workforce. Permintaan mengikat ID dan versi rencana, skenario, badan hukum, periode, actor, serta snapshot baris kebutuhan yang berisi unit, jabatan, kategori pekerja, kode tindakan, FTE, rentang tanggal kebutuhan, serta lokasi, biaya dan mata uang bila tersedia. Respons sukses harus mengembalikan ID otorisasi Finance yang stabil dan terikat pada snapshot yang sama.

Panggilan membawa identitas workload dan audience, scope badan hukum, deadline, serta bukti versi atau masa berlaku keputusan. Ketidakcocokan snapshot atau anggaran harus menghasilkan konflik bisnis. Gangguan dependency mengembalikan Unavailable dan mempertahankan status review. Verifikasi bersifat baca dan harus aman diulang.
