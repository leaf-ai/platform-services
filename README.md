# platform-services
A PoC with functioning service using simple Istio Mesh running on K8s

Version : <repo-version>0.5.0</repo-version>

# Installation

<pre><code><b>go get github.com/leaf-ai/platform-services
</b></code></pre>

# Development and Building from source

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

Creating a build container to isolate the build into a versioned environment

<pre><code><b>docker build -t platform-services:latest --build-arg USER=$USER --build-arg USER_ID=`id -u $USER` --build-arg USER_GROUP_ID=`id -g $USER` .
</b></code></pre>

Running the build using the container

Prior to doing the build a GitHub OAUTH token needs to be defined within your environment.  Use the github admin pages for your account to generate a token, in Travis builds the token is probably already defined by the Travis service.
<pre><code><b>docker run -e GITHUB_TOKEN=$GITHUB_TOKEN -v $GOPATH:/project platform-services ; echo "Done" ; docker container prune -f</b>
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

Install the kubectl CLI can be done using kubectl 1.11.x version.

<pre><code><b> curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.11.1/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/</b>
</code></pre>

Add kubectl autocompletion to your current shell:

<pre><code><b>source \<(kubectl completion bash)</b>
</code></pre>

You can verify that kubectl is installed by executing the following command:

<pre><code><b>kubectl version --client</b>
Client Version: version.Info{Major:"1", Minor:"9", GitVersion:"v1.9.2", GitCommit:"5fa2db2bd46ac79e5e00a4e6ed24191080aa463b", GitTreeState:"clean", BuildDate:"2018-01-18T10:09:24Z", GoVersion:"go1.9.2", Compiler:"gc", Platform:"linux/amd64"}
</code></pre>

### Install kops

At the time this guide was updated kops 1.11.1 was released, if you are reading this guide in April of 2018 or later look for the release version of kops 1.9 or later.  kops for the AWS use case at the alpha is a very restricted use case for our purposes and works in a stable fashion.  If you are using azure or GCP then options such as acs-engine, and skaffold are natively supported by the cloud vendors and written in Go so are readily usable and can be easily customized and maintained and so these are recommended for those cases.

<pre><code><b>curl -LO https://github.com/kubernetes/kops/releases/download/1.11.1/kops-linux-amd64
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
        image: {{.duat.awsecr}}/platform-services/{{.duat.module}}:{{.duat.version}}

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

Istio affords a control layer on top of the k8s data plane.  Instructions for deploying Istio are the vanilla instructions that can be found at, 
https://archive.istio.io/v1.0/docs/setup/kubernetes/quick-start/#prerequisites.
We recommend using the mTLS installation for the k8s cluster deployment, for example

<pre><code><b>cd ~
curl -LO https://github.com/istio/istio/releases/download/1.0.5/istio-1.0.5-linux.tar.gz
tar xzf istio-1.0.5-linux.tar.gz
export ISTIO_DIR=`pwd`/istio-1.0.5
export PATH=$ISTIO_DIR/bin:$PATH
cd -
kubectl apply -f $ISTIO_DIR/install/kubernetes/helm/istio/templates/crds.yaml
kubectl apply -f $ISTIO_DIR/install/kubernetes/istio-demo-auth.yaml
</b></code></pre>

In any custom resources are not applied or updated repeat the apply after waiting for a few seconds for the CRDs to get loaded.

## Deploying a straw-man service into the Istio control plane

### Postgres DB

To deploy the platform experiment service a database must be present.  The PoC is intended to use an in-cluster DB designed that is dropped after the service is destroyed.

If you wish to use Aurora then you will need to use the AWS CLI Tools or the Web Console to create your Postgres Database appropriately, and then to set your environment variables PGHOST, PGPORT, PGUSER, PGPASSWORD, and PGDATABASE appropriately.  You will also be expected to run the sql setup scripts yourself.

The first step is to install the postgres 11 client on your system and then to populate the schema on the remote database:

<pre><code><b></b>
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list'
sudo apt-get upgrade postgresql-client-11
</code></pre>

### Deploying an in-cluster Database

