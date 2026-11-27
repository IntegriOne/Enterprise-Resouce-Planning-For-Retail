# Ajukan payroll untuk review

RPC privat mengizinkan penginisiasi memasukkan run calculated ke review setelah rekonsiliasi hasil dan periode terbuka. Tes gRPC–PostgreSQL dengan race detector membuktikan penolakan hasil tidak lengkap, pengajuan oleh principal lain, dan pengajuan berulang. Pengajuan dan approval oleh penyetuju independen menghasilkan outbox terpisah. Go test dan go vet lulus.
