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
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
dep ensure -no-vendor
[ -e gen/echosrv ] || mkdir -p gen/echosrv
[ -e vendor/github.com/SentientTechnologies/platform-services ] || mkdir -p vendor/github.com/SentientTechnologies/platform-services
[ -e vendor/github.com/SentientTechnologies/platform-services/gen ] || ln -s `pwd`/gen vendor/github.com/SentientTechnologies/platform-services/gen
protoc -Icmd/echosrv -I/usr/include/google --plugin=$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:./gen/echosrv cmd/echosrv/echosrv.proto
mkdir -p cmd/echosrv/bin
go build -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -o cmd/echosrv/bin/echosrv cmd/echosrv/*.go
go build -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -race -o cmd/echosrv/bin/echosrv-race cmd/echosrv/*.go
go test -ldflags "-X github.com/SentientTechnologies/platform-services/version.TestRunMain=Use -X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -coverpkg="." -c -o cmd/echosrv/bin/echosrv-run-coverage cmd/echosrv/*.go
go test -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -coverpkg="." -c -o bin/echosrv-test-coverage cmd/echosrv/*.go
go test -ldflags "-X github.com/SentientTechnologies/platform-services/version.BuildTime=$DATE -X github.com/SentientTechnologies/platform-services/version.GitHash=$HASH" -race -c -o cmd/echosrv/bin/echosrv-test cmd/echosrv/*.go
if ! [ -z ${TRAVIS_TAG+x} ]; then
    if ! [ -z ${GITHUB_TOKEN+x} ]; then
        github-release release --user SentientTechnologies --repo platform-services --tag ${TRAVIS_TAG} --pre-release && \
        github-release upload --user SentientTechnologies --repo platform-services  --tag ${TRAVIS_TAG} --name platform-services --file cmd/echosrv/bin/echosrv
    fi
fi
