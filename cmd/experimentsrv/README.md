# experimentsrv

The experiment server is used to persist experiment details and to record changes to the state of experiments.  Items included within an experiment include layer definitions and meta-data items.

The experiment server offers a gRPC API that can be accessed using a machine-to-machine or human-to-machine (HCI) interface.  The HCI interface can be interacted with using the grpc_cli tool provided with the gRPC toolkit  More information about grpc_cli can be found at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

# Experiment Database

The experiment server makes use of a Postgres DB.  The installation process is specific to AWS Aurora.  To begin installation you will need to create a Postgres Aurora RDS instance.  Use values of your choosing for the DB name and user/password combinations.

Parameters that impact deployment of your Aurora instance include, RDS Endpoint, DB Name, user name, and the password.

The command to install the postgres schema into your DB instance will appear similar to the following:

<pre><code><b>PGUSER=pl PGHOST=dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com PGDATABASE=platform psql -f sql/platform.sql -d postgres
</b></code></pre>

## Installation

Before starting you should install several build and deployment tools that will be useful for managing the service configuration file.

<pre><code><b>wget -O $GOPATH/bin/stencil https://github.com/karlmutch/duat/releases/download/0.4.0/stencil
chmod +x $GOPATH/bin/stencil
</b></code></pre>

### Secrets

You should now create or edit the secrets.yaml file ready for deployment with the user name and the password.  The secrets can be injected into your kubernetes secrets DB using the kubectl apply -f [secrets_file_name] command.

To update the secrets file information that needs to be stored should be be encoded as Base 64 and then the result text added to the file entries as appropriate. For example:

<pre><code><b>base64 <(echo -n "dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com")</b>
ZGV2LXBsYXRmb3JtLmNsdXN0ZXItY2ZmMnVodGQyanpoLnVzLXdlc3QtMi5yZHMuYW1hem9uYXdzLmNvbQ==
# Edit the secret.yaml file and replace the host with your modified name from the RDS instance being used
<b>cat secret.yml</b>
apiVersion: v1
kind: Secret
metadata:
  name: postgres
type: Opaque
data:
  username: xxx=
  password: xxxxxxxxxxx=
  host: ZGV2LXBsYXRmb3JtLmNsdXN0ZXItY2ZmMnVodGQyanpoLnVzLXdlc3QtMi5yZHMuYW1hem9uYXdzLmNvbQ==
  port: NTQzMg==
  database: cGxhdGZvcm0=
<b>kubectl apply -f secret.yaml</b>
secret "postgres" created
</code></pre>
Once the host name is inserted into the secrets file the external IP that will be used by the service needs to be determined.  In this present example using ping will reveal the host with the external IP (AWS) that can used within the application deployment file.
<code><pre><b>ping dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com</b>
PING ec2-52-42-136-165.us-west-2.compute.amazonaws.com (172.31.25.240) 56(84) bytes of data.
^C
--- ec2-52-42-136-165.us-west-2.compute.amazonaws.com ping statistics ---
3 packets transmitted, 0 received, 100% packet loss, time 2016ms
</code></pre>
 The resulting change to the egress rule would appear as follows within the experiment.yaml file:

<pre><code>apiVersion: config.istio.io/v1alpha2
kind: EgressRule
metadata:
  name: rds-egress
spec:
  destination:
    service: 52.42.136.165/27
  ports:
    - port: 5432
      protocol: tcp
</code></pre>

### Deployment

The experiment service is deployed using Istio into a Kubernetes (k8s) cluster.  The k8s cluster installation instructions can be found within the README.md file at the top of this github repository.  

When using AWS the local workstation should first associate your local docker instance against your AWS ECR account. To do this run the following command using the AWS CLI.  You should change your named AWS_PROFILE to match that set in your ~/.aws/credentials file.

<pre><code><b>export AWS_PROFILE=platform
export AWS_REGION=us-west-2
`aws ecr get-login --no-include-email`
</b></code></pre>

To deploy the experiment service three commands will be used stencil (a SDLC aware templating tool), istioctl (a service mesh administration tool), and kubectl (a cluster orchestration tool):

When version controlled containers are being used with ECS or another docker registry the semver, and stencil tools can be used to extract a git cloned repository that has the version string embeeded inside the README.md or another file of your choice, and then use this with your application deployment yaml specification, as follows:

