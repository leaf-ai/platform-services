#!/bin/bash -x
set -e
./cmd/experimentsrv/build.sh
if [ $? -ne 0 ]; then
    echo "experimentsrv build failed"
    exit $?
fi
./cmd/echosrv/build.sh 
if [ $? -ne 0 ]; then
    echo "echosrv build failed"
    exit $?
fi
./cmd/timesrv/build.sh
if [ $? -ne 0 ]; then
    echo "timesrv build failed"
    exit $?
fi
./cmd/restpoc/build.sh
if [ $? -ne 0 ]; then
    echo "restpoc build failed"
    exit $?
fi
