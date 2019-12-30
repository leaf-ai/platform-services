# platform-services
A public PoC with functioning services using a simple Istio Mesh running on K8s

Version : <repo-version>0.8.1-feature-17-honeycomb-aaaagmibubw</repo-version>

This project is intended as a sand-box for experimenting with Istio and some example services in a similiar manner to what is used by the Cognizant Evolutionary AI services.  It also provides a good way of exposing and testing out non-proprietary platform functions while collaborating with other parties such as vendors and customers.

Because this project is intended to mirror production examples of services deploy it requires that the user has an account and a registered internet domain name.  A service such as domains.google.com, or cloudflare is a good start.  Your DNS registered host will be used to issue certificates on your behalf to secure the public connections that are exposed by services to the internet, and more specifically to secure the username and password based access exposed by the service mesh.

# Purpose

This proof of concept (PoC) implementation is intended as a means by which the LEAF team can experiment with features of Service Mesh, PaaS, and SaaS platforms provided by third parties.  This project serves as a way of exercising non cognizant services so that code can be openly shared while testing external services and technologies, and for support in relation to external open source offerings in a public support context.

In its current form the PoC is used to deploy two main services, an experiment service and a downstream service. These services are provisioned with a gRPC API and leverage an Authorizationm Athentication, and Accounting (AAA) capability and an Observability platform integration to services offered by thrid parties.

This project is intended as a sand-box for experimenting with Istio and some of the services we use in our Evolutionary AI services.  It also provides a good way of exposing and testing out non-proprietary platform functions with other parties such as vendors and customers.

# Installation

These instructions were used with Kubernetes 1.16.x, and Istio 1.4.0.

## Development and Building from source

Clone the repository using the following instructions when this is part of a larger project using multiple services:
<pre><code><b>mkdir ~/project
cd ~/project
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH
mkdir -p src/github.com/leaf-ai
cd src/github.com/leaf-ai
git clone https://github.com/leaf-ai/platform-services
cd platform-services
</b></code></pre>

To boostrap development you will need a copy of Go and the go dependency tools available.  Builds do not need this general however for our purposes we might want to change dependency versions so we should install go and the dep tool, along with several utilities needed when deploying using templates.

Go installation instructions can be found at, https://golang.org/doc/install.

Now download any dependencies, once, into our development environment.

<pre><code><b>go get -u github.com/golang/dep/cmd/dep
go get github.com/karlmutch/duat
go get github.com/karlmutch/petname
dep ensure
go install github.com/karlmutch/duat/cmd/semver
go install github.com/karlmutch/duat/cmd/github-release
go install github.com/karlmutch/duat/cmd/stencil
go install github.com/karlmutch/petname/cmd/petname
</b></code></pre>

## Running the build using the container

Creating a build container to isolate the build into a versioned environment

<pre><code><b>docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
</b></code></pre>

Prior to doing the build a GitHub OAUTH token needs to be defined within your environment.  Use the github admin pages for your account to generate a token, in Travis builds the token is probably already defined by the Travis service.
<pre><code><b>docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -v $GOPATH:/project platform-services ; echo "Done" ; docker container prune -f</b>
</code></pre>

A combined build script is provided 'platform-services/build.sh' to allow all stages of the build including producing docker images to be run together.

# Deployment

These deployment instructions are intended for use with the Ubuntu 18.04 LTS distribution.

The following instructions make use of the stencil tool for templating configuration files.

This major section describes two basic alternatives for deployment, AWS kops, and locally hosted microk8s.  Other Kubernetes distribution and deployment models will work but are not explicitly described here.

## Kubernetes (unmanaged)

The experimentsrv component comes with an Istio definition file for deployment into AWS, or microk8s using Kubernetes (k8s) and Istio.

The deployment definition file can be found at cmd/experimentsrv/experimentsrv.yaml.

Using AWS k8s will use both the kops, and the kubectl tools. You should have an AWS account configured prior to starting deployments, and your environment variables for using the AWS cli should also be done.

The microk8s tooling installation is documented below.

### Verify Docker Version

Docker fopr Ubuntu can be retrieved from the snap store using the following:
<pre><code><b>sudo snap install docker
docker --version</b>
Docker version 18.09.9, build 1752eb3
</code></pre>
You should have a similar or newer version.

## Install Kubectl CLI

Install the kubectl CLI can be done using kubectl 1.16.x version.

<pre><code><b>sudo snap install kubectl --classic</b></code></pre>

Add kubectl autocompletion to your current shell:

<pre><code><b>source \<(kubectl completion bash)</b>
</code></pre>

You can verify that kubectl is installed by executing the following command:

