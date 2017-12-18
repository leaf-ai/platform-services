#!/bin/bash -x

if ( find /project -maxdepth 0 -empty | read v );
then
  echo "source code must be mounted into the /project directory"
  exit 990
fi

export HASH=`git rev-parse --short HEAD`
export DATE=`date '+%Y-%m-%d_%H:%M:%S%z'`
export PATH=$PATH:$GOPATH/bin
go get -u -f github.com/golang/dep/cmd/dep
go get -u -f github.com/aktau/github-release
dep ensure -no-vendor
mkdir -p cmd/expmanager/bin
go build -ldflags "-X github.com/KarlMutch/MeshTest/version.BuildTime=$DATE -X github.com/KarlMutch/MeshTest/version.GitHash=$HASH" -o cmd/expmanager/bin/expmanager cmd/expmanager/*.go
go build -ldflags "-X github.com/KarlMutch/MeshTest/version.BuildTime=$DATE -X github.com/KarlMutch/MeshTest/version.GitHash=$HASH" -race -o cmd/expmanager/bin/expmanager-race cmd/expmanager/*.go
go test -ldflags "-X github.com/KarlMutch/MeshTest/version.TestRunMain=Use -X github.com/KarlMutch/MeshTest/version.BuildTime=$DATE -X github.com/KarlMutch/MeshTest/version.GitHash=$HASH" -coverpkg="." -c -o cmd/expmanager/bin/expmanager-run-coverage cmd/expmanager/*.go
go test -ldflags "-X github.com/KarlMutch/MeshTest/version.BuildTime=$DATE -X github.com/KarlMutch/MeshTest/version.GitHash=$HASH" -coverpkg="." -c -o cmd/expmanager/bin/expmanager-test-coverage cmd/expmanager/*.go
go test -ldflags "-X github.com/KarlMutch/MeshTest/version.BuildTime=$DATE -X github.com/KarlMutch/MeshTest/version.GitHash=$HASH" -race -c -o cmd/expmanager/bin/expmanager-test cmd/expmanager/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user karlmutch --repo MeshTest --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user karlmutch --repo MeshTest  --tag ${TRAVIS_TAG} --name MeshTest --file cmd/expmanager/bin/expmanager
    fi
fi
