#!/usr/bin/env bash
set -e
export PGPASSWORD=postgres
psql -d cars -c "GRANT ALL ON SCHEMA public TO program;"
psql -d rentals -c "GRANT ALL ON SCHEMA public TO program;"
psql -d payments -c "GRANT ALL ON SCHEMA public TO program;"