<pre><code><b>kubectl version --client</b>
Client Version: version.Info{Major:"1", Minor:"9", GitVersion:"v1.9.2", GitCommit:"5fa2db2bd46ac79e5e00a4e6ed24191080aa463b", GitTreeState:"clean", BuildDate:"2018-01-18T10:09:24Z", GoVersion:"go1.9.2", Compiler:"gc", Platform:"linux/amd64"}
</code></pre>

## Helm Kubernetes package manager

Helm is used by several packages that are deployed using Kubernetes.  Helm can be installed using instructions found at, https://helm.sh/docs/using\_helm/#installing-helm.  For snap based linux distributions the following can be used as a quick-start.

<pre><code><b>sudo snap install helm --channel=2.16/stable --classic
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --history-max 200 --service-account tiller --upgrade
helm repo update
</b></code></pre>

## Lets Encrypt

letsencrypt is a public SSL/TLS certificate provider that is being used to secure our service mesh for this project. The lets encrypt provisioning tool can be installed from github and accessed to produce TLS certificates for your service.  

Prior to running the lets encrypt tools you should identify the desired hostname and email you wish to make use of.  In our example we have a domain registered as an example, cognizant-ai.net.  This domain is available for management, and we have choosen to use the host name platform-service.cognizant-ai.net as the services hostname.

We first added a registered hosts entry for the platform-services.cognizant-ai.net host into the DNS account, if the host is unknown add an IP address such as 127.0.0.1.  During the generation process you will be prompted to add a DNS TXT record into the custom resource records for the domain, this requires the dummy entry to be present.

Setting up and initiating this process can be done using the following:

<pre><code><b>git clone https://github.com/letsencrypt/letsencrypt</b>
Cloning into 'letsencrypt'...
remote: Enumerating objects: 255, done.
remote: Counting objects: 100% (255/255), done.
remote: Compressing objects: 100% (188/188), done.
remote: Total 71278 (delta 135), reused 110 (delta 65), pack-reused 71023
Receiving objects: 100% (71278/71278), 23.55 MiB | 26.53 MiB/s, done.
Resolving deltas: 100% (52331/52331), done.
<b>cd letsencrypt</b>
<b>./letsencrypt-auto certonly --rsa-key-size 4096 --agree-tos --manual --preferred-challenges=dns --email=karlmutch@cognizant.com -d platform-services.cognizant-ai.net</b>
</code></pre>

You will be prompted with the IP address logging when starting the script, you should choose 'Y' to enabled the logging as this assists auditing of DNS changes on the internet by registras and regulatory bodies.

<pre><code>
Are you OK with your IP being logged?
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
(Y)es/(N)o: <b>Y
</b></code></pre>

After this step you will be asked to add a text record to your DNS records proving that you have control over the domain, this will appear much like the following:

<pre><code>
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Please deploy a DNS TXT record under the name
_acme-challenge.platform-services.cognizant-ai.net with the following value:

mbUa9_gb4RVYhTumHy3zIi3PIXFh0k_oOgCie4NvhqQ

Before continuing, verify the record is deployed.
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Press Enter to Continue
</code></pre>

You should wait 10 to 20 minutes for the TXT record to appear in the database of your DNS provider before selecting continue otherwise the verification will fail and you will need to restart it.

<pre><code>
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/platform-services.cognizant-ai.net/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/platform-services.cognizant-ai.net/privkey.pem
   Your cert will expire on 2020-02-23. To obtain a new or tweaked
   version of this certificate in the future, simply run
   letsencrypt-auto again. To non-interactively renew *all* of your
   certificates, run "letsencrypt-auto renew"
 - If you like Certbot, please consider supporting our work by:

   Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
   Donating to EFF:                    https://eff.org/donate-le
</code></pre>

Once the certificate generation is complete you will have the certificate saved at the location detailed in the 'IMPORTANT NOTES' section.  Keep a record of where this is you will need it later.

After the certificate has been issue feel free to delete the TXT record that served as proof of ownership as it is no longer needed.

## Base cluster installation

This documentation Kubernetes describes two means by which Kubernetes clusters can be installed, choose one however there are many other alternatives also available.

### Installing microk8s Kubernetes

The microk8s solution implements a single host deployment of Kubernetes, https://microk8s.io/. Use snap on Ubuntu to install this component to allow for management of the optional features of microk8s.  When using microk8s the Istio distribution is included in the Kubernetes install as an addon.

The following example details how to configure microk8s once it has been installed:

```
# Allow the storage and registry sub systems and containers within the cluster to communicate.  Also needed for postgres pkg to be fetched and installed
sudo ufw allow in on cbr0 && sudo ufw allow out on cbr0
sudo ufw default allow routed
sudo iptables -P FORWARD ACCEPT
sudo /snap/bin/microk8s.start
sudo /snap/bin/microk8s.enable dashboard dns ingress storage registry istio gpu

microk8s.config >> $HOME/.kube/config
microk8s.kubectl --kubeconfig=$HOME/.kube/config get no
```

