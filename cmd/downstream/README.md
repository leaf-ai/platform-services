# downstrearm

The downstream server is used to exercise the internal service mesh communications and support just a simple ping request.

The server offers a gRPC interface accessed using a machine-to-machine or human-to-machine (HCI) interface.  The HCI interface can be interacted with using the grpc_cli tool provided with the gRPC toolkit  More information about grpc_cli can be found at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

## Installation

Before starting you should install several build and deployment tools that will be useful for managing the service configuration file.

<pre><code><b>wget -O $GOPATH/bin/stencil https://github.com/karlmutch/duat/releases/download/0.4.0/stencil
chmod +x $GOPATH/bin/stencil
</b></code></pre>

### Deployment

The downstream service is deployed using Istio into a Kubernetes (k8s) cluster.  The k8s cluster installation instructions can be found within the README.md file at the top of this github repository.

When using AWS the local workstation should first associate your local docker instance against your AWS ECR account. To do this run the following command using the AWS CLI.  You should change your named AWS_PROFILE to match that set in your ~/.aws/credentials file.

<pre><code><b>export AWS_PROFILE=platform
export AWS_REGION=us-west-2
`aws ecr get-login --no-include-email`
</b></code></pre>

To deploy the service three commands will be used stencil (a SDLC aware templating tool), istioctl (a service mesh administration tool), and kubectl (a cluster orchestration tool):

When version controlled containers are being used with ECS or another docker registry the semver can be used to extract a git cloned repository that has the version string embeeded inside the README.md or another file of your choice, and then use this with your application deployment yaml specification, as follows:

<pre><code><b>cd ~/project/src/github.com/SentientTechnologies/platform-services/cmd/downstream</b>
<b>kubectl apply -f <(istioctl kube-inject --includeIPRanges="172.20.0.0/16"  -f <(stencil < downstream.yaml))
</b></code></pre>

This technique can be used to upgrade software versions etc and performing rolling upgrades.

# Using the service

In order to use the service you will need to use kubectl to attach to the running container of other services within the cluster.  Having done then use the grpc_cli commands to make RPC calls to the service testing the networking policies that have been created for the mesh.

The grpc_cli tool can be used to interact with the server for making ping calls to probe and test access to the service.  Other tools do exist as curl like environments for interacting with gRPC servers including the nodejs based tool found at, https://github.com/njpatel/grpcc.  For our purposes we use the less powerful but more common grpc_cli tool that comes with the gRPC project.  Documentation for the grpc_cli tool can be found at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

Should you be considering writing or using a service with gRPC then the following fully worked example might be informative, https://www.goheroe.org/2017/08/19/grpc-service-discovery-with-server-reflection-and-grpc-cli-in-go.

Two pieces of information are needed in order to make use of the service:

First, you will need the ingress iendpoint for your cluster.  The following command sets an environment variable that you will be using as the CLUSTER_INGRESS environment variable across all of the examples within this guide.

<pre><code><b>grpc_cli call $CLUSTER_INGRESS ai.sentient.experiment.Service.Create "experiment: {uid: 't', name: 'name', description: 'description'}"  --metadata authorization:"Bearer $AUTH0_TOKEN"</b>
</pre></code>
