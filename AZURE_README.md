# Using Azure ACS to reply kubernetes clusters

The document contains information as to how Microsoft supplied tools can be using to deploy kubernetes orchestration against the ACS container service.

Expect these instructions to change quickly as MSFT is investing in both standard kops and MSFT homegrown tooling as they now have several of the k8s contributors now working within the organization, much like AMZN.

# Account management

# Using ACS

acs-engine is the main CLI client for the Azure based service.  acs-engine is available on Windows, Linux, and OSX (darwin) from the github releases page for the open source by default, found at https://github.com/Azure/acs-engine/releases.  acs-engine is written in go and can be compiled fairly easily if you are feeling adventurous.

In general the ACS tool works much like other deployment tools and is used to generate templates for deployments.  acs-engine exposes the templates more visibly than kubernetes operations (kops) but is not much different.

AWS offers a CLI tool called, Azure CLI 2.0.  Instructions for installation for this tool can be found at https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest.a  This tool uses python.  The OSX version can be installed using Brew as follows:

```
brew update && brew install azure-cli
```

Having installed the CLI you can use the `az login` command to start a challenge response with the Azure OAuth system.  The web browser verification can be done on ANY system once the code has been printed on the console of the az login command.  The helps to prevent using a browser hosted inside a cloud provider which can be ludicrously slow.

To create a group into which the resources that will be used by the cluster will be gatherd you can use the output from the `az login` and the id json field to populate the subscription option.

```shell
$ az login
To sign in, use a web browser to open the page https://aka.ms/devicelogin and enter the code BHXXXXXN4 to authenticate.
[
  {
    "cloudName": "AzureCloud",
    "id": "a3xxxxxb-xxxx-xxxx-xxxx-39xxxxxxxx5f",
    "isDefault": true,
    "name": "Visual Studio Ultimate with MSDN",
    "state": "Enabled",
    "tenantId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "user": {
      "name": "karlmutch@hotmail.com",
      "type": "user"
    }
  }
]
$ az account set --subscription "a3xxxxxb-xxxx-xxxx-xxxx-39xxxxxxxx5f"
```

The next step is to define a basic cluster using a json file. Example ACS configurations for a basic K8 cluster can be seen at, https://github.com/Azure/acs-engine/tree/master/examples:

```shell
$ cat kubernetes.json
{
  "apiVersion": "vlabs",
  "properties": {
    "orchestratorProfile": {
      "orchestratorType": "Kubernetes"
    },
    "masterProfile": {
      "count": 1,
      "dnsPrefix": "",
      "vmSize": "Standard_D2_v2"
    },
    "agentPoolProfiles": [
      {
        "name": "agentpool1",
        "count": 3,
        "vmSize": "Standard_D2_v2",
        "availabilityProfile": "AvailabilitySet"
      }
    ],
    "linuxProfile": {
      "adminUsername": "azureuser",
      "ssh": {
        "publicKeys": [
          {
            "keyData": ""
          }
        ]
      }
    },
    "servicePrincipalProfile": {
      "clientId": "",
      "secret": ""
    }
  }
}

Next we will deploy the acs-engine python code and then begin deploying the clusteri, be sure to use a browser to activate the authentication:

```shell
# There is an OSX version available.  Check the github release page, https://github.com/Azure/acs-engine/releases
#
$ wget https://github.com/Azure/acs-engine/releases/download/v0.12.5/acs-engine-v0.12.5-linux-amd64.zip

$ acs-engine deploy --subscription-id "a3xxxxxb-xxxx-xxxx-xxxx-39xxxxxxxx5f" --dns-prefix k8-example --location westus2 --auto-suffix --api-model ./kubernetes.json
WARN[0000] To sign in, use a web browser to open the page https://aka.ms/devicelogin and enter the code BYE44GHA5 to authenticate.
INFO[0049] Registering subscription to resource provider. provider="Microsoft.Compute" subscription="a3e9d9fb-a359-4a0a-a5f8-3947a781ec5f"
INFO[0050] Registering subscription to resource provider. provider="Microsoft.Storage" subscription="a3e9d9fb-a359-4a0a-a5f8-3947a781ec5f"
INFO[0050] Registering subscription to resource provider. provider="Microsoft.Network" subscription="a3e9d9fb-a359-4a0a-a5f8-3947a781ec5f"
WARN[0051] apimodel: missing masterProfile.dnsPrefix will use "k8-example-5a78c32f"
WARN[0051] --resource-group was not specified. Using the DNS prefix from the apimodel as the resource group name: k8-example-5a78c32f
WARN[0054] apimodel: ServicePrincipalProfile was missing or empty, creating application...
WARN[0055] created application with applicationID (c6db57d6-0d94-41c7-9c81-58f343ef0eec) and servicePrincipalObjectID (3529896b-163c-4c3a-89ba-0dbe82762011).
WARN[0055] apimodel: ServicePrincipalProfile was empty, assigning role to application...
INFO[0075] Starting ARM Deployment (k8-example-5a78c32f-598563415). This will take some time...
```

Much in the same way kops generates files acs-engine will as well.  You will find these inside the _output directory by default.  As detailed in the deploy.md document within the Azure github repo configurations will be generated for all locations including the one into which the deploy was done so you will be need to be selectuve about the choosen configurations when using the kubernetes tools, for example:

```shell
$ export KUBECONFIG=_output/k8-example-5a78c32f/kubeconfig/kubeconfig.westus2.json
$ kubectl cluster-info
Kubernetes master is running at https://k8-example-5a78c32f.westus2.cloudapp.azure.com
Heapster is running at https://k8-example-5a78c32f.westus2.cloudapp.azure.com/api/v1/namespaces/kube-system/services/heapster/proxy
KubeDNS is running at https://k8-example-5a78c32f.westus2.cloudapp.azure.com/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
kubernetes-dashboard is running at https://k8-example-5a78c32f.westus2.cloudapp.azure.com/api/v1/namespaces/kube-system/services/kubernetes-dashboard/proxy
tiller-deploy is running at https://k8-example-5a78c32f.westus2.cloudapp.azure.com/api/v1/namespaces/kube-system/services/tiller-deploy:tiller/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

Now we have a running kubernetes deployment.  The cluster will be limited to production features as we have not enabled the rbac or api group extensions for tooling such as Istio.  This is a TODO.

More information can be found at https://github.com/Azure/acs-engine/blob/master/docs/kubernetes/deploy.md