### Installing AWS Kubernetes

#### Using kops

If you are using azure or GCP then options such as acs-engine, and skaffold are natively supported by the cloud vendors and written in Go so are readily usable and can be easily customized and maintained and so these are recommended for those cases.

When using AWS the TLS certificates used to secure the connections to your AWS LoadBalancer will require that an ElasticIP is used.  It is recommended that an ElasticIP is allocated for use and then your DNS entries on the domain registra are modified to used the IP as a registered host matching the LetsEncrypt certificate used.  Using an ElasticIP allows the cluster to be regenerated and for the LoadBalancer to be reassociated with the IP whenever the cluster is regenerated.

<pre><code><b>curl -LO https://github.com/kubernetes/kops/releases/download/1.15.0/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops

Add kubectl autocompletion to your current shell:

source <(kops completion bash)
</b></code></pre>

In order to seed your S3 KOPS\_STATE\_STORE version controlled bucket with a cluster definition the following command could be used:

<pre><code><b>export AWS_AVAILABILITY_ZONES="$(aws ec2 describe-availability-zones --query 'AvailabilityZones[].ZoneName' --output text | awk -v OFS="," '$1=$1')"

export S3_BUCKET=kops-platform-$USER
export KOPS_STATE_STORE=s3://$S3_BUCKET
aws s3 mb $KOPS_STATE_STORE
aws s3api put-bucket-versioning --bucket $S3_BUCKET --versioning-configuration Status=Enabled

export CLUSTER_NAME=test-$USER.platform.cluster.k8s.local

kops create cluster --name $CLUSTER_NAME --zones $AWS_AVAILABILITY_ZONES --node-count 1 --node-size=m4.2xlarge --cloud-labels="HostUser=$HOST:$USER"
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
        image: {{.duat.awsecr}}/platform-services/{{.duat.module}}:{{.duat.version}}

Cluster is starting.  It should be ready in a few minutes.

Suggestions:
 * validate cluster: kops validate cluster
 * list nodes: kubectl get nodes --show-labels
 * ssh to the master: ssh -i ~/.ssh/id_rsa admin@api.test.platform.cluster.k8s.local
 * the admin user is specific to Debian. If not using Debian please use the appropriate user based on your OS.
 * read about installing addons at: https://github.com/kubernetes/kops/blob/master/docs/addons.md.

<b>
while [ 1 ]; do
    kops validate cluster > /dev/null && break || sleep 10
done;
</b></code></pre>

The initial cluster spinup will take sometime, use kops commands such as 'kops validate cluster' to determine when the cluster is spun up ready for Istio and the platform services.

#### Istio 1.4.x

If you performed a microk8s installation of Kubernetes do not perform the steps in this subsection.

Istio affords a control layer on top of the k8s data plane.  Instructions for deploying Istio are the vanilla instructions that can be found at, https://istio.io/docs/setup/getting-started/#install.  Istio was at one time a Helm based installation but has since moved to using its own methodology.

<pre><code><b>cd ~
curl -LO https://github.com/istio/istio/releases/download/1.4.0/istio-1.4.0-linux.tar.gz
tar xzf istio-1.4.0-linux.tar.gz
export ISTIO_DIR=`pwd`/istio-1.4.0
export PATH=$ISTIO_DIR/bin:$PATH
cd -
istioctl manifest apply --set profile=demo --set values.tracing.enabled=true --set values.tracing.provider=zipkin --set values.global.tracer.zipkin.address=honeycomb-opentracing-proxy.default.svc.cluster.local:9411 --set values.pilot.traceSampling=100 --set values.gateways.istio-egressgateway.enabled=false --set values.gateways.istio-ingressgateway.sds.enabled=true
</b></code></pre>

## Configuration of secrets

This project makes use of several secrets that are used to access resources under its control, including the Postgres Database, the Honeycomb service, and the lets encrypt issues certificate.

The experiment service Honeycomb observability solution uses a key to access Datasets defined by the Honeycomb account and store events in the same.  Configuring the service is done by creating a Kubernetes secret.  For now we can define the Honeycomb API key using an environment variable and when we deploy the secrets for the Postgres Database the secret for the API will be injected using the stencil tool.

<pre><code><b>export O11Y_KEY a54d762df847474b22915
</b></code></pre>

The services also use a postgres Database instance to persist experiment data.  The following shows an example of what should be defined for Postgres support prior to running the stencil command:

<pre><code><b>export PGRELEASE=`petname`
export PGHOST=$PGRELEASE-postgresql.default.svc.cluster.local
export PORT=5432
export PGUSER=postgres
export PGPASSWORD=p355w0rd
export PGDATABASE=platform
</b></code></pre>

<pre><code><b>
stencil < cmd/experimentsrv/secret.yaml | kubectl apply -f -
</b></code></pre>

