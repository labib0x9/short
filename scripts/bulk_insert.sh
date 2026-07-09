#!/usr/bin/env bash

set -e

SERVICE="${1:-postgres}"
docker compose exec -T "$SERVICE" sh -c 'psql -U "$PG_USER" -d "$PG_NAME"' < scripts/bulk_insert.sql