# Enabling Grafana Monitoring for Trusted-Artifact-Signer

This guide provides the commands to deploy both the Grafana operator
and a Grafana instance in OpenShift. It also adds a Prometheus Datasource
and configures a dashboard for monitoring Sigstore components.

Prerequisites
1. Make sure you have the [oc command-line tool](https://docs.openshift.com/container-platform/4.12/cli_reference/openshift_cli/getting-started-cli.html) installed.
2. Ensure you are logged into your OpenShift cluster.
3. Ensure workload monitoring is enabled in your OpenShift cluster. If necessary, either add the line `enableUserWorkload: true` to an already existing `configmap/cluster-monitoring-config` in `-n openshift-monitoring` _or_ create the configmap as below. For more information, refer to [OpenShift documentation](https://docs.openshift.com/container-platform/4.13/monitoring/enabling-monitoring-for-user-defined-projects.html).

```yaml
oc create --save-config -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-monitoring-config
  namespace: openshift-monitoring
data:
  config.yaml: |
    enableUserWorkload: true
EOF
```

Note: This guide assumes you are using OpenShift version 4.12 or greater.

## Step 1: Installing the operator

This installs the Grafana operator into the `sigstore-monitoring` namespace.
```bash
oc apply -k grafana/operator
```

Tip: Verify the installation by running `oc get pods -n sigstore-monitoring`.

## Step 2: Creating a grafana instance

This creates a Grafana instance for the operator. Make sure to allow some time for the Grafana operator to install.

```bash
oc apply -k grafana/instance
```

## Step 3: Configuring grafana resources

Apply the necessary tokens and role bindings to the service account `grafana-serviceaccount` in the `sigstore-monitoring` namespace

```bash
oc apply -k grafana/resources
```
## Step 4: Retrieveing secrets

Retrieve all necessary secrets from the OpenShift cluster and apply them to the `datasource.yaml` file found at `grafana/dashboards/datasource.yaml`.

```bash
export BEARER_TOKEN=$(oc -n sigstore-monitoring get secrets grafana-sa-token -o=jsonpath="{.data.token}" | base64 -d)
export MYSQL_USER=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-user}" | base64 -d)
export MYSQL_PASSWORD=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-password}" | base64 -d)
export MYSQL_DATABASE=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-database}" | base64 -d)
```

## Step 5: Creating datasources & dashboards

Finally, the datasources and dashboards can be created.

```bash
envsubst < grafana/dashboards/datasource.yaml | oc apply -f -
oc apply -f grafana/dashboards/dashboard.yaml
```

## Step 6: Access the UI

To find the Grafana UI route, execute:

```bash
oc -n sigstore-monitoring get routes
```

Or, navigate to Networking -> Routes in the `sigstore-monitoring` namespace through the OpenShift cluster UI, the default username and password is `sigstore-rh`, please ensure to update this to something more secure. Once logged in, navigate to the dashboard by going to Dashboards -> Browse -> sigstore-monitoring -> Sigstore Monitoring.