The last set of secrets that need to be stored are related to securing the mesh for third party connections using TLS.  This secret contains the full certificate chain and private key needed to implement TLS on the gRPC connections exposed by the mesh.

<pre><code><b>
sudo kubectl create -n istio-system secret generic platform-services-tls-cert \
    --from-file=key=/etc/letsencrypt/live/platform-services.cognizant-ai.net/privkey.pem \
    --from-file=cert=/etc/letsencrypt/live/platform-services.cognizant-ai.net/fullchain.pem
</b></code></pre>

## Deploying the Observability proxy server

This proxy server is used to forward tracing and metrics from your istio mesh based deployment to the Honeycomb service.

<pre><code><b>
stencil < honeycomb-opentracing-proxy.yaml | kubectl apply -f -
stencil < new-telemetry.yaml | kubectl apply -f -
stencil < honeycomb-agent.yaml | kubectl apply -f -</b></code></pre>

</b></code></pre>

In order to instrument the base Kubernetes deployment for us with honeycomb you should follow the instructions found at https://docs.honeycomb.io/getting-data-in/integrations/kubernetes/.

### Postgres DB

To deploy the platform experiment service a database must be present.  The PoC is intended to use an in-cluster DB designed that is dropped after the service is destroyed.

If you wish to use Aurora then you will need to use the AWS CLI Tools or the Web Console to create your Postgres Database appropriately, and then to set your environment variables PGHOST, PGPORT, PGUSER, PGPASSWORD, and PGDATABASE appropriately.  You will also be expected to run the sql setup scripts yourself.

The first step is to install the postgres 11 client on your system and then to populate the schema on the remote database:

<pre><code><b></b>
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list'
sudo apt-get -y update
sudo apt-get upgrade postgresql-client-11
</code></pre>

### Deploying an in-cluster Database

This section gives guidence on how to install an in-cluster database for use-cases where data persistence beyond a single deployment is not a concern.  These instructions are therefore limited to testing only scenarios.  For information concerning Kubernetes storage strategies you should consult other sources and read about stateful sets in Kubernetes.  In production using a single source of truth then cloud provider offerings such as AWS Aurora are recommended.

A secrets file containing host information, passwords and other secrets is assumed to have already been applied using the instructions several sections above.  The secrets are needed to allows access to the postgres DB, and/or other external resources.  YAML files will be needed to populate secrets into the service mesh, individual services document the secrets they require within their README.md files found on github and provide examples, for example https://github.com/leaf-ai/platform-services/cmd/experimentsrv/README.md.

In order to deploy Postgres this document describes a helm based approach.  The bitnami postgresql distribution can be installed using the following:

<pre><code><b>
helm install --name $PGRELEASE \
  --set postgresqlPassword=$PGPASSWORD,postgresqlDatabase=postgres\
  stable/postgresql
</b></code></pre>

Special note should be taken of the output from the helm command it has a lot of valuable information concerning your postgres deployment that will be needed when you load the database schema.

<pre><code>
NAME:   internal-seasnail
LAST DEPLOYED: Tue May 21 14:25:32 2019
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/Pod(related)
NAME                            READY  STATUS   RESTARTS  AGE
internal-seasnail-postgresql-0  0/1    Pending  0         0s

==> v1/Secret
NAME                          TYPE    DATA  AGE
internal-seasnail-postgresql  Opaque  1     0s

==> v1/Service
NAME                                   TYPE       CLUSTER-IP     EXTERNAL-IP  PORT(S)   AGE
internal-seasnail-postgresql           ClusterIP  100.64.75.251  <none>       5432/TCP  0s
internal-seasnail-postgresql-headless  ClusterIP  None           <none>       5432/TCP  0s

==> v1beta2/StatefulSet
NAME                          READY  AGE
internal-seasnail-postgresql  0/1    0s


NOTES:
** Please be patient while the chart is being deployed **

PostgreSQL can be accessed via port 5432 on the following DNS name from within your cluster:

    internal-seasnail-postgresql.default.svc.cluster.local - Read/Write connection
To get the password for "postgres" run:

    export POSTGRES_PASSWORD=$(kubectl get secret --namespace default internal-seasnail-postgresql -o jsonpath="{.data.postgresql-password}" | base64 --decode)

To connect to your database run the following command:

    kubectl run internal-seasnail-postgresql-client --rm --tty -i --restart='Never' --namespace default --image docker.io/bitnami/postgresql:11.3.0 --env="PGPASSWORD=$POSTGRES_PASSWORD" --command -- psql --host internal-seasnail-postgresql -U $PGUSER -d $PGDATABASE



To connect to your database from outside the cluster execute the following commands:

    kubectl port-forward --namespace default svc/internal-seasnail-postgresql 5432:5432 &
    PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U $PGUSER -d $PGDATABASE