This section gives guidence on how to install an in-cluster database for use-cases where data persistence beyond a single deployment is not a concern.  These instructions are therefore limited to testing only scenarios.  For information concerning Kubernetes storage strategies you should consult other sources and read about stateful sets in Kubernetes.  In production using a single source of truth then cloud provider offerings such as AWS Aurora are recommended.

In order to deploy Postgres this document describes a helm based approach.  Helm can be installed using instructions found at, https://helm.sh/docs/using\_helm/#installing-helm.  For snap based linux distributions the following can be used as a quick-start.

<pre><code><b>sudo snap install helm --classic
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --history-max 200 --service-account tiller --upgrade
helm repo update
</b></code></pre>

Now moving to postgres the bitnami distribution can be installed using the following:

<pre><code><b>export PGRELEASE=`petname`
export PGHOST=$PGRELEASE-postgresql.default.svc.cluster.local
export PORT=5432
export PGUSER=postgres
export PGPASSWORD=p355w0rd
export PGDATABASE=postgres
helm install --name $PGRELEASE \
  --set postgresqlPassword=$PGPASSWORD,postgresqlDatabase=$PGDATABASE\
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
kubectl port-forward --namespace default svc/$PGRELEASE-postgresql 5432:5432 &amp;
</b></code></pre>

Further information about how to deployed the service specific database for the experiment service for example can be found in the cmd/experiment/README.md file.

### Configuring the database secrets overview

A secrets file containing host information, passwords and other secrets will be needed to allows access to the postgres DB, and/or other external resources.  YAML files will be needed to populate secrets into the service mesh, individual services document the secrets they require within their README.md files found on github and provide examples, for example https://github.com/leaf-ai/platform-services/cmd/experimentsrv/README.md.

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
<b>kubectl apply -f \<(istioctl kube-inject -f [cmd/[service]/application-deployment-yaml] --includeIPRanges="172.20.0.0/16" ) </b>
</code></pre>

Once the application is deployed you can discover the ingress points within the kubernetes cluster by using the following:
<pre><code><b>export CLUSTER_INGRESS=`kubectl get ingress -o wide | tail -1 | awk '{print $3":"$4}'`
</b></code></pre>

More information about deploying a real service and using the experimentsrv server can be found at, https://github.com/leaf-ai/platform-services/blob/master/cmd/experimentsrv/README.md.

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

Platform services are secured using the Auth0 service.  Auth0 is a service that provides support for headless machine to machine authentication.  Auth0 is being used initially to provide Bearer tokens for both headless and CLI clients to platform services proof of concept.

Auth0 supports the ability to create a hosted database for storing user account and credential information.  You should navigate to the Connections -> Database section and create a database with the name of "Username-Password-Authentication".  This will be used later when creating applications as your source of user information.

