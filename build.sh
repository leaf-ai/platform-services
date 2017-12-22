#!/bin/bash -e
docker build -t meshtest:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -e TRAVIS_TAG=$TRAVIS_TAG -v $GOPATH:/project meshtest ; echo "Done" ; docker container prune -f
cd cmd/echosrv
docker build -t echosrv .
cd ../timesrv
docker build -t timesrv .
cd ../expmanager
docker build -t expmanager .
cd ../..
