# Regresi transfer setelah terminasi

Tes gRPC dengan PostgreSQL sementara dan race detector membuktikan transfer tertunda tidak dapat disetujui setelah terminasi dan masih dapat ditolak secara eksplisit. Handoff Gateway mencatat perilaku tersebut. [Laporan pengujian](../../e2e/human_resource/41-personnel-mutation-terminate-worker.md)
