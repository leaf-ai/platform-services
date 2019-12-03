#!/bin/bash -e
set -x
go get github.com/karlmutch/duat
go install github.com/karlmutch/duat/cmd/semver
version=`$GOPATH/bin/semver`

set +e
`aws ecr get-login --no-include-email --region us-west-2`
if [ $? -eq 0 ]; then
    true
    set -e
    account=`aws sts get-caller-identity --output text --query Account 2> /dev/null || true`
    if [ $? -eq 0 ]; then
        for dir in cmd/*/ ; do
            base="${dir%%\/}"
            base="${base##*/}"
            if [ "$base" == "cli-experiment" ] ; then
                continue
            fi
            if [ "$base" == "cli-downstream" ] ; then
                continue
            fi
            docker tag $base:$version $account.dkr.ecr.us-west-2.amazonaws.com/platform-services/$base:$version
            docker push $account.dkr.ecr.us-west-2.amazonaws.com/platform-services/$base:$version
        done
    fi
fi

microk8s.status
if [ $? -eq 0 ]; then
    set -e
    microk8s.enable storage registry
    for dir in cmd/*/ ; do
        base="${dir%%\/}"
        base="${base##*/}"
            if [ "$base" == "cli-experiment" ] ; then
                continue
            fi
        if [ "$base" == "cli-downstream" ] ; then
            continue
        fi
        docker tag $base:$version localhost:32000/platform-services/$base:$version
        docker push localhost:32000/platform-services/$base:$version
    done
fi
