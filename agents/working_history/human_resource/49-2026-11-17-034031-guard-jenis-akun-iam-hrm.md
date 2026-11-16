# Guard jenis akun IAM pada HRM

Menambahkan pencocokan `user_kind` Gateway, token sesi dan snapshot IAM dalam policy HRM; hanya `employee` diterima untuk actor manusia bersesi. Tes policy diperbarui untuk mismatch, nilai kosong dan customer konsisten. Fmt/vet/test-race/build modul HRM lulus. Jalur IAM/Gateway nyata dan actor workload belum diuji atau diaktifkan.
