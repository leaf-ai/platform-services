# platform-services
A PoC with functioning service using simple Istio Mesh running on K8s

# Installation

<pre><code><b>
go get github.com/SentientTechnologies/platform-services
</b></code></pre>

# Development and Building from source

Clone the repository using the following instructions when this is part of a larger project using multiple services:
<pre><code><b>
mkdir ~/project
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH
mkdir -p src/github.com/SentientTechnologies
cd src/github.com/SentientTechnologies
git clone https://github.com/SentientTechnologies/platform-services
cd platform-services
</b></code></pre>

To boostrap development you will need a copy of Go and the go dependency tools available.  Builds do not need this general however for our purposes we might want to change dependency versions so we should install go and the dep tool.

Go installation instructions can be foubnd at, https://golang.org/doc/install.

Now download any dependencies, once, into our development environment.

<pre><code><b>
go get -u github.com/golang/dep/cmd/dep
dep ensure
</b></code></pre>

Creating a build container to isolate the build into a versioned environment

<pre><code><b>
docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
</b></code></pre>

Running the build using the container

Prior to doing the build a GitHub OAUTH token needs to be defined within your environment.  Use the gibhub admin pages for your account to generate a token, in Travis builds the token is probably already defined by the Travis service.
<pre><code>
<b>docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -e TRAVIS_TAG=$TRAVIS_TAG -v $GOPATH:/project platform-services ; echo "Done" ; docker container prune -f</b>
</code></pre>

A combined build script is provided 'platform-services/build.sh' to allow all stages of the build including producing docker images to be run together.

# Running the AWS Istio example

The experimentsrv component comes with an Istio definition file for deployment into AWS using Kubernetes (k8s) and Istio.

The definition file can be found at cmd/experimentsrv/experimentsrv.yaml.

Using k8s will use both the kops, and the kubectl tools. You should have an AWS account configured prior to starting deployments.

The kops, and kubectl based deployment for AWS clusters is documented and detailed in the AWS workshop guide found at, https://github.com/aws-samples/aws-workshop-for-kubernetes.  Completing the 100 level activities will give you the means to create a basic cluster onto which Istio can be deployed,  Some of the 200 section items are superceeded by Istio.

In order to seed your S3 KOPS_STATE_STORE version controlled bucket with a cluster definition the following command could be used:

<pre><code><b>
export AWS_AVAILABILITY_ZONES="$(aws ec2 describe-availability-zones --query 'AvailabilityZones[].ZoneName' --output text | awk -v OFS="," '$1=$1')"

export S3_BUCKET=kops-platform
export KOPS_STATE_STORE=s3://kops-platform
aws s3 mb $KOPS_STATE_STORE
aws s3api put-bucket-versioning --bucket $S3_BUCKET --versioning-configuration Status=Enabled

export CLUSTER_NAME=test.platform.cluster.k8s.local

kops create cluster --name $CLUSTER_NAME --zones $AWS_AVAILABILITY_ZONES --node-count 1
</b></code></pre>

Optionally use an image from your preferred zone e.g. --image=ami-0def3275.  Also you can modify the AWS machine types, recommended during developer testing using options such as '--master-size=m4.large --node-size=m4.large'.

The Istio install as of 1/1/2018 requires additions to the kops cluster specification. Using the 'kops edit cluster' command change the following:

1. Instead of allowAny on the autorization section use rbac.
<pre><code>
-   authroization:
-     allowAny: {}
+   authrization:
+     rbac: {}
</code></pre>
2. Add into the spec section add the following block as documented at the bottom of, https://github.com/kubernetes/kops/issues/4052 :

<pre><code>
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
</code></pre>

Starting the cluster can now be done using the following command:

<pre><code><b>
kops update cluster $CLUSTER_NAME --yes
</b></code></pre>

The initial cluster spinup will take sometime, use kops commands such as 'kops validate cluster' to determine when the cluster is spun up ready for Istio and the platform services.

You can follow up with the Istio on K8s installation to complete your service mesh cluster found at https://istio.io/docs/setup/kubernetes/quick-start.html. Complete the Installation steps for the Istio tools.  The following commands could be used once the Istio installation is done to the appropriate location as one example:

<pre><code>
<b>export ISTIO_DIR=~/istio-0.4.0</b>
<b>export PATH=$PATH:$ISTIO_DIR/bin</b>
# Begin the istio deploy
<b>kubectl apply -f $ISTIO_DIR/install/kubernetes/istio.yaml</b>
# Wait until the crd times are all valid durations and then continue to apply the 
# initializer, if you saw errors from the initial apply step go back and 
# reapply the instio.yaml state
<b>kubectl get crd</b>
# Now after validating the above continue with the following
<b>kubectl apply -f $ISTIO_DIR/install/kubernetes/istio-initializer.yaml</b>
# Now continue to the optional deployment of horizontal mesh functionality</b>
<b>kubectl apply -f $ISTIO_DIR/install/kubernetes/addons/grafana.yaml</b>
<b>kubectl apply -f $ISTIO_DIR/install/kubernetes/addons/prometheus.yaml</b>
<b>kubectl apply -f $ISTIO_DIR/install/kubernetes/addons/servicegraph.yaml</b>
<b>kubectl apply -f $ISTIO_DIR/install/kubernetes/addons/zipkin.yaml</b>
</code></pre>

