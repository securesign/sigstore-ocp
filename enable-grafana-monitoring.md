# Enabling Grafana Monitoring for SecureSign

This guide provides step-by-step instructions to enable Grafana monitoring in a openshift cluster using the grafana operator.

Prerequisites
1. Make sure you have the oc command-line tool installed.
2. Ensure you are logged into your OpenShift cluster.

Note: This guide assumes you are using OpenShift version 4.12.

## Step 1: Installing the operator

This installs the Grafana operator into the `grafana-operator` namespace.
```bash
oc apply -k grafana/operator
```

Tip: Verify the installation by running `oc get pods -n grafana-operator`.

## Step 2: Creating a grafana instance

This creates a Grafana instance for the operator. Make sure to allow some time for the Grafana operator to install.

```bash
oc apply -k grafana/instance
```

## Step 3: Configuring grafana resources

Apply the necessary tokens and role bindings to the service account `grafana-serviceaccount` in the `grafana-operator` namespace

```bash
oc apply -k grafana/resources
```
## Step 4: Retrieveing secrets

Retrieve all necessary secrets from the OpenShift cluster and apply them to the `datasource.yaml` file found at `grafana/dashboards/datasource.yaml`.

```bash
export BEARER_TOKEN=$(oc -n grafana-operator get secrets grafana-sa-token -o=jsonpath="{.data.token}" | base64 -d)
export MYSQL_USER=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-user}" | base64 -d)
export MYSQL_PASSWORD=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-password}" | base64 -d)
export MYSQL_DATABASE=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-database}" | base64 -d)
```

```bash
sed -i "s/\${BEARER_TOKEN}/${BEARER_TOKEN}/g" grafana/dashboards/datasource.yaml
sed -i "s/\${MYSQL_USER}/${MYSQL_USER}/g" grafana/dashboards/datasource.yaml
sed -i "s/\${MYSQL_PASSWORD}/${MYSQL_PASSWORD}/g" grafana/dashboards/datasource.yaml
sed -i "s/\${MYSQL_DATABASE}/${MYSQL_DATABASE}/g" grafana/dashboards/datasource.yaml
```
## Step 5: Creating datasources & dashboards

Finally, the datasources and dashboards can be created.

```bash
oc apply -k grafana/dashboards
```
## Step 6: Access the UI

To find the Grafana UI route, execute:

```bash
oc -n grafana-operator get routes
```

Or, navigate to Networking -> Routes in the `grafana-operator` namespace through the OpenShift cluster UI, use `rhel` as the username & password. Once logged in, navigate to the dashboard by going to Dashboards -> Browse -> grafana-operator -> Sigstore Monitoring.