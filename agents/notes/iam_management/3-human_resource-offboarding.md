# Verifikasi keputusan pencabutan akses offboarding

HRM memerlukan verifikasi privat dari IAM atas keputusan offboarding yang mengikat worker melalui hubungan principal, rencana, scope organisasi, dan tanggal efektif. Respons perlu membedakan keputusan efektif dari pending serta menyertakan versi otoritatif atau waktu validasi. HRM memakai identitas workload, audience, deadline, dan actor terverifikasi. Jika IAM tidak dapat membuktikan keputusan efektif, penutupan rencana tidak tersimpan. HRM tidak mencabut sesi atau grant melalui command penutupan.