The service mesh will be using an Ingress that leverages a version of Envoy called Ambassador.  Ambassador can be injected using the following command:

<pre><code><b>
kubectl apply -f https://getambassador.io/yaml/ambassador/ambassador-rbac.yaml
</b></code></pre>

Ambassador provides a gRPC HTTP/2 ingress which default AWS ELB based load balancers are not able to.  Also provisioned are services for handling authentication and token generation for the users making gRPC requests.

To deploy the platform service passwords and other secrets will be needed to allows access to Aurora and other external resources.  YAML files will be needed to populate secrets into the service mesh, individual services document the secrets they require within their README.md files found on github and provide examples, for example https://github.com/SentientTechnologies/platform-services/cmd/experimentsrv/README.md.  Secrets for these services are currently held within the Kubernetes secrets store and can be populated using the following command:

<pre><code>
# Read https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables
# create secrets, postgres:username, postgres:password
<b>kubectl create -f ./cmd/experimentsrv/secret.yaml</b>
</code></pre>

Platform services use Dockerfiles to encapsulate their build steps which are documented within their respective README.md files.  Building services are single step CLI operations and require only the installation of Docker, and any version of Go 1.7 or later.  Builds will produce containers and will upload these to your current AWS account users ECS docker registry.  Deployments are staged from this registry.  

When creating a cluster an IPv4 address range will have been assigned by AWS and kops for your service cluster.  The details for your address can be found by running the 'kubectl get nodes' command.  Take note of the range and determine the mask as this will be used when deploying the service images into the cluster.  For example:

<pre><code>
<b>kubectl get nodes</b>
NAME                                           STATUS    ROLES     AGE       VERSION
ip-172-20-118-127.us-west-2.compute.internal   Ready     node      17m       v1.8.4
ip-172-20-41-63.us-west-2.compute.internal     Ready     node      17m       v1.8.4
ip-172-20-55-189.us-west-2.compute.internal    Ready     master    18m       v1.8.4
</code></pre>

Which gives a working range of 172.20.0.0/16.

Once secrets are loaded individual services can be deployed from a checked out developer copy of the service repo using the following command :

<pre><code>
<b>cd ~/mesh/src/github.com/SentientTechnologies/platform-services</b>
<b>kubectl apply -f <(istioctl kube-inject -f ./experimentsrv.yaml --includeIPRanges="172.20.0.0/16")</b>
</code></pre>

# Logging and Observability

Currently the service mesh is deployed with Observability tools.  These instruction do not go into Observability at this time.  However we do address logging.

Individual services do offering logging using the systemd facilities and these logs are routed to Kubernetes.  Logs can be obtained from pods and containers. The 'kubectl get services' command can be used to identify the running platform services and the 'kubectl get pod' command can be used to get the health of services.  Once a pod isidentified with a running service instance the logs can be extract using a combination of the pod instance and the service name together, for example:

<code><pre>
<b>kuebctl get services</b>
NAME          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
experiments   ClusterIP   100.68.93.48   <none>        30001/TCP   12m
kubernetes    ClusterIP   100.64.0.1     <none>        443/TCP     1h
<b>kubectl get pod</b>
NAME                             READY     STATUS    RESTARTS   AGE
experiments-v1-bc46b5d68-tltg9   2/2       Running   0          12m
<b>kubectl logs experiments-v1-bc46b5d68-tltg9 experiments</b>
./experimentsrv built at 2018-01-18_15:22:47+0000, against commit id 34e761994b895ac48cd832ac3048854a671256b0
2018-01-18T16:50:18+0000 INF experimentsrv git hash version 34e761994b895ac48cd832ac3048854a671256b0 _: [host experiments-v1-bc46b5d68-tltg9]
2018-01-18T16:50:18+0000 INF experimentsrv database startup / recovery has been performed dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com:5432 name platform _: [host experiments-v1-bc46b5d68-tltg9]
2018-01-18T16:50:18+0000 INF experimentsrv database has 1 connections  dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com:5432 name platform dbConnectionCount 1 _: [host experiments-v1-bc46b5d68-tltg9]
2018-01-18T16:50:33+0000 INF experimentsrv database has 1 connections  dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com:5432 name platform dbConnectionCount 1 _: [host experiments-v1-bc46b5d68-tltg9]
</code></pre>

The container name can also include the istio mesh and kubernetes installed system containers for indepth debugging purposes.

# AAA using Auth0

Platform services are secured using the Auth0 service.  Auth0 is a service that provides support for headless machine to machine authentication.  Auth0 is being used initially to provide Bearer tokens for both headless and CLI clients to Sentient platform services.

Auth0 authorizations can be done using a Demo account.  To do this you will need to add clients to the Auth0 dashboard.  

The first client to be added will be the client that accesses the Auth0 service itself in order to then perform per user authentication and token generation. When you being creating a client you will be able to select the "Auth0 Management API" as the API you wish to secure.  You will then be lead through a set of screens to authorize the Auth0 administration capabilities (scopes) for this API.  After saving the initial version of the client you will need to go to the settings page and scroll to the bottom of the page to open the advanced settings section, in this section you should add to the grant types the password grant method.

