#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${APP_ENV:-${ENV:-dev}}"
DB_TYPE=""
HOST=""
PORT=""
USER=""
PASSWORD=""
DATABASE=""
SSL_MODE=""
SQL_DIR=""

usage() {
  cat <<'EOF'
Usage: shell/import_model_sql.sh [options]

Options:
  --env ENV             Load config/<ENV>/app.yaml when explicit DB args are omitted.
  --host HOST           Database host.
  --port PORT           Database port.
  --user USER           Database user.
  --password PASS       Database password.
  --database DB         Target database name.
  --ssl-mode MODE       PostgreSQL sslmode. Defaults to config app.database.ssl_mode or require.
  --sql-dir DIR         Override SQL folder. Defaults to docs/model_sql_pg.
  --help                Show this message.

Examples:
  bash shell/import_model_sql.sh --env prod
  bash shell/import_model_sql.sh --host db.example.com --port 5432 --user postgres --password secret --database postgres
EOF
}

yaml_value() {
  local key="$1"
  local file="config/${ENV_NAME}/app.yaml"
  python3 - "$file" "$key" <<'PY'
import re
import sys
from pathlib import Path

path = Path(sys.argv[1])
key = sys.argv[2]
if not path.exists():
    raise SystemExit(0)

in_database = False
for raw in path.read_text().splitlines():
    if re.match(r'^  database:\s*$', raw):
        in_database = True
        continue
    if in_database and re.match(r'^  [A-Za-z_]+:', raw):
        break
    if in_database:
        match = re.match(rf'^    {re.escape(key)}:\s*(.*)\s*$', raw)
        if match:
            value = match.group(1).strip()
            if (value.startswith('"') and value.endswith('"')) or (value.startswith("'") and value.endswith("'")):
                value = value[1:-1]
            print(value)
            break
PY
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)
      ENV_NAME="$2"
      shift 2
      ;;
    --host)
      HOST="$2"
      shift 2
      ;;
    --port)
      PORT="$2"
      shift 2
      ;;
    --user)
      USER="$2"
      shift 2
      ;;
    --password)
      PASSWORD="$2"
      shift 2
      ;;
    --database|--db)
      DATABASE="$2"
      shift 2
      ;;
    --ssl-mode)
      SSL_MODE="$2"
      shift 2
      ;;
    --sql-dir)
      SQL_DIR="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

DB_TYPE="${DB_TYPE:-$(yaml_value type)}"
DB_TYPE="${DB_TYPE:-pgsql}"
HOST="${HOST:-$(yaml_value host)}"
USER="${USER:-$(yaml_value user)}"
PASSWORD="${PASSWORD:-$(yaml_value password)}"
DATABASE="${DATABASE:-$(yaml_value name)}"
SSL_MODE="${SSL_MODE:-$(yaml_value ssl_mode)}"
SSL_MODE="${SSL_MODE:-require}"

case "$DB_TYPE" in
  pgsql|postgres|postgresql)
    DB_TYPE="pgsql"
    PORT="${PORT:-$(yaml_value port)}"
    PORT="${PORT:-5432}"
    SQL_DIR="${SQL_DIR:-docs/model_sql_pg}"
    ;;
  *)
    echo "Error: unsupported database type '$DB_TYPE'." >&2
    exit 1
    ;;
esac

if [[ -z "$HOST" || -z "$USER" || -z "$DATABASE" ]]; then
  echo "Error: host, user, and database are required." >&2
  usage >&2
  exit 1
fi

if [[ ! -d "$SQL_DIR" ]]; then
  echo "Error: SQL directory '$SQL_DIR' does not exist." >&2
  exit 1
fi

SQL_FILES=()
while IFS= read -r sql_path; do
  SQL_FILES+=("$sql_path")
done < <(find "$SQL_DIR" -maxdepth 1 -type f -name '*.sql' | sort)

if [[ ${#SQL_FILES[@]} -eq 0 ]]; then
  echo "No SQL files found in $SQL_DIR" >&2
  exit 1
fi

echo "Importing ${#SQL_FILES[@]} SQL files from $SQL_DIR into ${DB_TYPE}://${USER}@${HOST}:${PORT}/${DATABASE}"
for sql_file in "${SQL_FILES[@]}"; do
  echo "--> Applying $sql_file"
  if command -v psql >/dev/null 2>&1; then
    PGPASSWORD="$PASSWORD" psql "host=$HOST port=$PORT user=$USER dbname=$DATABASE sslmode=$SSL_MODE connect_timeout=10" -v ON_ERROR_STOP=1 -f "$sql_file"
  else
    docker run --rm --network host -v "$(pwd):/workspace" -e PGPASSWORD="$PASSWORD" postgres:16 sh -lc "psql 'host=$HOST port=$PORT user=$USER dbname=$DATABASE sslmode=$SSL_MODE connect_timeout=10' -v ON_ERROR_STOP=1 -f '/workspace/$sql_file'"
  fi
done

echo "All SQL files imported successfully."
