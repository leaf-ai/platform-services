# platform-services
A PoC with functioning service using simple Istio Mesh running on K8s

Version : <repo-version>0.0.0-master-1ei1Hn</repo-version>

# Installation

<pre><code><b>go get github.com/SentientTechnologies/platform-services
</b></code></pre>

# Development and Building from source

Clone the repository using the following instructions when this is part of a larger project using multiple services:
<pre><code><b>mkdir ~/project
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

<pre><code><b>go get -u github.com/golang/dep/cmd/dep
dep ensure
</b></code></pre>

Creating a build container to isolate the build into a versioned environment

<pre><code><b>docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
</b></code></pre>

Running the build using the container

Prior to doing the build a GitHub OAUTH token needs to be defined within your environment.  Use the gibhub admin pages for your account to generate a token, in Travis builds the token is probably already defined by the Travis service.
<pre><code><b>docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -e TRAVIS_TAG=$TRAVIS_TAG -v $GOPATH:/project platform-services ; echo "Done" ; docker container prune -f</b>
</code></pre>

A combined build script is provided 'platform-services/build.sh' to allow all stages of the build including producing docker images to be run together.

# Deploying the Istio Service platform on AWS with Kubernetes

The k8s instructions in this section are for unmanaged solutions.  They are included as a baseline for AWS prior to the wide availability of EKS.  Once EKS is in wide distribution and managed offerings for k8s is available from the big three cloud vendors and k8s receeds into the cloud platform then these instructions will become redundant and the cloud vendors tooling will take over this function.

## Kubernetes (unmanaged)

The experimentsrv component comes with an Istio definition file for deployment into AWS using Kubernetes (k8s) and Istio.

The deployment definition file can be found at cmd/experimentsrv/experimentsrv.yaml.

Using k8s will use both the kops, and the kubectl tools. You should have an AWS account configured prior to starting deployments, and your environment variables for using the AWS cli should also be done.

### Verify Docker Version

Docker is preinstalled.  You can verify the version by running the following:
<pre><code><b>docker --version</b>
Docker version 17.12.0-ce, build c97c6d6
</code></pre>
You should have a similar or newer version.
## Install Kubectl CLI

Install the kubectl CLI can be done using any 1.9.x version.

<pre><code><b> curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.9.3/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/</b>
</code></pre>

Add kubectl autocompletion to your current shell:

<pre><code><b>source <(kubectl completion bash)</b>
</code></pre>

You can verify that kubectl is installed by executing the following command:

<pre><code><b>kubectl version --client</b>
Client Version: version.Info{Major:"1", Minor:"9", GitVersion:"v1.9.2", GitCommit:"5fa2db2bd46ac79e5e00a4e6ed24191080aa463b", GitTreeState:"clean", BuildDate:"2018-01-18T10:09:24Z", GoVersion:"go1.9.2", Compiler:"gc", Platform:"linux/amd64"}
</code></pre>

### Install kops

At the time this guide was updated kops 1.9.0 Alpha 1 was released, if you are reading this guide in April of 2018 or later look for the release version of kops 1.9 or later.  kops for the AWS use case at the alpha is a very restricted use case for our purposes and works in a stable fashion.  If you are using azure or GCP then options such as acs-engine, and skaffold are natively supported by the cloud vendors and written in Go so are readily usable and can be easily customized and maintained and so these are recommended for those cases.

<pre><code><b>curl -LO https://github.com/kubernetes/kops/releases/download/1.9.0-alpha.1/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops
</b></code></pre>

In order to seed your S3 KOPS_STATE_STORE version controlled bucket with a cluster definition the following command could be used:

<pre><code><b>export AWS_AVAILABILITY_ZONES="$(aws ec2 describe-availability-zones --query 'AvailabilityZones[].ZoneName' --output text | awk -v OFS="," '$1=$1')"

export S3_BUCKET=kops-platform
export KOPS_STATE_STORE=s3://kops-platform
aws s3 mb $KOPS_STATE_STORE
aws s3api put-bucket-versioning --bucket $S3_BUCKET --versioning-configuration Status=Enabled

export CLUSTER_NAME=test.platform.cluster.k8s.local

kops create cluster --name $CLUSTER_NAME --zones $AWS_AVAILABILITY_ZONES --node-count 1
</b></code></pre>

Optionally use an image from your preferred zone e.g. --image=ami-0def3275.  Also you can modify the AWS machine types, recommended during developer testing using options such as '--master-size=m4.large --node-size=m4.large'.


Starting the cluster can now be done using the following command:

