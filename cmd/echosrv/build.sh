#!/bin/bash -x

if ( find /project -maxdepth 0 -empty | read v );
then
  echo "source code must be mounted into the /project directory"
  exit 990
fi

set -e
set -o pipefail

export HASH=`git rev-parse HEAD`
export PATH=$PATH:$GOPATH/bin
go get -u -f github.com/itchio/gothub
go get -u google.golang.org/grpc
go get -u google.golang.org/protobuf/cmd/protoc-gen-go
[ -e internal/gen/echosrv ] || mkdir -p internal/gen/echosrv
#[ -e vendor/github.com/leaf-ai/platform-services/internal ] || mkdir -p vendor/github.com/leaf-ai/platform-services/internal
#[ -e vendor/github.com/leaf-ai/platform-services/internal/gen ] || ln -s `pwd`/internal/gen vendor/github.com/leaf-ai/platform-services/internal/gen
protoc -Icmd/echosrv -I/usr/include/google --plugin=$GOPATH/bin/protoc-gen-go --go_out=./internal/gen/echosrv --go_opt=paths=source_relative --plugin=$GOPATH/bin/protoc-gen-go-grpc --go-grpc_out=./internal/gen/echosrv cmd/echosrv/echosrv.proto --go-grpc_opt=paths=source_relative

if [ "$1" == "gen" ]; then
    exit 0
fi

go mod vendor
mkdir -p cmd/echosrv/bin
go build -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -o cmd/echosrv/bin/echosrv cmd/echosrv/*.go
go build -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -race -o cmd/echosrv/bin/echosrv-race cmd/echosrv/*.go
if ! [ -z "${TRAVIS_TAG}" ]; then
    if ! [ -z "${GITHUB_TOKEN}" ]; then
        gothub-release release --user leaf-ai --repo platform-services --tag ${TRAVIS_TAG} --pre-release || true
        gothub-release upload --user leaf-ai --repo platform-services  --tag ${TRAVIS_TAG} --name echosrv --file cmd/echosrv/bin/echosrv
    fi
fi
