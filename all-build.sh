#!/bin/bash -xe
go get github.com/karlmutch/duat/cmd/semver
mkdir -p internal/gen

# Code generation must be done first to allow vendoring and modules to play nice
for dir in cmd/*/ ; do
    base="${dir%%\/}"
    base="${base##*/}"
    if [ "$base" == "cli-experiment" ] ; then
        continue
    fi
    if [ "$base" == "cli-downstream" ] ; then
        continue
    fi
    echo "$base code generator starting"
    $dir/build.sh gen
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "$base code generator failed"
        exit $exit_code
    fi
done

for dir in cmd/*/ ; do
    base="${dir%%\/}"
    base="${base##*/}"
    if [ "$base" == "cli-experiment" ] ; then
        continue
    fi
    if [ "$base" == "cli-downstream" ] ; then
        continue
    fi
    $dir/build.sh
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "$base code compilation failed"
        exit $exit_code
    fi
done