<pre><code><b>kops update cluster $CLUSTER_NAME --yes</b>
I0309 13:48:49.798777    6195 apply_cluster.go:442] Gossip DNS: skipping DNS validation
I0309 13:48:49.961602    6195 executor.go:91] Tasks: 0 done / 81 total; 30 can run
I0309 13:48:50.383671    6195 vfs_castore.go:715] Issuing new certificate: "ca"
I0309 13:48:50.478788    6195 vfs_castore.go:715] Issuing new certificate: "apiserver-aggregator-ca"
I0309 13:48:50.599605    6195 executor.go:91] Tasks: 30 done / 81 total; 26 can run
I0309 13:48:51.013957    6195 vfs_castore.go:715] Issuing new certificate: "kube-controller-manager"
I0309 13:48:51.087447    6195 vfs_castore.go:715] Issuing new certificate: "kube-proxy"
I0309 13:48:51.092714    6195 vfs_castore.go:715] Issuing new certificate: "kubelet"
I0309 13:48:51.118145    6195 vfs_castore.go:715] Issuing new certificate: "apiserver-aggregator"
I0309 13:48:51.133527    6195 vfs_castore.go:715] Issuing new certificate: "kube-scheduler"
I0309 13:48:51.157876    6195 vfs_castore.go:715] Issuing new certificate: "kops"
I0309 13:48:51.167195    6195 vfs_castore.go:715] Issuing new certificate: "apiserver-proxy-client"
I0309 13:48:51.172542    6195 vfs_castore.go:715] Issuing new certificate: "kubecfg"
I0309 13:48:51.179730    6195 vfs_castore.go:715] Issuing new certificate: "kubelet-api"
I0309 13:48:51.431304    6195 executor.go:91] Tasks: 56 done / 81 total; 21 can run
I0309 13:48:51.568136    6195 launchconfiguration.go:334] waiting for IAM instance profile "nodes.test.platform.cluster.k8s.local" to be ready
I0309 13:48:51.576067    6195 launchconfiguration.go:334] waiting for IAM instance profile "masters.test.platform.cluster.k8s.local" to be ready
I0309 13:49:01.973887    6195 executor.go:91] Tasks: 77 done / 81 total; 3 can run
I0309 13:49:02.489343    6195 vfs_castore.go:715] Issuing new certificate: "master"
I0309 13:49:02.775403    6195 executor.go:91] Tasks: 80 done / 81 total; 1 can run
I0309 13:49:03.074583    6195 executor.go:91] Tasks: 81 done / 81 total; 0 can run
I0309 13:49:03.168822    6195 update_cluster.go:279] Exporting kubecfg for cluster
kops has set your kubectl context to test.platform.cluster.k8s.local

Cluster is starting.  It should be ready in a few minutes.

Suggestions:
 * validate cluster: kops validate cluster
 * list nodes: kubectl get nodes --show-labels
 * ssh to the master: ssh -i ~/.ssh/id_rsa admin@api.test.platform.cluster.k8s.local
 * the admin user is specific to Debian. If not using Debian please use the appropriate user based on your OS.
 * read about installing addons at: https://github.com/kubernetes/kops/blob/master/docs/addons.md.

</code></pre>

The initial cluster spinup will take sometime, use kops commands such as 'kops validate cluster' to determine when the cluster is spun up ready for Istio and the platform services.

## Istio

Istio affords a control layer on top of the k8s data plane.  These instructions have been updated for istio 0.6.0

Instructions for deploying Istio are the vanilla instructions that can be found at, https://istio.io/docs/setup/kubernetes/quick-start.html#installation-steps.  We recommend using the mTLS installation for the k8s cluster deployment, for example

<pre><code><b>cd ~
curl -LO https://github.com/istio/istio/releases/download/0./.0/istio-0.4.0-linux.tar.gz
tar xzf istion-0.4.0-linux.tar.gz
export ISTIO_DIR=`pwd`/istio-0.4.0
export PATH=$ISTIO_DIR/bin:$PATH
cd -
kubectl apply -f $ISTIO_DIR/install/kubernetes/istio-auth.yaml
sleep 3
# Wait for a few seconds for the CRDs get loaded then reapply the states to fill in any missing resources
kubectl apply -f $ISTIO_DIR/install/kubernetes/istio-auth.yaml
</b></code></pre>

## Deploying a straw-man service into the Istio control plane

To deploy the platform service passwords and other secrets will be needed to allows access to Aurora and other external resources.  YAML files will be needed to populate secrets into the service mesh, individual services document the secrets they require within their README.md files found on github and provide examples, for example https://github.com/SentientTechnologies/platform-services/cmd/experimentsrv/README.md.  Secrets for these services are currently held within the Kubernetes secrets store and can be populated using the following command:

<pre><code># Read https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables
# create secrets, postgres:username, postgres:password
<b>kubectl create -f ./cmd/experimentsrv/secret.yaml</b>
</code></pre>

Platform services use Dockerfiles to encapsulate their build steps which are documented within their respective README.md files.  Building services are single step CLI operations and require only the installation of Docker, and any version of Go 1.7 or later.  Builds will produce containers and will upload these to your current AWS account users ECS docker registry.  Deployments are staged from this registry.  

