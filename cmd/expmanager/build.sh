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
mkdir -p cmd/expmanager/bin
go build -ldflags "-X version.BuildTime=$DATE -X version.GitHash=$HASH" -o cmd/expmanager/bin/expmanager cmd/expmanager/*.go
go build -ldflags "-X version.BuildTime=$DATE -X version.GitHash=$HASH" -race -o cmd/expmanager/bin/expmanager-race cmd/expmanager/*.go
go test -ldflags "-X command-line-arguments.TestRunMain=Use -X command-line-arguments.BuildTime=$DATE -X command-line-arguments.GitHash=$HASH" -coverpkg="." -c -o cmd/expmanager/bin/expmanager-run-coverage cmd/expmanager/*.go
go test -ldflags "-X command-line-arguments.BuildTime=$DATE -X command-line-arguments.GitHash=$HASH" -coverpkg="." -c -o cmd/expmanager/bin/expmanager-test-coverage cmd/expmanager/*.go
go test -ldflags "-X command-line-arguments.BuildTime=$DATE -X command-line-arguments.GitHash=$HASH" -race -c -o cmd/expmanager/bin/expmanager-test cmd/expmanager/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user karlmutch --repo MeshTest --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user karlmutch --repo MeshTest  --tag ${TRAVIS_TAG} --name MeshTest --file cmd/expmanager/bin/expmanager
    fi
fi