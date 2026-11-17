#!/usr/bin/env bash
# Memvalidasi DBML dan mengekspor skema kini tanpa menimpa migrasi historis.
set -euo pipefail
schema_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
migration_dir="$schema_dir/../migrations"
if (( $# > 1 )); then
  printf '%s\n' 'Usage: generate.sh [output.sql]' >&2
  exit 2
fi
temp_sql="$(mktemp)"
trap 'rm -f "$temp_sql"' EXIT
npm exec --yes --package=@dbml/cli@12.0.0 -- dbml2sql "$schema_dir/human_resource_management.dbml" --postgres -o "$temp_sql"
python3 - "$schema_dir" "$temp_sql" "$migration_dir" "${1:-}" <<'PY'
from pathlib import Path
import re
import sys
schema_dir, temp_sql, migration_dir = map(Path, sys.argv[1:4])
output = sys.argv[4]
# Timestamp export tidak termasuk skema; hilangkan agar diff deterministik.
body = re.sub(r'^-- Generated at:.*\n', '', temp_sql.read_text(), flags=re.M)
invariants = (schema_dir / 'postgresql-invariants.sql').read_text()
tables = re.findall(r'^CREATE TABLE "([^"]+)"', body, re.M)
enums = re.findall(r'^CREATE TYPE "([^"]+)"', body, re.M)
if output:
    target = Path(output).resolve()
    if target.is_relative_to(migration_dir.resolve()):
        raise SystemExit('Historical HRM migrations are immutable; choose another output path')
    target.write_text('-- Skema HRM saat ini; migrasi historis tetap immutable.\n' + body + '\n' + invariants)
    print(f'Exported current schema: {len(tables)} tables, {len(enums)} enums')
else:
    print(f'Validated current DBML: {len(tables)} tables, {len(enums)} enums')
PY
