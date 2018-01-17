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

# AAA using Auth0

Platform services are secured using the Auth0 service.  Auth0 is a service that provides support for headless machine to machine authentication.  Auth0 is being used initially to provide Bearer tokens for both headless and CLI clients to Sentient platform services.

Auth0 authorizations can be done using a Demo account.  To do this you will need to add clients to the Auth0 dashboard.  

The first client to be added will be the client that accesses the Auth0 service itself in order to then perform per user authentication and token generation. When you being creating a client you will be able to select the "Auth0 Management API" as the API you wish to secure.  You will then be lead through a set of screens to authorize the Auth0 administration capabilities (scopes) for this API.  After saving the initial version of the client you will need to go to the settings page and scroll to the bottom of the page to open the advanced settings section, in this section you should add to the grant types the password grant method.

When adding the API against which clients for the platform services you will use a 'Non Interactive' client in the first page, after being prompted to do the create you will be asked for an API and you should create a New API by using the drop down dialog, "Select an API".  The New API Dialog will ask for a name and an Identifier, Identifiers are used as the 'audience' setting when generating tokens.

You can now use various commands to manipulate the APIs outside of what will exist in the application code, this is a distinct advantage over directly using enterprise tools such as Okta.  Should you wish to use Okta as an Identity provider, or backend, to Auth0 then this can be done however you will need help from our Tech Ops department to do this.  At this time the user and passwords being used for securing APIs can be managed through the Auth0 dashboard including the ability to invite users to become admins.

```
curl --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"RjWuqwm1CM72iQ5G32aUjwIYx6vKTXBa", "client_secret": "MK_jpHrTcthM_HoNETnytYpqgMNS4e7zLMgp1_Wj2aePaPpubjN1UNKKCAfZlD_r", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "openid", "realm": "Username-Password-Authentication" }'

c.f. https://auth0.com/docs/quickstart/backend/golang/02-using#obtaining-an-access-token-for-testing.
```

If you are using the test API you can do something like:

```
cd cmd/experimentsrv
export AUTH0_DOMAIN=sentientai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
LOGXI_FORMAT=happy,maxcol=1024 LOGXI=*=TRC go test -v
```
