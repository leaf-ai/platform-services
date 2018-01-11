#!/bin/bash -x

if ( find /project -maxdepth 0 -empty | read v );
then
  experiment "source code must be mounted into the /project directory"
  exit 990
fi

export HASH=`git rev-parse HEAD`
export DATE=`date '+%Y-%m-%d_%H:%M:%S%z'`
export PATH=$PATH:$GOPATH/bin
go get -u -f github.com/golang/dep/cmd/dep
go get -u -f github.com/aktau/github-release
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
dep ensure -no-vendor
[ -e gen/experimentsrv ] || mkdir -p gen/experimentsrv
[ -e vendor/github.com/SentientTechnologies/platform-services ] || mkdir -p vendor/github.com/SentientTechnologies/platform-services
[ -e vendor/github.com/SentientTechnologies/platform-services/gen ] || ln -s `pwd`/gen vendor/github.com/SentientTechnologies/platform-services/gen
protoc -Icmd/experimentsrv -I/usr/include/google --plugin=$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:./gen/experimentsrv cmd/experimentsrv/experimentsrv.proto
mkdir -p cmd/experimentsrv/bin
CGO_ENABLED=0 go build -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -o cmd/experimentsrv/bin/experimentsrv cmd/experimentsrv/*.go
go build -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -race -o cmd/experimentsrv/bin/experimentsrv-race cmd/experimentsrv/*.go
CGO_ENABLED=0 go test -ldflags "-X github.com/SentientTechnologies/platform-services/version.TestRunMain=Use -X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -coverpkg="." -c -o cmd/experimentsrv/bin/experimentsrv-run-coverage cmd/experimentsrv/*.go
CGO_ENABLED=0 go test -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -coverpkg="." -c -o bin/experimentsrv-test-coverage cmd/experimentsrv/*.go
go test -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -race -c -o cmd/experimentsrv/bin/experimentsrv-test cmd/experimentsrv/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user SentientTechnologies --repo platform-services --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user SentientTechnologies --repo platform-services  --tag ${TRAVIS_TAG} --name platform-services --file cmd/experimentsrv/bin/experimentsrv
    fi
fi
