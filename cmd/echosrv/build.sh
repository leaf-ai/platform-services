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
go get -u google.golang.org/grpc
go install ./vendor/github.com/golang/protobuf/protoc-gen-go
dep ensure -no-vendor
[ -e internal/gen/echosrv ] || mkdir -p internal/gen/echosrv
#[ -e vendor/github.com/leaf-ai/platform-services/internal ] || mkdir -p vendor/github.com/leaf-ai/platform-services/internal
#[ -e vendor/github.com/leaf-ai/platform-services/internal/gen ] || ln -s `pwd`/internal/gen vendor/github.com/leaf-ai/platform-services/internal/gen
protoc -Icmd/echosrv -I/usr/include/google --plugin=$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:./internal/gen/echosrv cmd/echosrv/echosrv.proto
if [ "$1" == "gen" ]; then
    exit 0
fi
mkdir -p cmd/echosrv/bin
go build -ldflags "-X github.com/leaf-ai/platform-services/internal/version.BuildTime=$DATE -X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -o cmd/echosrv/bin/echosrv cmd/echosrv/*.go
go build -ldflags "-X github.com/leaf-ai/platform-services/internal/version.BuildTime=$DATE -X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -race -o cmd/echosrv/bin/echosrv-race cmd/echosrv/*.go
if ! [ -z "${TRAVIS_TAG}" ]; then
    if ! [ -z "${GITHUB_TOKEN}" ]; then
        github-release release --user leaf-ai --repo platform-services --tag ${TRAVIS_TAG} --pre-release || true
        github-release upload --user leaf-ai --repo platform-services  --tag ${TRAVIS_TAG} --name echosrv --file cmd/echosrv/bin/echosrv
    fi
fi
