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
mkdir -p cmd/timesrv/bin
go build -ldflags "-X version.BuildTime=$DATE -X version.GitHash=$HASH" -o cmd/timesrv/bin/timesrv cmd/timesrv/*.go
go build -ldflags "-X version.BuildTime=$DATE -X version.GitHash=$HASH" -race -o cmd/timesrv/bin/timesrv-race cmd/timesrv/*.go
go test -ldflags "-X command-line-arguments.TestRunMain=Use -X command-line-arguments.BuildTime=$DATE -X command-line-arguments.GitHash=$HASH" -coverpkg="." -c -o cmd/timesrv/bin/timesrv-run-coverage cmd/timesrv/*.go
go test -ldflags "-X command-line-arguments.BuildTime=$DATE -X command-line-arguments.GitHash=$HASH" -coverpkg="." -c -o bin/timesrv-test-coverage cmd/timesrv/*.go
go test -ldflags "-X command-line-arguments.BuildTime=$DATE -X command-line-arguments.GitHash=$HASH" -race -c -o cmd/timesrv/bin/timesrv-test cmd/timesrv/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user karlmutch --repo MeshTest --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user karlmutch --repo MeshTest  --tag ${TRAVIS_TAG} --name MeshTest --file cmd/timesrv/bin/timesrv
    fi
fi
