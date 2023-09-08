# Enabling Grafana Monitoring for this Helm Chart

This guide provides a step-by-step instruction to enable Grafana monitoring in a openshift cluster using Helm charts. Make sure you have Helm and oc command-line tool installed, and you are logged into the cluster.

Note: This guide assumes you are using Helm version 3.11 and OpenShift version 4.12.

## Step 1: Configuring Access Token

Before enabling Grafana monitoring, retrieve an access token from a service account located in the openshift-user-workload-monitoring namespace.

```
export ACC_TOKEN=$(oc -n openshift-user-workload-monitoring get secrets -o name | grep 'prometheus-user-workload-token-' | xargs -I {} oc -n openshift-user-workload-monitoring get {} -o=jsonpath="{.data.token}" | base64 -d)
```

This command retrieves the correct access token and sets it to the ACC_TOKEN environmental variable.You may want to `echo $ACC_TOKEN` before proceeding to ensure the variable is set if its not the cluster may not be fully provisioned yet.

## Step 2: Deploying the Helm Chart with Grafana

Run the following command to deploy the Helm charts with Grafana enabled:

```
helm upgrade -i scaffolding --debug . -n sigstore --create-namespace -f examples/values-ez.yaml --set grafana.enabled=true --set grafana.accToken=$ACC_TOKEN
```

## Step 3: Accessing the Grafana UI

To find the Grafana UI route, execute:

```
oc -n sigstore get routes
```

Or navigate to Networking -> Routes in the sigstore namespace through the OpenShift cluster UI.

Retrieve the admin password for Grafana:
```
oc -n sigstore get secrets scaffolding-grafana -o=jsonpath="{.data.admin-password}" | base64 -d
```

Use `admin` as the username and the retrieved password to log in.

Once logged in, navigate to the dashboard by going to Dashboards -> General -> Sigstore Monitoring.