<pre><code><b>cd ~/project/src/github.com/leaf-ai/platform-services/cmd/experimentsrv</b>
<b>kubectl apply -f <(istioctl kube-inject --includeIPRanges="172.20.0.0/16"  -f <(stencil < experimentsrv.yaml))
</b></code></pre>

This technique can be used to upgrade software versions etc and performing rolling upgrades.

## Service Authentication

Service authentication is explained within the top level README.md file for the github.com/leaf-ai/platform-services repository.  All calls into the experiment service must contain metadata for the autorization brearer token and have the all:experiments claim in order to be accepted.

<pre><code><b>export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
</b></code></pre>

# Using the service

In order to use the service the ingress should be determined using the following command.  Production systems will be configured using AWS Route 53 or similar.

<pre><code><b>export CLUSTER_INGRESS=`kubectl get ingress -o wide | tail -1 | awk '{print $3":"$4}'`
</b></code></pre>

## grpc_cli

The grpc_cli tool can be used to interact with the server for creating and getting experiments.  Other tools do exist as curl like environments for interacting with gRPC servers including the nodejs based tool found at, https://github.com/njpatel/grpcc.  For our purposes we use the less powerful but more common grpc_cli tool that comes with the gRPC project.  Documentation for the grpc_cli tool can be found at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

Should you be considering writing or using a service with gRPC then the following fully worked example might be informative, https://www.goheroe.org/2017/08/19/grpc-service-discovery-with-server-reflection-and-grpc-cli-in-go.

Two pieces of information are needed in order to make use of the service:

First, you will need the ingress endpoint for your cluster.  The following command sets an environment variable that you will be using as the CLUSTER_INGRESS environment variable across all of the examples within this guide.

<pre><code><b>grpc_cli call $CLUSTER_INGRESS dev.cognizant-ai.experiment.Service.Create "experiment: {uid: 't', name: 'name', description: 'description'}"  --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
</pre></code>

# Manually exercising the server from within the mesh

<pre><code><b>export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
/tmp/grpc_cli call localhost:30001 dev.cognizant-ai.experiment.Service.Get "uid: ''" --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
connecting to localhost:30001
Sending client initial metadata:
authorization : ...
Rpc failed with status code 2, error message: selecting an experiment requires either the DB id or the experiment unique id to be specified stack="[db.go:533 server.go:42 experimentsrv.pb.go:375 auth.go:88 experimentsrv.pb.go:377 server.go:900 server.go:1122 server.go:617]"

<b>/tmp/grpc_cli call $CLUSTER_INGRESS:30001 dev.cognizant-ai.experiment.Service.Get "uid: 't'" --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
connecting to localhost:30001
Sending client initial metadata:
authorization : ...
Rpc failed with status code 2, error message: no matching experiments found for caller specified input parameters
</code></pre>

If you wish to exercise the server while it is deployed into an Istio orchestrated cluster then you should kubectl exec into the istio-proxy to get localized access to the service.  The application server container runs only Alpine so for a full set of tools such as curl and jq using the istio-proxy is a better option, with some minor additions. The Istio container will need the following commands run in order to activate the needed features for local testing:

<pre><code># Copy the grpc CLI tool into the running container
<b>kubectl cp `which grpc_cli` experiments-v1-bc46b5d68-bcdkv:/tmp/grpc_cli -c istio-proxy
kubectl exec -it experiments-v1-bc46b5d68-bcdkv -c istio-proxy /bin/bash
sudo apt-get update
sudo apt-get install -y libgflags2v5 ca-certificates jq
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client
_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:
experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
/tmp/grpc_cli call 100.96.1.14:30001 dev.cognizant-ai.experiment.Service.Get "uid: 't'"  --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
</code></pre>

When Istio is used without a Load balancer the IP of the host on which the pod is running can be determined by using the following command:

<pre><code><b>kubectl -n istio-system get po -l istio=ingress -o jsonpath='{.items[0].status.hostIP}'
</b></code></pre>

## evans expressive grpc client (under development, only for engineering use)

It is recommended that the grpc_cli tool is used by end users as it is more complete at this time, instructions can be found in the cmd/experimentsrv/README.md file.

evans is an end user tool for interacting with grpc servers using a REPL for command line interface.  It is intended to be used for casual testing and inspection by engineering staff who have access to a proto file for a service, and uses the protoc tool.