Auth0 authorizations can be done using a Demo auth0.com account.  To do this you will need to add a custom API to the Auth0 account, call it something like "Experiments API" and give it an identifier of "http://api.cognizant-ai.dev/experimentsrv", you should also enable RBAC and the "Add Permissions in the Access Token" options.  Then use the save button to persist the new API.  Identifiers are used as the 'audience' setting when generating tokens via web calls against the AAA features of the Auth0 platform.

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
export AUTH0_REQUEST=$(printf '{"client_id": "%s", "client_secret": "%s", "audience":"http://api.cognizant-ai.dev/experimentsrv","grant_type":"client_credentials", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' "$AUTH0_CLIENT_ID" "$AUTH0_CLIENT_SECRET")
export AUTH0_TOKEN=$(curl -s --request POST --url https://cognizant-ai.auth0.com/oauth/token --header 'content-type: application/json' --data "$AUTH0_REQUEST" | jq -r '"\(.access_token)"')

curl --request POST --url 'https://cognizant-ai.auth0.com/oauth/token' --header 'content-type: application/json' --data '{ "client_id":"RjWuqwm1CM72iQ5G32aUjwIYx6vKTXBa", "client_secret": "MK_jpHrTcthM_HoNETnytYpqgMNS4e7zLMgp1_Wj2aePaPpubjN1UNKKCAfZlD_r", "audience": "http://api.cognizant-ai.dev/experimentsrv", "grant_type": "http://auth0.com/oauth/grant-type/password-realm", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "openid", "realm": "Username-Password-Authentication" }'
</b>
c.f. https://auth0.com/docs/quickstart/backend/golang/02-using#obtaining-an-access-token-for-testing.
</code></pre>

If you are using the test API and you are either running a kubectl port-forward or have a local instance of the postgres DB, you can do something like:

<pre><code><b> kubectl port-forward --namespace default svc/$PGRELEASE-postgresql 5432:5432 &
cd cmd/downstream
go run . --ip-port=":30008" &
cd ../..
cd cmd/experimentsrv
export AUTH0_DOMAIN=cognizant-ai.auth0.com
export AUTH0_CLIENT_ID=pL3iSUmOB7EPiXae4gPfuEasccV7PATs
export AUTH0_CLIENT_SECRET=KHSCFuFumudWGKISCYD79ZkwF2YFCiQYurhjik0x6OKYyOb7TkfGKJrHKXXADzqG
export AUTH0_REQUEST=$(printf '{"client_id": "%s", "client_secret": "%s", "audience":"http://api.cognizant-ai.dev/experimentsrv","grant_type":"client_credentials", "username": "karlmutch@gmail.com", "password": "Passw0rd!", "scope": "all:experiments", "realm": "Username-Password-Authentication" }' "$AUTH0_CLIENT_ID" "$AUTH0_CLIENT_SECRET")
export AUTH0_TOKEN=$(curl -s --request POST --url https://cognizant-ai.auth0.com/oauth/token --header 'content-type: application/json' --data "$AUTH0_REQUEST" | jq -r '"\(.access_token)"')
go test -v --dbaddr=localhost:5432 -ip-port="[::]:30007" -dbname=platform -downstream="[::]:30008"
</b></code></pre>

# Manually invoking and using services

Services used within the platform require that not only is the link integrity and security is maintained using mTLS but that an authorization block is also supplied to verify the user requesting a service.  The authorization can be supplied when using the gRPC command line tool using the metadata options.  First we retrieve a token using curl and then make a request against the service as follows:

<pre><code><b>grpc_cli call localhost:30001 dev.cognizant-ai.experiment.Service.Get "id: 'test'" --metadata authorization:"Bearer $AUTH0_TOKEN"
</b></code></pre>

The services used within the platform all support reflection when using gRPC.  To examine calls available for a server you should first identify the endpoint through which the ingress is being routed, for example:

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
package: dev.cognizant-ai.experiment;
service Service {
  rpc Create(dev.cognizant-ai.experiment.CreateRequest) returns (dev.cognizant-ai.experiment.CreateResponse) {}
  rpc Get(dev.cognizant-ai.experiment.GetRequest) returns (dev.cognizant-ai.experiment.GetResponse) {}
}
</code></pre>

To drill further into interfaces and examine the types being used within calls you can perform commands such as:

<pre><code><b>grpc_cli type $CLUSTER_INGRESS dev.cognizant-ai.experiment.CreateRequest -l</b>
message CreateRequest {
.dev.cognizant-ai.experiment.Experiment experiment = 1[json_name = "experiment"];
}
<b>grpc_cli type $CLUSTER_INGRESS dev.cognizant-ai.experiment.Experiment -l</b>
message Experiment {
string uid = 1[json_name = "uid"];
string name = 2[json_name = "name"];
string description = 3[json_name = "description"];
.google.protobuf.Timestamp created = 4[json_name = "created"];
map&lt;uint32, .dev.cognizant-ai.experiment.InputLayer&gt; inputLayers = 5[json_name = "inputLayers"];
map&lt;uint32, .dev.cognizant-ai.experiment.OutputLayer&gt; outputLayers = 6[json_name = "outputLayers"];
}
<b>grpc_cli type $CLUSTER_INGRESS dev.cognizant-ai.experiment.InputLayer -l</b>
message InputLayer {
enum Type {
	Unknown = 0;
	Enumeration = 1;
	Time = 2;
	Raw = 3;
}
string name = 1[json_name = "name"];
.dev.cognizant-ai.experiment.InputLayer.Type type = 2[json_name = "type"];
repeated string values = 3[json_name = "values"];
}
</code></pre>


# Shutting down a service, or cluster

<pre><code><b>kubectl delete -f experimentsrv.yaml
</b></code></pre>

<pre><code><b>kops delete cluster $CLUSTER_NAME --yes
</b></code></pre>
