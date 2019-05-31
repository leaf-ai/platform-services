#!/bin/bash -x
go get -u github.com/golang/dep/cmd/dep
go install github.com/golang/dep/cmd/dep
go get github.com/karlmutch/duat
go install github.com/karlmutch/duat/cmd/semver
mkdir -p internal/gen

for dir in cmd/*/ ; do
    base="${dir%%\/}"
    base="${base##*/}"
    if [ "$base" == "cli-downstream" ] ; then
        continue
    fi
    $dir/build.sh gen
    if [ $? -ne 0 ]; then
        echo "$base code generator failed"
        exit $?
    fi
    $dir/build.sh
    if [ $? -ne 0 ]; then
        echo "$base code compilation failed"
        exit $?
    fi
done
