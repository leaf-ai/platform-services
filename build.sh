#!/bin/bash -e
[ -z "$USER" ] && echo "env variable USER must be set" && exit 1;
docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -v $GOPATH:/project platform-services 
if [ $? -ne 0 ]; then
    echo ""
    exit $?
fi

echo "Done" ; docker container prune -f

go get github.com/karlmutch/duat
go install github.com/karlmutch/duat/cmd/semver
version=`$GOPATH/bin/semver`

for dir in cmd/*/ ; do
    base="${dir%%\/}"
    base="${base##*/}"
    if [ "$base" == "cli-downstream" ] ; then
        continue
    fi
    cd $dir
    docker build -t $base:$version .
    cd -
done

set +e
`aws ecr get-login --no-include-email --region us-west-2 1>/dev/null 2>/dev/null`
if [ $? -eq 0 ]; then
    true
    set -e
    account=`aws sts get-caller-identity --output text --query Account 2> /dev/null || true`
    if [ $? -eq 0 ]; then
        for dir in cmd/*/ ; do
            base="${dir%%\/}"
            base="${base##*/}"
            if [ "$base" == "cli-downstream" ] ; then
                continue
            fi
            docker tag $base:$version $account.dkr.ecr.us-west-2.amazonaws.com/platform-services/$base:$version
            docker push $account.dkr.ecr.us-west-2.amazonaws.com/platform-services/$base:$version
        done
    fi
fi

set -e
for dir in cmd/*/ ; do
    base="${dir%%\/}"
    base="${base##*/}"
    if [ "$base" == "cli-downstream" ] ; then
        continue
    fi
    docker tag $base:$version localhost:32000/platform-services/$base:$version
    docker push localhost:32000/platform-services/$base:$version
done
