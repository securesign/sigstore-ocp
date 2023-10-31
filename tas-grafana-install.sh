#!/usr/bin/env sh
# Script for creating a Grafana dashboard for the Trusted Artifact Signer service
#

# Define the maximum number of attempts and the sleep interval (in seconds)
max_attempts=30
sleep_interval=10
# Function to check pod status
check_pod_status() {
    local namespace="$1"
    local pod_name_prefix="$2"
    local attempts=0

    while [[ $attempts -lt $max_attempts ]]; do
        pod_name=$(oc get pod -n "$namespace" | grep "$pod_name_prefix" | grep "Running" | awk '{print $1}')
        if [ -n "$pod_name" ]; then
            pod_status=$(oc get pod -n "$namespace" "$pod_name" -o jsonpath='{.status.phase}')
            if [ "$pod_status" == "Running" ]; then
                echo "$pod_name is up and running in namespace $namespace."
                return 0
            else
                echo "$pod_name is in state: $pod_status. Retrying in $sleep_interval seconds..."
            fi
        else
            echo "No pods with the prefix '$pod_name_prefix' found in namespace $namespace. Retrying in $sleep_interval seconds..."
        fi

        sleep $sleep_interval
        attempts=$((attempts + 1))
    done

    echo "Timed out. No pods with the prefix '$pod_name_prefix' reached the 'Running' state within the specified time."
    return 1
}

# Ensure workload monitoring is enabled in the OpenShift cluster.
enable_workload_monitoring="true"
oc get configmap -n openshift-monitoring > /tmp/g1
if grep -q "cluster-monitoring-config" /tmp/g1; then
    oc get configmap cluster-monitoring-config -n openshift-monitoring -o json| jq -r '.data."config.yaml"' > /tmp/g1
    if grep -q "enableUserWorkload" /tmp/g1; then
        extracted_text=$(grep "enableUserWorkload" /tmp/g1 | cut -f2 -d ":")
	trim_extracted_text=$(echo "$extracted_text" | sed -e 's/^[[:space:]]*//')
        if [ "$trim_extracted_text" = "true" ]; then
            echo "Workload monitoring is already enabled"
            enable_workload_monitoring="false"
        fi
    fi
fi

if [ "$enable_workload_monitoring" = "true" ]; then
    echo "setup cluster monitoring"
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
fi

# Install SSO Operator and Keycloak service
install_tas_grafana() {
    oc apply -k grafana/operator
    check_pod_status "sigstore-monitoring" "grafana-operator-controller-manager"
    # Check the return value from the function
    if [ $? -ne 0 ]; then
        echo "Pod status check failed. Exiting the script."
        exit 1
    fi

    oc apply -k grafana/instance
    check_pod_status "sigstore-monitoring" "grafana-deployment"
    # Check the return value from the function
    if [ $? -ne 0 ]; then
        echo "Pod status check failed. Exiting the script."
        exit 1
    fi

    # Create Grafana secret token
    oc apply -k grafana/resources
    sleep 15

    # Setup environment variables
    export BEARER_TOKEN=$(oc -n sigstore-monitoring get secrets grafana-sa-token -o=jsonpath="{.data.token}" | base64 -d)
    export MYSQL_USER=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-user}" | base64 -d)
    export MYSQL_PASSWORD=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-password}" | base64 -d)
    export MYSQL_DATABASE=$(oc -n trillian-system get secrets trillian-mysql -o=jsonpath="{.data.mysql-database}" | base64 -d)

    # Modify datasource and create dashboard
    envsubst < grafana/dashboards/datasource.yaml | oc apply -f -
    echo "Wait for restart of the grafana deployment pod ..."
    sleep 30
    oc apply -f grafana/dashboards/dashboard.yaml
    echo "Wait for creation of the dashboard ..."
    sleep 15

    # Get route to connect to the dashboard
    oc -n sigstore-monitoring get routes
}

install_tas_grafana

