# armclient
Azure Resource Manager client in Go

<pre>
Usage: armclient [&lt;flags&gt;] <command> [&lt;args&gt; ...]

Flags:
  --help   Show context-sensitive help (also try --help-long and --help-man).
  --config.file="sample-azure.yml"
           Azure configuration file
  --debug  Debug flag

Commands:
  help [&lt;command&gt;...]
    Show help.

  get &lt;url&gt;
    Perform GET &lt;url&gt; against Azure Resource Manager API

  resources [&lt;maxcontinuation&gt;]
    Print out the Azure resources that exist on this subscription
</pre>

To use armclient, you must first create a service principal which has Reader permission to access your Azure subscription.
https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-create-service-principal-portal

Example: sample-azure.yml
<pre>
Credentials:
  environment: public
  subscription_id: &lt;subscriptionId&gt;
  client_id: &lt;clientId&gt;
  client_secret: &lt;clientSecret&gt;
  tenant_id: &lt;tenantId&gt;
</pre
