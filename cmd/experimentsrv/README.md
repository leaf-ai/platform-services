# experimentsrv

The experiment server is used to persist experiment details and to record changes to the state of experiments.  Items included within an experiment include layer definitions and meta-data items.

The experiment server offers a gRPC API that can be accessed using a machine-to-machine or human-to-machine (HCI) interface.  The HCI interface can be interacted with using the grpc_cli tool provided with the gRPC toolkit  More information about grpc_cli can be found at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

# Experiment Deployment

The experiment server makes use of a Postgres DB.  The installation process by default uses a cluster local postgres DB instances.  

In order to deploy against the default postgres in cluster database you will first follow the instructions in the main README.md file and on completion of the databaser initialization using helm deploy the schema.  The envionment variables being used are expected to have been already defined, for example PGRELEASE, as described in the top level README.md file.

<pre><code><b>
kubectl port-forward --namespace default svc/$PGRELEASE-postgresql 5432:5432 &amp;
PGHOST=127.0.0.1 PGDATABASE=platform psql -f sql/platform.sql -d postgres
</b>
Handling connection for 5432
SET
Time: 36.345 ms
psql:/home/kmutch/.psqlrc.local:1: WARNING:  25P01: there is no transaction in progress
LOCATION:  EndTransactionBlock, xact.c:3675
COMMIT
Time: 40.695 ms
SET
Time: 34.603 ms
...
</code></pre>

Once the schema has been successfully created you can move to the next section.

If you wish to use AWS Aurora then you will need to obtain the host, user and password for the database and update your environment variables to use the same.  Including parameters that impact deployment of your Aurora instance include, RDS Endpoint, DB Name, user name, and the password.

The command to install the postgres schema into your DB instance, for AWS Aurora, will appear similar to the following:

<pre><code><b>PGUSER=pl PGHOST=dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com PGDATABASE=platform psql -f sql/platform.sql -d postgres
</b></code></pre>

## Installation

Before starting you should have already installed the duat tools documented in the root README.md file.

### Secrets

You should now create or edit the secrets.yaml file ready for deployment with the user name and the password.  The secrets can be injected into your kubernetes secrets DB using the kubectl apply -f [secrets\_file\_name] command.

In the case of both the default in-cluster database and AWS Aurora the credentials for the user are fixed.  If these need to be changed and the platform.sql file was modified to have the new user access details then you should base64 encode the values and place these inside the supplied secrets.yaml files.

This command will use the PGHOST, and possibly other environment variables from the helm based installation. Should you be using the AWS Aurora deployment then you will need to define the PGHOST environment variable to point at the AWS hostname for your database instance prior to running this command.

Portions of the embeeded secrets have already been defined by the services that will use the database and encoded into the secrets file.  The definition of encoded secrets means that this proof of concept should not be cut and pasted into a production context.

Secrets for these services are currently held within the Kubernetes secrets store and can be populated using the following command:

<pre><code># Read https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables
# create secrets, postgres:username, postgres:password
<b>stencil &lt; ./cmd/experimentsrv/secret.yaml | kubectl create -f - </b>
</code></pre>

Having deployed and defined the secrets and the postgres specific environment variables the postgres client can now used to populate the database schema, these instrctions can be found within the service specific README.md files.  When the PGHOST value has been updated to point at the host name that the service within the deployment cluster can access you are ready to add the secrets to the cluster:

The in-cluster value for the host would appear as follows:

<pre><code><b>
export PGHOST=$PGRELEASE-postgresql.default.svc.cluster.local
</b></code></pre>

### AWS Specific secrets notes

If you were using an external DB then something like the following would be done:

<pre><code><b>
<pre><code><b>export PGHOSTbase64 <(echo -n "dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com")</b>
</b></code></pre>
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

### AWS Specific deployment notes

When using AWS the local workstation should first associate your local docker instance against your AWS ECR account. To do this run the following command using the AWS CLI.  You should change your named AWS_PROFILE to match that set in your ~/.aws/credentials file.

<pre><code><b>export AWS_PROFILE=platform
export AWS_REGION=us-west-2
`aws ecr get-login --no-include-email`
</b></code></pre>

To deploy the experiment service three commands will be used stencil (a SDLC aware templating tool), istioctl (a service mesh administration tool), and kubectl (a cluster orchestration tool):

When version controlled containers are being used with ECS or another docker registry the semver, and stencil tools can be used to extract a git cloned repository that has the version string embeeded inside the README.md or another file of your choice, and then use this with your application deployment yaml specification, as follows:

<pre><code><b>cd ~/project/src/github.com/leaf-ai/platform-services/cmd/experimentsrv</b>
<b>kubectl apply -f <(istioctl kube-inject -f <(stencil < experimentsrv.yaml 2>/dev/null))
</b></code></pre>

This technique can be used to upgrade software versions etc and performing rolling upgrades.

## Service Authentication

Service authentication is introduced inside the top level README.md file for the github.com/leaf-ai/platform-services repository, along with instructions for adding applications and users.  All calls into the experiment service must contain metadata for the autorization brearer token and have the all:experiments claim in order to be accepted.

