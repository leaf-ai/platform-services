#!/bin/bash -e

if ( find /project -maxdepth 0 -empty | read v );
then
  echo "source code must be mounted into the /project directory"
  exit -1
fi

set -e
set -o pipefail

export HASH=`git rev-parse HEAD`
export DATE=`date '+%Y-%m-%d_%H:%M:%S%z'`
export PATH=$PATH:$GOPATH/bin
go get -u -f github.com/golang/dep/cmd/dep
go get -u -f github.com/aktau/github-release
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
go install github.com/karlmutch/bump-ver/cmd/bump-ver
go install github.com/golang/dep/cmd/dep
dep ensure -no-vendor
export SEMVER=`bump-ver -f README.md -git ../.. extract`
TAG_PARTS=$(echo $SEMVER | sed "s/-/\n-/g" | sed "s/\./\n\./g" | sed "s/+/\n+/g")
PATCH=""
for part in $TAG_PARTS
do
    start=`echo "$part" | cut -c1-1`
    if [ "$start" == "+" ]; then
        break
    fi
    if [ "$start" == "-" ]; then
        PATCH+=$part
    fi
done
flags='-X github.com/SentientTechnologies/platform-services/version.BuildTime="$DATE" -X github.com/SentientTechnologies/platform-services/version.GitHash="$HASH" -X github.com/SentientTechnologies/platform-services/version.SemVer="$SEMVER"'
flags="$(eval echo $flags)"
[ -e gen/downstream ] || mkdir -p gen/downstream
[ -e vendor/github.com/SentientTechnologies/platform-services ] || mkdir -p vendor/github.com/SentientTechnologies/platform-services
[ -e vendor/github.com/SentientTechnologies/platform-services/gen ] || ln -s `pwd`/gen vendor/github.com/SentientTechnologies/platform-services/gen
protoc -Icmd/downstream -I/usr/include/google --plugin=$GOPATH/bin/protoc-gen-go --go_out=plugins=grpc:./gen/downstream cmd/downstream/downstream.proto
mkdir -p cmd/downstream/bin
CGO_ENABLED=0 go build -ldflags "$flags" -o cmd/downstream/bin/downstream cmd/downstream/*.go
go build -ldflags "$flags" -race -o cmd/downstream/bin/downstream-race cmd/downstream/*.go
CGO_ENABLED=0 go test -ldflags "$flags" -coverpkg="." -c -o cmd/downstream/bin/downstream-run-coverage cmd/downstream/*.go
CGO_ENABLED=0 go test -ldflags "$flags" -coverpkg="." -c -o bin/downstream-test-coverage cmd/downstream/*.go
go test -ldflags "$flags" -race -c -o cmd/downstream/bin/downstream-test cmd/downstream/*.go
if [ -z "$PATCH" ]; then
    if ! [ -z "${SEMVER}" ]; then
        if ! [ -z "${GITHUB_TOKEN}" ]; then
            github-release release --user SentientTechnologies --repo platform-services --tag ${SEMVER} --pre-release && \
            github-release upload --user SentientTechnologies --repo platform-services  --tag ${SEMVER} --name platform-services --file cmd/downstream/bin/downstream
        fi
    fi
fi
