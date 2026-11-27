# Pembatalan rencana workforce

CancelWorkforcePlan menutup draf atau rencana review oleh pembuat manusia yang memiliki izin step-up. Validator hanya mengizinkan perubahan status dan menolak rencana yang memiliki bukti approval atau anggaran. Pembatalan administratif mempertahankan seluruh baris dan tidak memerlukan pembacaan data master Foundation. Actor, waktu dan status tercatat di outbox atomik.

Tes gRPC PostgreSQL membuktikan pembatalan review, penolakan pengulangan dan penghapusan baris, serta tepat satu event pembatalan. Policy memakai stub. Approval dengan bukti Finance dan supersede masih terbuka.
