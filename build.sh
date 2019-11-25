#!/bin/bash -e
[ -z "$USER" ] && echo "env variable USER must be set" && exit 1;
docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
docker_name=`petname`
docker run --name $docker_name -e GITHUB_TOKEN=$GITHUB_TOKEN -v $GOPATH:/project platform-services 
exit_code=`docker inspect $docker_name --format='{{.State.ExitCode}}'`
if [ $exit_code -ne 0 ]; then
    echo "Error" $exit_code
    exit $exit_code
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

./push.sh