</code></pre>

Setting up the proxy will be needed prior to running the SQL database provisioning scripts.  When doing this prior to running the postgres client set the PGHOST environment variable to 127.0.0.1 so that the proxy on the localhost is used.  The proxy will timeout after inactivity and shutdown so be prepared to restart it when needed.

<pre><code><b>
kubectl wait --for=condition=Ready pod/$PGRELEASE-postgresql-0
kubectl port-forward --namespace default svc/$PGRELEASE-postgresql 5432:5432 &amp;
PGHOST=127.0.0.1 PGDATABASE=platform psql -f sql/platform.sql -d postgres
</b></code></pre>

Further information about how to deployed the service specific database for the experiment service for example can be found in the cmd/experiment/README.md file.

## Deploying into the Istio mesh

### Configuring the service DNS

When using this mesh instance with a TLS based deployment the DNS domain name used for the LetsEncrypt certificate (CN), will need to have its address record (A) updated to point at the AWS load balancer assigned to the Kubernetes cluster.  In AWS this is done via Route53:

<pre><code><b>
INGRESS_HOST=`kubectl get svc --namespace istio-system -o go-template='{{range .items}}{{range .status.loadBalancer.ingress}}{{.hostname}}{{printf "\n"}}{{end}}{{end}}'`
dig +short $INGRESS_HOST
</b></code></pre>

Take the IP addresses from the above output and use these as the A record for the LetsEncrypt host name and this will enable accessing the mesh and validation of the common name (CN) in the certificate.

### Service deployment overview

Platform services use Dockerfiles to encapsulate their build steps which are documented within their respective README.md files.  Building services are single step CLI operations and require only the installation of Docker, and any version of Go 1.7 or later.  Builds will produce containers and will upload these to your current AWS account users ECS docker registry.  Deployments are staged from this registry.  

<pre><code><b>kubectl get nodes</b>
NAME                                           STATUS    ROLES     AGE       VERSION
ip-172-20-118-127.us-west-2.compute.internal   Ready     node      17m       v1.9.3
ip-172-20-41-63.us-west-2.compute.internal     Ready     node      17m       v1.9.3
ip-172-20-55-189.us-west-2.compute.internal    Ready     master    18m       v1.9.3
</code></pre>

Once secrets are loaded individual services can be deployed from a checked out developer copy of the service repo using a command like the following :

