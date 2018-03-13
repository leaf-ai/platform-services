#!/bin/bash -e
[ -z "$USER" ] && echo "env variable USER must be set" && exit 1;
docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -v $GOPATH:/project platform-services 
if [ $? -ne 0 ]; then
    echo ""
    exit $?
fi

echo "Done" ; docker container prune -f

cd cmd/experimentsrv
go install github.com/karlmutch/bump-ver/cmd/bump-ver
version=`$GOPATH/bin/bump-ver -git ../.. -f ../../README.md extract`

docker build -t experimentsrv:$version .

`aws ecr get-login --no-include-email --region us-west-2`
if [ $? -eq 0 ]; then
    account=`aws sts get-caller-identity --output text --query Account`
    if [ $? -eq 0 ]; then
        docker tag experimentsrv:$version $account.dkr.ecr.us-west-2.amazonaws.com/experimentsrv:$version
        docker push $account.dkr.ecr.us-west-2.amazonaws.com/experimentsrv:$version
    fi
fi
cd ../echosrv
docker build -t echosrv .
cd ../timesrv
docker build -t timesrv .
cd ../restpoc
docker build -t restpoc .
cd ../..
