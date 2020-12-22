#!/bin/bash -x

if ( find /project -maxdepth 0 -empty | read v );
then
  echo "source code must be mounted into the /project directory"
  exit 990
fi

set -e
set -o pipefail

export HASH=`git rev-parse --short HEAD`
export PATH=$PATH:$GOPATH/bin
go get -u github.com/itchio/gothub

if [ "$1" == "gen" ]; then
    exit 0
fi

go mod vendor
mkdir -p cmd/restpoc/bin
go build -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -o cmd/restpoc/bin/restpoc cmd/restpoc/*.go
go build -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -race -o cmd/restpoc/bin/restpoc-race cmd/restpoc/*.go
go test -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.TestRunMain=Use -X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -coverpkg="." -c -o cmd/restpoc/bin/restpoc-run-coverage cmd/restpoc/*.go
go test -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -coverpkg="." -c -o cmd/restpoc/bin/restpoc-test-coverage cmd/restpoc/*.go
go test -asmflags -trimpath -ldflags "-X github.com/leaf-ai/platform-services/internal/version.GitHash=$HASH" -race -c -o cmd/restpoc/bin/restpoc-test cmd/restpoc/*.go
if ! [ -z "${TRAVIS_TAG}" ]; then
    if ! [ -z "${GITHUB_TOKEN}" ]; then
        gothub-release release --user leaf-ai --repo platform-services --tag ${TRAVIS_TAG} --pre-release || true
        gothub-release upload --user leaf-ai --repo platform-services  --tag ${TRAVIS_TAG} --name restpoc --file cmd/restpoc/bin/restpoc
    fi
fi
