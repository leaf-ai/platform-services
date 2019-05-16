#!/bin/bash -x
set -e
go get -u github.com/golang/dep/cmd/dep
go install github.com/golang/dep/cmd/dep
go get github.com/karlmutch/duat
go install github.com/karlmutch/duat/cmd/semver
mkdir -p internal/gen
./cmd/experimentsrv/build.sh gen
if [ $? -ne 0 ]; then
    echo "experimentsrv build failed"
    exit $?
fi
./cmd/downstream/build.sh gen
if [ $? -ne 0 ]; then
    echo "experimentsrv build failed"
    exit $?
fi
./cmd/echosrv/build.sh gen
if [ $? -ne 0 ]; then
    echo "echosrv build failed"
    exit $?
fi
./cmd/timesrv/build.sh gen
if [ $? -ne 0 ]; then
    echo "timesrv build failed"
    exit $?
fi
./cmd/restpoc/build.sh gen
if [ $? -ne 0 ]; then
    echo "restpoc build failed"
    exit $?
fi
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
