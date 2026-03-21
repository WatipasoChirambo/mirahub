#!/bin/sh
# wait-for-db.sh
set -e

host="${1:-localhost}"  # default to localhost if not provided
shift
cmd="$@"

echo "Waiting for database at $host..."

until pg_isready -h "$host" -p "${DB_PORT:-5432}" -U "$DB_USER"; do
  echo "Waiting for database at $host..."
  sleep 2
done

echo "Database is ready. Starting app..."
exec $cmd