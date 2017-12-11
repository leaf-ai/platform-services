# MeshTest
A PoC with functioning service using simple Istio Mesh running on K8s

# Installation

go get github.com/karlmutch/MeshTest

Or from source when this is part of a larger project using multiple services:
```
mkdir ~/project
export GOPATH=`pwd`
mkdir -p src/github.com/KarlMutch
cd src/github.com/KarlMutch
git clone github.com/KarlMutch/MeshTest
cd MeshTest
# Multistage builds for Docker
```
