#!/bin/bash -xe
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
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "$base code generator failed"
        exit $exit_code
    fi
    $dir/build.sh
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "$base code compilation failed"
        exit $exit_code
    fi
done
