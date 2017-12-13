#!/bin/bash -x

if ( find /project -maxdepth 0 -empty | read v );
then
  echo "source code must be mounted into the /project directory"
  exit 990
fi

export HASH=`git rev-parse HEAD`
export DATE=`date '+%Y-%m-%d_%H:%M:%S%z'`
export PATH=$PATH:$GOPATH/bin
go get -u -f github.com/golang/dep/cmd/dep
go get -u -f github.com/aktau/github-release
dep ensure -no-vendor
mkdir -p bin
go build -ldflags "-X main.buildTime=$DATE -X main.gitHash=$HASH" -o bin/timesrv cmd/timesrv/*.go
go build -ldflags "-X main.buildTime=$DATE -X main.gitHash=$HASH" -race -o bin/timesrv-race cmd/timesrv/*.go
go test -ldflags "-X command-line-arguments.TestRunMain=Use -X command-line-arguments.buildTime=$DATE -X command-line-arguments.gitHash=$HASH" -coverpkg="." -c -o bin/timesrv-run-coverage cmd/timesrv/*.go
go test -ldflags "-X command-line-arguments.buildTime=$DATE -X command-line-arguments.gitHash=$HASH" -coverpkg="." -c -o bin/timesrv-test-coverage cmd/timesrv/*.go
go test -ldflags "-X command-line-arguments.buildTime=$DATE -X command-line-arguments.gitHash=$HASH" -race -c -o bin/timesrv-test cmd/timesrv/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user karlmutch --repo MeshTest --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user karlmutch --repo MeshTest  --tag ${TRAVIS_TAG} --name MeshTest --file bin/timesrv
    fi
fi
