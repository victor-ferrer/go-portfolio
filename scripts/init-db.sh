#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Database is already created by POSTGRES_DB env var
    -- This script can be used for additional initialization if needed
    \echo 'Database go-portfolio initialized successfully'
EOSQL
