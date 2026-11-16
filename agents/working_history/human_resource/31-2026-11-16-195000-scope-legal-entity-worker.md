# Scope legal entity pada pembacaan worker

Policy HRM kini memetakan grant `legal_entity` ke `primary_legal_entity_reference_id` khusus query `hr_workers`. Resource Get menggunakan filter scope dan ID dalam satu query; list, finder dan count menyaring sebelum pagination. Mutasi worker serta scope parsial resource lain tetap fail-closed.

Tes unit policy dan service membuktikan dua legal entity, grant global, data luar scope dan snapshot nonaktif. Tes gRPC `bufconn` melalui repository PostgreSQL `_test` membuktikan GetByID, GetAll berhalaman, finder dan count. `go fmt`, `go vet`, race test, build dan Docker healthy lulus. Belum ada grant IAM/E2E Gateway nyata.
