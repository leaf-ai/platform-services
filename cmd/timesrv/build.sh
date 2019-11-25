#!/bin/bash -x

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
[ -e internal/gen ] || mkdir internal/gen
swagger generate server -q -t internal/gen -f cmd/timesrv/swagger.yaml --exclude-main -A timesrv
# go get -u -f gen/...
dep ensure -no-vendor
#[ -e vendor/github.com/leaf-ai/platform-services/internal ] || mkdir -p vendor/github.com/leaf-ai/platform-services/internal
#[ -e vendor/github.com/leaf-ai/platform-services/internal/gen ] || ln -s `pwd`/internal/gen vendor/github.com/leaf-ai/platform-services/internal/gen
if [ "$1" == "gen" ]; then
    exit 0
fi
mkdir -p cmd/timesrv/bin
go build -ldflags "-X github.com/leaf-ai/platform-services/internal/version.BuildTime=$DATE -X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -o cmd/timesrv/bin/timesrv cmd/timesrv/*.go
go build -ldflags "-X github.com/leaf-ai/platform-services/internal/version.BuildTime=$DATE -X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -race -o cmd/timesrv/bin/timesrv-race cmd/timesrv/*.go
if ! [ -z "${TRAVIS_TAG}" ]; then
    if ! [ -z "${GITHUB_TOKEN}" ]; then
        github-release release --user leaf-ai --repo platform-services --tag ${TRAVIS_TAG} --pre-release || true
        github-release upload --user leaf-ai --repo platform-services  --tag ${TRAVIS_TAG} --name timesrv --file cmd/timesrv/bin/timesrv
    fi
fi
