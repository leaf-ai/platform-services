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
go get -u -f github.com/go-swagger/go-swagger/cmd/swagger
[ -e gen ] || mkdir gen
swagger generate server -t gen -f cmd/timesrv/swagger.yaml --exclude-main -A timesrv
#go get -u -f gen/...
dep ensure -no-vendor
[ -e vendor/github.com/karlmutch/MeshTest ] || mkdir -p vendor/github.com/karlmutch/MeshTest
[ -e vendor/github.com/karlmutch/MeshTest/gen ] || ln -s `pwd`/gen vendor/github.com/karlmutch/MeshTest/gen
mkdir -p cmd/timesrv/bin
go build -ldflags "-X github.com/karlmutch/MeshTest/version.BuildTime=$DATE -X github.com/karlmutch/MeshTest/version.GitHash=$HASH" -o cmd/timesrv/bin/timesrv cmd/timesrv/*.go
go build -ldflags "-X github.com/karlmutch/MeshTest/version.BuildTime=$DATE -X github.com/karlmutch/MeshTest/version.GitHash=$HASH" -race -o cmd/timesrv/bin/timesrv-race cmd/timesrv/*.go
go test -ldflags "-X github.com/karlmutch/MeshTest/version.TestRunMain=Use -X github.com/karlmutch/MeshTest/version.BuildTime=$DATE -X github.com/karlmutch/MeshTest/version.GitHash=$HASH" -coverpkg="." -c -o cmd/timesrv/bin/timesrv-run-coverage cmd/timesrv/*.go
go test -ldflags "-X github.com/karlmutch/MeshTest/version.BuildTime=$DATE -X github.com/karlmutch/MeshTest/version.GitHash=$HASH" -coverpkg="." -c -o bin/timesrv-test-coverage cmd/timesrv/*.go
go test -ldflags "-X github.com/karlmutch/MeshTest/version.BuildTime=$DATE -X github.com/karlmutch/MeshTest/version.GitHash=$HASH" -race -c -o cmd/timesrv/bin/timesrv-test cmd/timesrv/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user karlmutch --repo MeshTest --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user karlmutch --repo MeshTest  --tag ${TRAVIS_TAG} --name MeshTest --file cmd/timesrv/bin/timesrv
    fi
fi
