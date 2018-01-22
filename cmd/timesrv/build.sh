#!/bin/bash

if ( find /project -maxdepth 0 -empty | read v );
then
  echo "source code must be mounted into the /project directory"
  exit 990
fi

set -e
set -o pipefail

export HASH=`git rev-parse HEAD`
export DATE=`date '+%Y-%m-%d_%H:%M:%S%z'`
export PATH=$PATH:$GOPATH/bin
go get -u -f github.com/golang/dep/cmd/dep
go get -u -f github.com/aktau/github-release
go get -u -f github.com/go-swagger/go-swagger/cmd/swagger
[ -e gen ] || mkdir gen
swagger generate server -q -t gen -f cmd/timesrv/swagger.yaml --exclude-main -A timesrv
# go get -u -f gen/...
dep ensure -no-vendor
[ -e vendor/github.com/SentientTechnologies/platform-services ] || mkdir -p vendor/github.com/SentientTechnologies/platform-services
[ -e vendor/github.com/SentientTechnologies/platform-services/gen ] || ln -s `pwd`/gen vendor/github.com/SentientTechnologies/platform-services/gen
mkdir -p cmd/timesrv/bin
go build -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -o cmd/timesrv/bin/timesrv cmd/timesrv/*.go
go build -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -race -o cmd/timesrv/bin/timesrv-race cmd/timesrv/*.go
if ! [ -z "${TRAVIS_TAG}" ]; then
    if ! [ -z "${GITHUB_TOKEN}" ]; then
        github-release release --user SentientTechnologies --repo platform-services --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user SentientTechnologies --repo platform-services  --tag ${TRAVIS_TAG} --name platform-services --file cmd/timesrv/bin/timesrv
    fi
fi
