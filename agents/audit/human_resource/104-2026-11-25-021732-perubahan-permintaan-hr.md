# Audit perubahan permintaan HR

Tes integrasi gRPC dengan PostgreSQL disposable mereproduksi perubahan tenggat melalui CRUD sesudah penugasan. Operasi update generik kini hanya berlaku selama status tetap open. Integrasi membuktikan perubahan tenggat saat open, penolakan sesudah penugasan, penolakan event triaged, penerimaan komentar, dan kelanjutan workflow penugasan, resolusi, serta penutupan dengan verifier tiruan.
