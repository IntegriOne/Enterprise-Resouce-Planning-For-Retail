# Periksa handoff Gateway dan izin

Alat rekonsiliasi menemukan 18 RPC CRUD yang tetap ada di proto tetapi tidak memiliki grant pada resource baca-saja. Handoff kini menandai operasi tersebut nonaktif dan memuat 669 kandidat Query serta 323 kandidat Mutation dari 1.010 RPC. Pemeriksaan kode izin yang sengaja diubah pada salinan dokumen menghasilkan kegagalan sesuai harapan.
