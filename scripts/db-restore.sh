#!/usr/bin/env bash
set -euo pipefail

# Named configuration (override via environment variables where appropriate)
readonly REQUIRED_PG_MAJOR="16"
readonly DEFAULT_PROD_HOST_PATTERN='\.neon\.tech$'
PROD_HOST_PATTERN="${PROD_HOST_PATTERN:-$DEFAULT_PROD_HOST_PATTERN}"

usage() {
  cat <<'EOF'
Usage:
  scripts/db-restore.sh [--i-know-this-is-prod] <backup.dump.age> <age-identity-file> <target-database-url>

Arguments:
  backup.dump.age       Path to the age-encrypted pg_dump custom archive
  age-identity-file     Path to the offline age private key file (e.g. popcorn-backup.key)
  target-database-url   Target PostgreSQL connection URI (postgresql://... or postgres://...)

Options:
  --i-know-this-is-prod Allow restoring to a host matching PROD_HOST_PATTERN (default: *.neon.tech)
  -h, --help            Show this help message
EOF
}

extract_db_host() {
  local dsn="$1"
  local without_scheme authority hostport host

  if [[ "$dsn" == *"://"* ]]; then
    without_scheme="${dsn#*://}"
    authority="${without_scheme%%/*}"
    authority="${authority%%\?*}"
    hostport="${authority##*@}"

    if [[ "$hostport" == \[* ]]; then
      host="${hostport#\[}"
      host="${host%%\]*}"
    else
      host="${hostport%%:*}"
    fi
  else
    # Support libpq keyword=value DSNs such as "host=example.neon.tech dbname=neondb"
    host="$(printf '%s' "$dsn" | sed -nE 's/.*(^|[[:space:]])host=([^[:space:]]+).*/\2/p')"
  fi

  printf '%s' "$host" | tr '[:upper:]' '[:lower:]'
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: required command '$cmd' was not found in PATH" >&2
    exit 1
  fi
}

main() {
  local allow_prod="false"
  local positional=()

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --i-know-this-is-prod)
        allow_prod="true"
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      --)
        shift
        positional+=("$@")
        break
        ;;
      -*)
        echo "Error: unknown option '$1'" >&2
        usage >&2
        exit 2
        ;;
      *)
        positional+=("$1")
        shift
        ;;
    esac
  done

  if [[ ${#positional[@]} -ne 3 ]]; then
    echo "Error: expected 3 positional arguments, got ${#positional[@]}" >&2
    usage >&2
    exit 2
  fi

  local backup_file="${positional[0]}"
  local identity_file="${positional[1]}"
  local target_db_url="${positional[2]}"

  require_cmd age
  require_cmd pg_restore
  require_cmd psql

  local pg_restore_major
  pg_restore_major="$(pg_restore --version | sed -nE 's/^pg_restore \(PostgreSQL\) ([0-9]+).*/\1/p')"
  if [[ -z "$pg_restore_major" ]] || (( pg_restore_major < REQUIRED_PG_MAJOR )); then
    echo "Error: pg_restore major version >= ${REQUIRED_PG_MAJOR} is required (found '${pg_restore_major:-unknown}')." >&2
    echo "On macOS with Homebrew, prepend postgresql@16 to PATH or run inside a postgres:16 container." >&2
    exit 1
  fi

  if [[ ! -f "$backup_file" || ! -r "$backup_file" ]]; then
    echo "Error: backup file '$backup_file' does not exist or is not readable" >&2
    exit 1
  fi

  if [[ ! -f "$identity_file" || ! -r "$identity_file" ]]; then
    echo "Error: age identity file '$identity_file' does not exist or is not readable" >&2
    exit 1
  fi

  if [[ -z "$target_db_url" ]]; then
    echo "Error: target DATABASE_URL cannot be empty" >&2
    exit 1
  fi

  local target_host
  target_host="$(extract_db_host "$target_db_url")"
  if [[ -z "$target_host" ]]; then
    echo "Error: could not determine target host from target DATABASE_URL; aborting for safety" >&2
    exit 1
  fi

  if [[ "$target_host" =~ $PROD_HOST_PATTERN && "$allow_prod" != "true" ]]; then
    echo "Refusing to restore to '$target_host' (matches production host pattern '$PROD_HOST_PATTERN')." >&2
    echo "Double-check that you are targeting a non-production Neon branch or intentional recovery target, then re-run with --i-know-this-is-prod." >&2
    exit 1
  fi

  echo "Restoring '${backup_file}' into target host '${target_host}'..."
  age -d -i "$identity_file" "$backup_file" \
    | pg_restore --clean --if-exists --no-owner -d "$target_db_url"

  echo "Restore complete. Running sanity-check row counts..."
  psql "$target_db_url" -v ON_ERROR_STOP=1 -c \
    "SELECT 'users' AS table_name, COUNT(*) AS total_rows, COUNT(*) FILTER (WHERE deleted_at IS NULL) AS active_rows FROM users
     UNION ALL
     SELECT 'reviews' AS table_name, COUNT(*) AS total_rows, COUNT(*) FILTER (WHERE deleted_at IS NULL) AS active_rows FROM reviews;"
}

main "$@"