### Installation

<pre><code><b>go get github.com/ktr0731/evans
</b></code></pre>

### Usage

evans supports several options that will be needed to interact with a remote grpc server:

<pre><code><b>
evans 0.1.2
Usage: evans [--interactive] [--editconfig] [--host HOST] [--port PORT] [--package PACKAGE] [--service SERVICE] [--call CALL] [--file FILE] [--path PATH] [--header HEADER] [PROTO [PROTO ...]]

Positional arguments:
  PROTO                  .proto files

Options:
  --interactive, -i      use interactive mode
  --editconfig, -e       edit config file by $EDITOR
  --host HOST, -h HOST   gRPC host
  --port PORT, -p PORT   gRPC port [default: 50051]
  --package PACKAGE      default package
  --service SERVICE      default service. evans parse package from this if --package is nothing.
  --call CALL, -c CALL   call specified RPC
  --file FILE, -f FILE   the script file which will be executed (used only command-line mode)
  --path PATH            proto file path
  --header HEADER        headers set to each requests
  --help, -h             display this help and exit
  --version              display version and exit
</b></code></pre>

In order to contact a remote service deployed on AWS using the Kubernetes based service mesh you should be familar with the instructions within the grpc_cli description that detail the use of metadata to pass authorization tokens to the service.  The authorization header can be specified using the --header option as follows:

<pre><code><b>
export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
export AUTH0_HEADER="Bearer $AUTH0_TOKEN"
export CLUSTER_INGRESS_HOST=`kubectl get ingress -o wide | tail -1 | awk '{print $3}'`
export CLUSTER_INGRESS_PORT=`kubectl get ingress -o wide | tail -1 | awk '{print $4}'`
evans --interactive --host $CLUSTER_INGRESS_HOST --port $CLUSTER_INGRESS_PORT --header "Authorization"=$AUTH0_HEADER ./cmd/experimentsrv/experimentsrv.proto
</b></code></pre>

# Using IP Port addresses for serving grpc

The grpc listener is configured so that both an IPv4 and IPv6 adapter is attemped individually.  Tgis is because the behavior on some OS's such as Alpine is to not use the 'tcp' protocol with an empty address for both IPv4 and Ipv6, this is the case on Alpine.  Some OS's such as Ubuntu do open both IPv6 and IPv4 in these cases.  As such when using Ubuntu you will always need to specify a single IP, ":30001" for example, to override the default.  The default is designed for use with Alpine containerized deployments such as those in production.

# Functional testing

The server provides some testing for the DB and core functionality of the server.  In order to run this you can use the go test command and point at the relevant go package directories from which you wish to run the tests, for example to run the experiment DB and server tests you could use commands like the following:

<pre><code><b>cd cmd/experimentsrv
export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
LOGXI=*=TRC PGUSER=pl PGHOST=dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com PGDATABASE=platform go test -v . -ip-port ":30001"
</b></code></pre>

# Version management

The experiment server README.md file is designed to be used with the semver utility that is obtained by using 'go get github.com/karlmutch/duat/cmd/semver'.  This utility allows the semantic version string to be bumped at the major, minor, patch or pre-release level.  The build scripts will extract the version from the README.md file and use the version to tag docker images that are being released to AWS ECS or your local docker images.  Also the version strings are present within the experiment server code as well and will be processed by the build scripts to add the semantic version into the compiled binaries.

The semver utility can also be used to manually set your patch levels on every build, or to promote builds.  When using semantic versioning and generating developer versions of builds when you first start from a current release you should first use the 'semver -f [Your README.md] patch' command then add the developer tags using 'semver -f [Your README.md] pre'.  The dev or prerelease versions actually come before the patch version without the pre-release tag in them according to semver sorting.  This means that if you run the patch option again the pre-release tags are stripped off and you will be left with the naked patch version.  Real Semver!

During builds the scripts will run the semver command to retrieve your current version strings from the README.md and then use then to tags things like docker images etc.  It is also used during deployment to inject the versions into deployment yaml files used by kubernetes and Istio as they are being applied to version managed services inside your service mesh or directly on Kubernetes docker image artifacts.  This is really useful for cases where rolling blue/green upgrades are being done etc.

Version management when doing pre-release style builds can result in a lot of images so you'll want to be familiar with the wildcarding features of docker to expunge then selectively from your local repository.

