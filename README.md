# platform-services
A PoC with functioning service using simple Istio Mesh running on K8s

# Installation

go get github.com/SentientTechnologies/platform-services

# Development and Building from source

Clone the repository using the following instructions when this is part of a larger project using multiple services:
```
mkdir ~/project
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH
mkdir -p src/github.com/SentientTechnologies
cd src/github.com/SentientTechnologies
git clone https://github.com/SentientTechnologies/platform-services
cd platform-services
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
docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
```

Running the build using the container

Prior to doing the build a GitHub OAUTH token needs to be defined within your environment.  Use the gibhub admin pages for your account to generate a token, in Travis builds the token is probably already defined by the Travis service.
```
docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -e TRAVIS_TAG=$TRAVIS_TAG -v $GOPATH:/project platform-services ; echo "Done" ; docker container prune -f
```

A combined build script is provided 'platform-services/build.sh' to allow all stages of the build including producing docker images to be run together.

# Running the AWS Istio example

The experimentsrv component comes with an Istio definition file for deployment into AWS using Kubernetes (k8s) and Istio.

The definition file can be found at cmd/experimentsrv/experimentsrv.yaml.

Using k8s will use both the kops, and the kubectl tools. You should have an AWS account configured prior to starting deployments.

The kops, and kubectl based deployment for AWS clusters is documented and detailed in the AWS workshop guide found at, https://github.com/aws-samples/aws-workshop-for-kubernetes.  Completing the 100 level activities will give you the means to create a basic cluster onto which Istio can be deployed,  Some of the 200 section items are superceeded by Istio.

The Istio install as of 1/1/2018 requires additions to the kops cluster specification. Using the 'kops edit cluster' command change the following:

1. Instead of allowAny on the autorization section use rbac.
```
-   authroization:
-     allowAny: {}
+   authrization:
+     rbac: {}
```
2. Add into the spec section add the following block as documented at the bottom of, https://github.com/kubernetes/kops/issues/4052 :
```
  kubeAPIServer:
    admissionControl:
    - Initializers
    - NamespaceLifecycle
    - LimitRanger
    - ServiceAccount
    - PersistentVolumeLabel
    - DefaultStorageClass
    - DefaultTolerationSeconds
    - NodeRestriction
    - Priority
    - ResourceQuota
    runtimeConfig:
      admissionregistration.k8s.io/v1alpha1: "true"
```

kops update cluster [your cluster name] --yes

You can follow up with the Istio on K8s installation to complete your service mesh cluster found at https://istio.io/docs/setup/kubernetes/quick-start.html. Complete the Installation steps for the Istio tools.

To deploy the service use the kubectl apply -f section with the service file, cmd/experimentsrv/experimentsrv.yaml.
