#!/bin/sh
if [ -n "${PGPASSWORD}" ]; then
    echo "${PGHOST}:${PGPORT}:${PGDATABASE}:${PGUSER}:${PGPASSWORD}" >> /root/.pgpass
    chmod 0600 /root/.pgpass
    export PGPASSWORDFILE="/root/.pgpass"
fi
./experimentsrv
