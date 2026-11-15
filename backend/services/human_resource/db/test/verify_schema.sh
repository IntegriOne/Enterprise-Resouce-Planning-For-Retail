#!/usr/bin/env bash
# Verifikasi pada database sekali pakai milik task, bukan database aplikasi.
set -euo pipefail
test_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git -C "$test_dir" rev-parse --show-toplevel)"
cd "$repo_root"
test_db="hrm_schema_$(date +%s)_${RANDOM}_test"
created=false
cleanup() {
  if [[ "$created" == true ]]; then
    docker compose exec -T postgres sh -c 'dropdb -U "$POSTGRES_USER" "$1"' sh "$test_db"
  fi
}
trap cleanup EXIT
docker compose exec -T postgres sh -c 'createdb -U "$POSTGRES_USER" "$1"' sh "$test_db"
created=true
run_sql() {
  docker compose exec -T postgres sh -c 'psql -X -q -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1"' sh "$test_db"
}
run_sql < "$test_dir/../migrations/000001_hrm.up.sql"
run_sql < "$test_dir/schema_integrity.sql"
run_sql < "$test_dir/../migrations/000001_hrm.down.sql"
printf '%s\n' "DO \$\$ BEGIN IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'hr_%') THEN RAISE EXCEPTION 'Rollback left HR tables'; END IF; END \$\$;" | run_sql
run_sql < "$test_dir/../migrations/000001_hrm.up.sql"
printf '%s\n' 'Baseline up, integrity cases, down, and up passed.'