When adding the API client definition against which the platform services will interact you will use a 'Non Interactive' client in the first page, after being prompted to do the create you will be asked for an API and you should create a New API by using the drop down dialog, "Select an API".  The New API Dialog will ask for a name and an Identifier, Identifiers are used as the 'audience' setting when generating tokens.

You can now use various commands to manipulate the APIs outside of what will exist in the application code, this is a distinct advantage over directly using enterprise tools such as Okta.  Should you wish to use Okta as an Identity provider, or backend, to Auth0 then this can be done however you will need help from our Tech Ops department to do this and is an expensive option.  At this time the user and passwords being used for securing APIs can be managed through the Auth0 dashboard including the ability to invite users to become admins.

```
curl --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"RjWuqwm1CM72iQ5G32aUjwIYx6vKTXBa", "client_secret": "MK_jpHrTcthM_HoNETnytYpqgMNS4e7zLMgp1_Wj2aePaPpubjN1UNKKCAfZlD_r", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "openid", "realm": "Username-Password-Authentication" }'

c.f. https://auth0.com/docs/quickstart/backend/golang/02-using#obtaining-an-access-token-for-testing.
```

If you are using the test API you can do something like:

<pre><code><b>
cd cmd/experimentsrv
export AUTH0_DOMAIN=sentientai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
LOGXI_FORMAT=happy,maxcol=1024 LOGXI=*=TRC go test -v
</b></code></pre>

# Manually invoking and using services

Services used within the platform require that not only is the link integrity and security is maintained using mTLS but that an authorization block is also supplied to verify the user requesting a service.  The authorization can be supplied when using the gRPC command line tool using the metadata options.  First we retrieve a token using curl and then make a request against the service as follows:

<pre><code><b>
export AUTH="eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IlJ...............qYzRSRUV5TmpFM09UQXdSRFZCTmtSQ056QkNPVEJET"
grpc_cli call localhost:30001 ai.sentient.experiment.Service.Get "id: 'test'" --metadata authorization:"Bearer $AUTH"
</b></code></pre>

The services used within the platfor all support reflection when using gRPC.  To examine calls available for a server you should first identify the endpoint through which the ingress is being routed, for example:

<pre><code>
<b>export CLUSTER_INGRESS=`kubectl get ingress -o wide | tail -1 | awk '{print $3":"$4}'`</b>
<b>grpc_cli ls $CLUSTER_INGRESS -l</b>
filename: grpc_health_v1/health.proto
package: grpc.health.v1;
service Health {
  rpc Check(grpc.health.v1.HealthCheckRequest) returns (grpc.health.v1.HealthCheckResponse) {}
}

filename: grpc_reflection_v1alpha/reflection.proto
package: grpc.reflection.v1alpha;
service ServerReflection {
  rpc ServerReflectionInfo(stream grpc.reflection.v1alpha.ServerReflectionRequest) returns (stream grpc.reflection.v1alpha.ServerReflectionResponse) {}
}

filename: experimentsrv.proto
package: ai.sentient.experiment;
service Service {
  rpc Create(ai.sentient.experiment.CreateRequest) returns (ai.sentient.experiment.CreateResponse) {}
  rpc Get(ai.sentient.experiment.GetRequest) returns (ai.sentient.experiment.GetResponse) {}
}
</code></pre>

To drill further into interfaces and examine the types being used within calls you can perform commands such as:

<pre><code>
<b>grpc_cli type $CLUSTER_INGRESS ai.sentient.experiment.CreateRequest -l</b>
message CreateRequest {
.ai.sentient.experiment.Experiment experiment = 1[json_name = "experiment"];
}
grpc_cli type $CLUSTER_INGRESS ai.sentient.experiment.Experiment -l 
message Experiment {
string uid = 1[json_name = "uid"];
string name = 2[json_name = "name"];
string description = 3[json_name = "description"];
.google.protobuf.Timestamp created = 4[json_name = "created"];
map<uint32, .ai.sentient.experiment.InputLayer> inputLayers = 5[json_name = "inputLayers"];
map<uint32, .ai.sentient.experiment.OutputLayer> outputLayers = 6[json_name = "outputLayers"];
}
<b>grpc_cli type $CLUSTER_INGRESS ai.sentient.experiment.InputLayer -l</b>
message InputLayer {
enum Type {
	Unknown = 0;
	Enumeration = 1;
	Time = 2;
	Raw = 3;
}
string name = 1[json_name = "name"];
.ai.sentient.experiment.InputLayer.Type type = 2[json_name = "type"];
repeated string values = 3[json_name = "values"];
}
</code></pre>


# Shutting down a service, or cluster

<pre><code><b>
kubectl delete service experiments ; kubectl delete ingress ingress-exp ; kubectl delete deployment experiments-v1 ; kubectl delete egressrule aurora-postgres-egress-rule
</b></code></pre>

<pre><code><b>
kops delete cluster $CLUSTER_NAME --yes
</b></code></pre>
