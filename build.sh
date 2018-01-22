#!/bin/bash -e
docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -e TRAVIS_TAG=$TRAVIS_TAG -v $GOPATH:/project platform-services ; echo "Done" ; docker container prune -f
if [ $? -ne 0 ]; then
    echo ""
    exit $?
fi
cd cmd/experimentsrv
docker build -t experimentsrv .
`aws ecr get-login --no-include-email --region us-west-2`
if [ $? -eq 0 ]; then
    account=`aws sts get-caller-identity --output text --query Account`
    if [ $? -eq 0 ]; then
        docker tag experimentsrv:latest $account.dkr.ecr.us-west-2.amazonaws.com/experimentsrv:latest
        docker push $account.dkr.ecr.us-west-2.amazonaws.com/experimentsrv:latest
    fi
fi
cd ../echosrv
docker build -t echosrv .
cd ../timesrv
docker build -t timesrv .
cd ../restpoc
docker build -t restpoc .
cd ../..