<pre><code><b>cd ~/project/src/github.com/leaf-ai/platform-services</b>
<b>cd cmd/[service] ; kubectl apply -f \<(istioctl kube-inject -f \<( stencil [service].yaml 2>/dev/null)" ); cd - </b>
</code></pre>

In order to locate the image repository the stencil tool will test for the presence of AWS credentials and if found will use the account as the source of AWS ECR images.  In the case where the credentials are not present then the default microk8s registry will be used for image deployment.

Once the application is deployed you can discover the gateway points within the kubernetes cluster by using the kubectl commands as documented in the cmd/experimentsrv/README.md file.

More information about deploying a real service and using the experimentsrv server can be found at, https://github.com/leaf-ai/platform-services/blob/master/cmd/experimentsrv/README.md.

### Debugging

There are several pages of debugging instructions that can be used for situations when grpc failures to occur without much context, this applies to unexplained GRPC errors that reference C++ files within envoy etc.  These pages can be found by using the search function on the Istio web site at, https://istio.io/search.html?q=debugging.

You might find the following use cases useful for avoiding using hard coded pod names etc when debugging.

The following example shows enabling debugging for http2 and rbac layers within the Ingress Envoy instance.

<pre><code><b>
kubectl exec $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -c istio-proxy -n istio-system -- curl -X POST "localhost:15000/logging?rbac=debug" -s
kubectl exec $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -c istio-proxy -n istio-system -- curl -X POST "localhost:15000/logging?filter=debug" -s
kubectl exec $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -c istio-proxy -n istio-system -- curl -X POST "localhost:15000/logging?http2=debug" -s
</b></code></pre>

After making a test request the log can be retrieved using something like the following:

<pre><code><b>
kubectl logs $(kubectl get pods --namespace istio-system -l istio=ingressgateway -o jsonpath='{.items[0].metadata.name}') --namespace istio-system</b></code></pre>

When debugging the istio proxy side cars for services you can do the following to enable all of the modules within the proxy:

<pre><code><b>
kubectl exec $(kubectl get pods -l app=experiment -o jsonpath='{.items[0].metadata.name}') -c istio-proxy -- curl -X POST "localhost:15000/logging?level=debug" -s</b></code></pre>

And then the logs can be captured during the testing using the following:

<pre><code><b>
kubectl logs $(kubectl get pods -l app=experiment -o jsonpath='{.items[0].metadata.name}') -c istio-proxy</b></code></pre>

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
./experimentsrv built at 2018-01-18\_15:22:47+0000, against commit id 34e761994b895ac48cd832ac3048854a671256b0
2018-01-18T16:50:18+0000 INF experimentsrv git hash version 34e761994b895ac48cd832ac3048854a671256b0 _: [host experiments-v1-bc46b5d68-tltg9]
2018-01-18T16:50:18+0000 INF experimentsrv database startup / recovery has been performed dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com:5432 name platform _: [host experiments-v1-bc46b5d68-tltg9]
2018-01-18T16:50:18+0000 INF experimentsrv database has 1 connections  dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com:5432 name platform dbConnectionCount 1 _: [host experiments-v1-bc46b5d68-tltg9]
2018-01-18T16:50:33+0000 INF experimentsrv database has 1 connections  dev-platform.cluster-cff2uhtd2jzh.us-west-2.rds.amazonaws.com:5432 name platform dbConnectionCount 1 _: [host experiments-v1-bc46b5d68-tltg9]
</code></pre>

The container name can also include the istio mesh and kubernetes installed system containers for indepth debugging purposes.

### Kubernetes Web UI and console

In addition to the kops information for a cluster being hosted on S3, the kubectl information for accessing the cluster is stored within the ~/.kube directory.  The web UI can be deployed using the instruction at https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/#deploying-the-dashboard-ui, the following set of instructions include the deployment as it stood at k8s 1.9.  Take the opportunity to also review the document at the above location.

Kubectl service accounts can be created at will and given access to cluster resources.  To create, authorize and then authenticate a service account the following steps can be used:

```
kubectl create -f https://raw.githubusercontent.com/kubernetes/heapster/release-1.5/deploy/kube-config/influxdb/influxdb.yaml
kubectl create -f https://raw.githubusercontent.com/kubernetes/heapster/release-1.5/deploy/kube-config/influxdb/heapster.yaml
kubectl create -f https://raw.githubusercontent.com/kubernetes/heapster/release-1.5/deploy/kube-config/influxdb/grafana.yaml
kubectl create -f https://raw.githubusercontent.com/kubernetes/heapster/release-1.5/deploy/kube-config/rbac/heapster-rbac.yaml
kubectl create -f https://raw.githubusercontent.com/kubernetes/dashboard/master/src/deploy/recommended/kubernetes-dashboard.yaml
kubectl create serviceaccount studioadmin
secret_name=`kubectl get serviceaccounts studioadmin -o json | jq '.secrets[] | [.name] | join("")' -r`
secret_kube=`kubectl get secret $secret_name -o json | jq '.data.token' -r | base64 --decode`
# The following will open up all service accounts for admin, review the k8s documentation specific to your
# install version of k8s to narrow the roles
kubectl create clusterrolebinding serviceaccounts-cluster-admin --clusterrole=cluster-admin --group=system:serviceaccounts
```

The value in secret kube can be used to login to the k8s web UI.  First start 'kube proxy' in a terminal window to create a proxy server for the cluster.  Use a browser to navigate to http://localhost:8001/ui.  Then use the value in the secret\_kube variable as your 'Token' (Service Account Bearer Token).

You will now have access to the Web UI for your cluster with full privs.


# AAA using Auth0

Platform services are secured using the Auth0.com service.  Auth0 is a service that provides support for headless machine to machine authentication.  Auth0 is being used initially to provide Bearer tokens for both headless and CLI clients to platform services proof of concept.

Auth0 supports the ability to create a hosted database for storing user account and credential information.  You should navigate to the Connections -> Database section and create a database with the name of "Username-Password-Authentication".  This will be used later when creating applications as your source of user information.

Auth0 authorizations can be done using a Demo auth0.com account.  To do this you will need to add a custom API to the Auth0 account, call it something like "Experiments API" and give it an identifier of "http://api.cognizant-ai.net/experimentsrv", you should also enable RBAC and the "Add Permissions in the Access Token" options.  Then use the save button to persist the new API.  Identifiers are used as the 'audience' setting when generating tokens via web calls against the AAA features of the Auth0 platform.

The next stop is to use the menu bar to select the "Permissions" tab.  This tab allows you to create a scope to be used for the permissions granted to user.  Create a scope called "all:experiments" with a description, and select the Add button.  This scope will become available for use by authenticated user roles to allow them to access the API.

Next, click the "Machine To Machine Applications" tab.  This should show that a new Test Application has been created and authorized against this API.  To the right of the Authorized switch is a drop down button that can be used to expose more detailed information related to the scopes that are permitted via this API.  You should see that the all:experiments scope is not yet selected, select it and then use the update button.

Now navigate using the left side panel to the Applications screen.  Click to select your new "Experiments (Test Application)".  The screen displayed as a result will show the Client ID, and the Client secret that will be needed later on, take a note of thes values as they will be needed during AAA operation.  Go to the bottom of the page and you will be able to expose some advanced settings".  Inside the advanced settings you will see a ribbon bar with a "Grant Types" tab that can be clicked on revealing the selections available for grant type, ensure that the "password" radio button is selected to enable passwords for authentication, and click on the Save Changes button to save the selection.

The first API added by the Auth0 platform will be the client that accesses the Auth0 service itself providing per user authentication and token generation. When you begin creating a client you will be able to select the "Auth0 Management API" as one of the APIs you wish to secure.

The next step is to create a User and assign the user a role.  The left hand panel has a "Users & Roles" menu.  Using the menu you can select the "User" option and then use the "CREATE USER" button on the right side of the screen.  This where the real power of the Auth0 platform comes into play as you can use your real email address and perform operations related to identity management and passwords resets without needing to implement these features yourself.  When creating the user the connection field should be filled in with the Database connection that you created initially in these instructions. "Username-Password-Authentication".  After creating your User you can go to the Users panel and click on the email, then click on the permissions tab.  Add the all:experiments permission to the users prodile using the "ASSIGN PERMISSIONS" button.

You can now use various commands to manipulate the APIs outside of what will exist in the application code, this is a distinct advantage over directly using enterprise tools such as Okta.  Should you wish to use Okta as an Identity provider, or backend, to Auth0 then this can be done however you will need help from our Tech Ops department to do this and is an expensive option.  At this time the user and passwords being used for securing APIs can be managed through the Auth0 dashboard including the ability to invite users to become admins.

<pre><code><b>
export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_CLIENT_ID=pL3iSUmOB7EPiXae4gPfuEasccV7PATs
export AUTH0_CLIENT_SECRET=KHSCFuFumudWGKISCYD79ZkwF2YFCiQYurhjik0x6OKYyOb7TkfGKJrHKXXADzqG
export AUTH0_REQUEST=$(printf '{"client_id": "%s", "client_secret": "%s", "audience":"http://api.cognizant-ai.net/experimentsrv","grant_type":"password", "username": "karlmutch@gmail.com", "password": Ap9ss2345f"", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' "$AUTH0_CLIENT_ID" "$AUTH0_CLIENT_SECRET")
export AUTH0_TOKEN=$(curl -s --request POST --url https://cognizant-ai.auth0.com/oauth/token --header 'content-type: application/json' --data "$AUTH0_REQUEST" | jq -r '"\(.access_token)"')

</b>
c.f. https://auth0.com/docs/quickstart/backend/golang/02-using#obtaining-an-access-token-for-testing.
</code></pre>

If you are using the test API and you are either running a kubectl port-forward or have a local instance of the postgres DB, you can do something like:

<pre><code><b>kubectl port-forward --namespace default svc/$PGRELEASE-postgresql 5432:5432 &
cd cmd/downstream
go run . --ip-port=":30008" &
cd ../..
cd cmd/experimentsrv
export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_CLIENT_ID=pL3iSUmOB7EPiXae4gPfuEasccV7PATs
export AUTH0_CLIENT_SECRET=KHSCFuFumudWGKISCYD79ZkwF2YFCiQYurhjik0x6OKYyOb7TkfGKJrHKXXADzqG
export AUTH0_REQUEST=$(printf '{"client_id": "%s", "client_secret": "%s", "audience":"http://api.cognizant-ai.net/experimentsrv","grant_type":"password", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' "$AUTH0_CLIENT_ID" "$AUTH0_CLIENT_SECRET")
export AUTH0_TOKEN=$(curl -s --request POST --url https://cognizant-ai.auth0.com/oauth/token --header 'content-type: application/json' --data "$AUTH0_REQUEST" | jq -r '"\(.access_token)"')
go test -v --dbaddr=localhost:5432 -ip-port="[::]:30007" -dbname=platform -downstream="[::]:30008"
</b></code></pre>

## Auth0 claims extensibility

Auth0 can be configured to include additional headers with user metadata such as email addresses etc using custom rules in the Auth0 rules configuration.  Header that are added can be queried and extracted from gRPC HTTP authorization header meta data as shown in the experimentsrv server.go file. An example of a rule is as follows:

<pre><code>
function (user, context, callback) {
  context.accessToken["http://cognizant-ai.dev/user"] = user.email;
  callback(null, user, context);
 }</code></pre>

 An example of extracting this item on the gRPC client side can be found in cmd/experimentsrv/server.go in the GetUserFromClaims function.

# Manually invoking and using production services with TLS

When using the gRPC services within a secured cluster these instructions can be used to access and exercise the services.

A very basic test of the TLS can be done using the curl command:

<pre><code><b>
curl -Iv https://helloworld.letsencrypt.org
</b></code></pre>

An example client for running a simple ping style test against the cluster is provided in the cmd/cli-experiment directory.  This client acts as a test for the presence of the service.  If the commands to obtain a JWT have been followed then this command can be run against the cluster as follows:

<pre><code><b>
cd cmd/cli-experiment
go run . --server-addr=platform-services.cognizant-ai.net:443 --auth0-token="$AUTH0_TOKEN"</b>
(*dev_cognizant_ai_experiment.CheckResponse)(0xc00003c280)(modules:"downstream" )
</code></pre>



# Manually invoking and using services without TLS

When using the gRPC services within an unsecured cluster these instructions can be used to access and exercise the services.

A pre-requiste of manually invoking GRPC servcies is that the grpc_cli tooling is installed.  The instructions for doing this can be found within the grpc source code repository at, https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md.

The following instructions identify a $INGRESS_HOST value for cases where a LoadBalancer is being used.  If you are using minikube or microk8s and the cluster is hosted locally then the INGRESS_HOST value should be 127.0.0.1 for the following instructions.

Services used within the platform require that not only is the link integrity and security is maintained using mTLS but that an authorization block is also supplied to verify the user requesting a service.  The authorization can be supplied when using the gRPC command line tool using the metadata options.  First we retrieve a token using curl and then make a request against the service, run in this case as a local docker container, as follows:

<pre><code><b>grpc_cli call localhost:30001 dev.cognizant-ai.experiment.Service.Get "id: 'test'" --metadata authorization:"Bearer $AUTH0_TOKEN"
</b></code></pre>

The services used within the platform all support reflection when using gRPC.  To examine calls available for a server you should first identify the endpoint through which the gateway is being routed, in this case as part of an Istio cluster on AWS, for example:

<pre><code><b>export SECURE_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')
export INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}')
export CLUSTER_INGRESS=$INGRESS_HOST:$SECURE_INGRESS_PORT
<b>grpc_cli ls $CLUSTER_INGRESS -l</b>
filename: experimentsrv.proto
package: dev.cognizant_ai.experiment;
service Service {
  rpc Create(dev.cognizant_ai.experiment.CreateRequest) returns (dev.cognizant_ai.experiment.CreateResponse) {}
  rpc Get(dev.cognizant_ai.experiment.GetRequest) returns (dev.cognizant_ai.experiment.GetResponse) {}
  rpc MeshCheck(dev.cognizant_ai.experiment.CheckRequest) returns (dev.cognizant_ai.experiment.CheckResponse) {}
}

filename: grpc/health/v1/health.proto
package: grpc.health.v1;
service Health {
  rpc Check(grpc.health.v1.HealthCheckRequest) returns (grpc.health.v1.HealthCheckResponse) {}
  rpc Watch(grpc.health.v1.HealthCheckRequest) returns (stream grpc.health.v1.HealthCheckResponse) {}
}

filename: grpc_reflection_v1alpha/reflection.proto
package: grpc.reflection.v1alpha;
service ServerReflection {
  rpc ServerReflectionInfo(stream grpc.reflection.v1alpha.ServerReflectionRequest) returns (stream grpc.reflection.v1alpha.ServerReflectionResponse) {}
}
</code></pre>

To drill further into interfaces and examine the types being used within calls you can perform commands such as:

<pre><code><b>grpc_cli type $CLUSTER_INGRESS dev.cognizant_ai.experiment.CreateRequest -l</b>
message CreateRequest {
.dev.cognizant_ai.experiment.Experiment experiment = 1[json_name = "experiment"];
}
<b>grpc_cli type $CLUSTER_INGRESS dev.cognizant_ai.experiment.Experiment -l</b>
message Experiment {
string uid = 1[json_name = "uid"];
string name = 2[json_name = "name"];
string description = 3[json_name = "description"];
.google.protobuf.Timestamp created = 4[json_name = "created"];
map&lt;uint32, .dev.cognizant_ai.experiment.InputLayer&gt; inputLayers = 5[json_name = "inputLayers"];
map&lt;uint32, .dev.cognizant_ai.experiment.OutputLayer&gt; outputLayers = 6[json_name = "outputLayers"];
}
<b>grpc_cli type $CLUSTER_INGRESS dev.cognizant_ai.experiment.InputLayer -l</b>
message InputLayer {
enum Type {
	Unknown = 0;
	Enumeration = 1;
	Time = 2;
	Raw = 3;
}
string name = 1[json_name = "name"];
.dev.cognizant_ai.experiment.InputLayer.Type type = 2[json_name = "type"];
repeated string values = 3[json_name = "values"];
}
</code></pre>


# Shutting down a service, or cluster

<pre><code><b>kubectl delete -f experimentsrv.yaml
</b></code></pre>

<pre><code><b>kops delete cluster $CLUSTER_NAME --yes
</b></code></pre>
