#!/bin/bash -x
set -e
go get -u github.com/golang/dep/cmd/dep
go install github.com/golang/dep/cmd/dep
go get github.com/karlmutch/duat
go install github.com/karlmutch/duat/cmd/semver
./cmd/experimentsrv/build.sh
if [ $? -ne 0 ]; then
    echo "experimentsrv build failed"
    exit $?
fi
./cmd/downstream/build.sh
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