<pre><code><b>export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_CLIENT_ID=71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj
export AUTH0_CLIENT_SECRET=AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms
export AUTH0_REQUEST=$(printf '{"client_id": "%s", "client_secret": "%s", "audience":"http://api.cognizant-ai.dev/experimentsrv","grant_type":"password", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' "$AUTH0_CLIENT_ID" "$AUTH0_CLIENT_SECRET")
export AUTH0_TOKEN=$(curl -s --request POST --url https://cognizant-ai.auth0.com/oauth/token --header 'content-type: application/json' --data "$AUTH0_REQUEST" | jq -r '"\(.access_token)"')
</b></code></pre>

# Using the service

In order to use the service the gateway should be determined.  The Istio instruction for determining the host and port under various circumstances can be found at https://istio.io/docs/tasks/traffic-management/ingress/#determining-the-ingress-ip-and-ports, and in the following subsections.

If you are using a microk8s based solution the following commands will work:

<pre><code><b>export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
export SECURE_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
export INGRESS_HOST=127.0.0.1
export CLUSTER_INGRESS=$INGRESS_HOST:$INGRESS_PORT
</b></code></pre>

For AWS kops the commands are slightly different:

<pre><code><b>export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export SECURE_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
export CLUSTER_INGRESS=$INGRESS_HOST:$INGRESS_PORT
</b></code></pre>

## grpc\_cli

The grpc\_cli tool can be used to interact with the server for creating and getting experiments.  Other tools do exist as curl like environments for interacting with gRPC servers including the nodejs based tool found at, https://github.com/njpatel/grpcc.  For our purposes we use the less powerful but more common grpc\_cli tool that comes with the gRPC project.  Documentation for the grpc\_cli tool can be found at, https://github.com/grpc/grpc/blob/master/doc/command\_line\_tool.md.

Should you be considering writing or using a service with gRPC then the following fully worked example might be informative, https://www.goheroe.org/2017/08/19/grpc-service-discovery-with-server-reflection-and-grpc-cli-in-go.

Two pieces of information are needed in order to make use of the service:

First, you will need the gateway endpoint for your cluster.  The following command sets an environment variable that you will be using as the CLUSTER_INGRESS environment variable across all of the examples within this guide.

<pre><code><b>grpc_cli call $CLUSTER_INGRESS dev.cognizant_ai.experiment.Service.Create "experiment: {uid: 't', name: 'name', description: 'description'}"  --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
</pre></code>

# Manually exercising the server from within the mesh

<pre><code><b>grpc_cli ls $CLUSTER_INGRESS dev.cognizant_ai.experiment.Service -l</b>

filename: experimentsrv.proto
package: dev.cognizant_ai.experiment;
service Service {
  rpc Create(dev.cognizant_ai.experiment.CreateRequest) returns (dev.cognizant_ai.experiment.CreateResponse) {}
  rpc Get(dev.cognizant_ai.experiment.GetRequest) returns (dev.cognizant_ai.experiment.GetResponse) {}
  rpc MeshCheck(dev.cognizant_ai.experiment.CheckRequest) returns (dev.cognizant_ai.experiment.CheckResponse) {}
}
</code></pre>

<pre><code><b>export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_TOKEN=$(curl -s --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"71eLNu9Bw1rgfYz9PA2gZ4Ji7ujm3Uwj", "client_secret": "AifXD19Y1EKhAKoSqI5r9NWCdJJfyN0x-OywIumSd9hqq_QJr-XlbC7b65rwMjms", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' | jq -r '"\(.access_token)"')
/tmp/grpc_cli call $CLUSTER_INGRESS dev.cognizant_ai.experiment.Service.Get "uid: ''" --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
connecting to localhost:30001
Sending client initial metadata:
authorization : ...
Rpc failed with status code 2, error message: selecting an experiment requires either the DB id or the experiment unique id to be specified stack="[db.go:533 server.go:42 experimentsrv.pb.go:375 auth.go:88 experimentsrv.pb.go:377 server.go:900 server.go:1122 server.go:617]"

<b>/tmp/grpc_cli call $CLUSTER_INGRESS dev.cognizant_ai.experiment.Service.Get "uid: 't'" --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
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
/tmp/grpc_cli call 100.96.1.14:30001 dev.cognizant_ai.experiment.Service.Get "uid: 't'"  --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
</code></pre>

When Istio is used without a Load balancer the IP of the host on which the pod is running can be determined by using the following command:

<pre><code><b>kubectl -n istio-system get po -l istio=ingress -o jsonpath='{.items[0].status.hostIP}'
</b></code></pre>

## evans expressive grpc client (under development, only for engineering use)

It is recommended that the grpc\_cli tool is used by end users as it is more complete at this time, instructions can be found in the cmd/experimentsrv/README.md file.

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

In order to contact a remote service deployed on AWS using the Kubernetes based service mesh you should be familar with the instructions within the grpc\_cli description that detail the use of metadata to pass authorization tokens to the service.  The authorization header can be specified using the --header option as follows:

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

