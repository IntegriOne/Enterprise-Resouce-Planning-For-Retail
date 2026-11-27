# Tolak persetujuan kompensasi sendiri

Validator persetujuan mengunci worker dan menolak principal IAM milik pekerja penerima. Tes gRPC PostgreSQL membuktikan penolakan worker meskipun ia bukan pembuat atau editor, diikuti persetujuan oleh pihak independen. Race detector dan go vet lulus.
