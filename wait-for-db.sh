#!/bin/sh
# wait-for-db.sh: wait until Postgres is ready

set -e

host="$1"
shift
cmd="$@"

until pg_isready -h "$host" -p 5432 -U "mirahub" >/dev/null 2>&1; do
  echo "Waiting for database at $host..."
  sleep 2
done

exec $cmd