<pre><code><b>kubectl get nodes</b>
NAME                                           STATUS    ROLES     AGE       VERSION
ip-172-20-118-127.us-west-2.compute.internal   Ready     node      17m       v1.9.3
ip-172-20-41-63.us-west-2.compute.internal     Ready     node      17m       v1.9.3
ip-172-20-55-189.us-west-2.compute.internal    Ready     master    18m       v1.9.3
</code></pre>

Once secrets are loaded individual services can be deployed from a checked out developer copy of the service repo using a command like the following :

<pre><code><b>cd ~/mesh/src/github.com/SentientTechnologies/platform-services</b>
<b>kubectl apply -f <(istioctl kube-inject -f [application-deployment-yaml])</b>
</code></pre>

When version controlled containers are being used with ECS or another docker registry the bump-ver can be used to extract a git cloned repository that has the version string embeeded inside the README.md or another file of your choice, and then use this with your application deployment yaml specification, as follows:

<pre><code><b>cd ~/mesh/src/github.com/SentientTechnologies/platform-services</b>
<b>kubectl apply -f <(istioctl kube-inject -f <(bump-ver -t ./experimentsrv.yaml -f ./README.md inject))</b>
</b></code></pre>

The bump-ver tool can be installed using `go install github.com/karlmutch/bump-ver`.  It uses the semver repos to extract and manipulate sem vers for the build tagging and docker.

Once the application is deployed you can discover the ingress points within the kubernetes cluster by using the following:
<pre><code><b>export CLUSTER_INGRESS=`kubectl get ingress -o wide | tail -1 | awk '{print $3":"$4}'`
</b></code></pre>

More information aboutr deploying and using the experimentsrv server can be found at, https://github.com/SentientTechnologies/platform-services/blob/master/cmd/experimentsrv/README.md.

# Logging and Observability

Currently the service mesh is deployed with Observability tools.  These instruction do not go into Observability at this time.  However we do address logging.

Individual services do offering logging using the systemd facilities and these logs are routed to Kubernetes.  Logs can be obtained from pods and containers. The 'kubectl get services' command can be used to identify the running platform services and the 'kubectl get pod' command can be used to get the health of services.  Once a pod isidentified with a running service instance the logs can be extract using a combination of the pod instance and the service name together, for example:

<pre><code><b>kuebctl get services</b>
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

<pre><code><b>curl --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"RjWuqwm1CM72iQ5G32aUjwIYx6vKTXBa", "client_secret": "MK_jpHrTcthM_HoNETnytYpqgMNS4e7zLMgp1_Wj2aePaPpubjN1UNKKCAfZlD_r", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "openid", "realm": "Username-Password-Authentication" }'
</b>
c.f. https://auth0.com/docs/quickstart/backend/golang/02-using#obtaining-an-access-token-for-testing.
</code></pre>

If you are using the test API you can do something like:

<pre><code><b>cd cmd/experimentsrv
export AUTH0_DOMAIN=sentientai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://sentientai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.sentient.ai/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
LOGXI_FORMAT=happy,maxcol=1024 LOGXI=*=TRC go test -v -ip-port ":30001"
</b></code></pre>

# Manually invoking and using services

Services used within the platform require that not only is the link integrity and security is maintained using mTLS but that an authorization block is also supplied to verify the user requesting a service.  The authorization can be supplied when using the gRPC command line tool using the metadata options.  First we retrieve a token using curl and then make a request against the service as follows:

<pre><code><b>grpc_cli call localhost:30001 ai.sentient.experiment.Service.Get "id: 'test'" --metadata authorization:"Bearer $AUTH0_TOKEN"
</b></code></pre>

The services used within the platfor all support reflection when using gRPC.  To examine calls available for a server you should first identify the endpoint through which the ingress is being routed, for example:

<pre><code><b>export CLUSTER_INGRESS=`kubectl get ingress -o wide | tail -1 | awk '{print $3":"$4}'`</b>
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

<pre><code><b>grpc_cli type $CLUSTER_INGRESS ai.sentient.experiment.CreateRequest -l</b>
message CreateRequest {
.ai.sentient.experiment.Experiment experiment = 1[json_name = "experiment"];
}
<b>grpc_cli type $CLUSTER_INGRESS ai.sentient.experiment.Experiment -l</b>
message Experiment {
string uid = 1[json_name = "uid"];
string name = 2[json_name = "name"];
string description = 3[json_name = "description"];
.google.protobuf.Timestamp created = 4[json_name = "created"];
map&lt;uint32, .ai.sentient.experiment.InputLayer&gt; inputLayers = 5[json_name = "inputLayers"];
map&lt;uint32, .ai.sentient.experiment.OutputLayer&gt; outputLayers = 6[json_name = "outputLayers"];
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

<pre><code><b>kubectl delete service experiments ; kubectl delete ingress ingress-exp ; kubectl delete deployment experiments-v1 ; kubectl delete secrets postgres
</b></code></pre>

<pre><code><b>kops delete cluster $CLUSTER_NAME --yes
</b></code></pre>
