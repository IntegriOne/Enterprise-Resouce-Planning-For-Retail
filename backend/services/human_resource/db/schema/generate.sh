#!/usr/bin/env bash
# Menghasilkan satu baseline pra-rilis dari DBML dan invariant PostgreSQL.
set -euo pipefail
schema_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
migration_dir="$schema_dir/../migrations"
temp_sql="$(mktemp)"
trap 'rm -f "$temp_sql"' EXIT
npm exec --yes --package=@dbml/cli@12.0.0 -- dbml2sql "$schema_dir/human_resource_management.dbml" --postgres -o "$temp_sql"
python3 - "$schema_dir" "$temp_sql" "$migration_dir" <<'PY'
from pathlib import Path
import re
import sys
schema_dir, temp_sql, migration_dir = map(Path, sys.argv[1:])
# Timestamp export tidak termasuk skema; hilangkan agar diff deterministik.
body = re.sub(r'^-- Generated at:.*\n', '', temp_sql.read_text(), flags=re.M)
invariants = (schema_dir / 'postgresql-invariants.sql').read_text()
header = '-- Dihasilkan oleh db/schema/generate.sh; jangan edit baseline secara manual.\n'
up = header + body + '\n' + invariants
migration_dir.mkdir(parents=True, exist_ok=True)
(migration_dir / '000001_hrm.up.sql').write_text(up)
# DROP serentak menangani siklus FK internal tanpa CASCADE ke objek asing.
tables = re.findall(r'^CREATE TABLE "([^"]+)"', body, re.M)
enums = re.findall(r'^CREATE TYPE "([^"]+)"', body, re.M)
functions = re.findall(r'^CREATE FUNCTION (\w+)\(', invariants, re.M)
down = header + 'DROP TABLE ' + ',\n  '.join('"' + t + '"' for t in tables) + ';\n'
down += ''.join(f'DROP FUNCTION {name}();\n' for name in functions)
down += ''.join(f'DROP TYPE "{name}";\n' for name in enums)
down += '-- Extension btree_gist dipertahankan karena dapat digunakan objek lain.\n'
(migration_dir / '000001_hrm.down.sql').write_text(down)
print(f'Generated baseline: {len(tables)} tables, {len(enums)} enums')
PY
