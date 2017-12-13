# MeshTest
A PoC with functioning service using simple Istio Mesh running on K8s

# Installation

go get github.com/karlmutch/MeshTest

# Development and Building from source

Clone the repository using the following instructions when this is part of a larger project using multiple services:
```
mkdir ~/project
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH
mkdir -p src/github.com/KarlMutch
cd src/github.com/KarlMutch
git clone https://github.com/KarlMutch/MeshTest
cd MeshTest
```

To boostrap development you will need a copy of Go and the go dependency tools available.  Builds do not need this general however for our purposes we might want to change dependency versions so we should install go and the dep tool.

Go installation instructions can be foubnd at, https://golang.org/doc/install.

Now download any dependencies, once, into our development environment.

```
go get -u github.com/golang/dep/cmd/dep
dep ensure
```

Creating a build container to isolate the build into a versioned environment

```
docker build -t meshtest:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
```

Running the build using the container

Prior to doing the build a GitHub OAUTH token needs to be defined within your environment.  Use the gibhub admin pages for your account to generate a token, in Travis builds the token is probably already defined by the Travis service.
```
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -e TRAVIS_TAG=$TRAVIS_TAG -v $GOPATH:/project meshtest ; echo "Done" ; docker container prune -f
```
