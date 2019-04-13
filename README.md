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

  grafana &lt;title&gt; &lt;dataSource&gt; &lt;resourcetype&gt; [&lt;maxdashboardresource&gt;] [&lt;maxcontinuation&gt;]
    Generate Grafana dashboard JSON files for given Azure resource type
</pre>

To use armclient with Azure CLI / shell.azure.com, generate an access token and populate the config file.

      az account get-access-token -o yaml

Example: sample-azure.yml

    credentials:
      environment: public
      accessToken:
      expiresOn:
      subscription:
      tenant:
      tokenType:


armclient will pull Grafana dashboard templates from the following repository.

https://github.com/robdyke/azure-grafana-dashboard-templates
