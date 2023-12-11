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

    # customizing grafana creds
    read -p "Would you like to use the default credentials for grafana? (Y/N): " -n1 modify_creds
    echo ""
    declare admin_user
    declare admin_pass
    if [[ $modify_creds == "Y" || $modify_creds == "y" ]]; then
        read -p "Enter what you would like the admin username to be (default = \"sigstore-rh\"): " admin_user
        read -s -p "Enter the password you would like for the admin account: " admin_pass
        if [[ -z $admin_user ]]; then
            admin_user="sigstore-rh"
        fi
        if [[ -z $admin_pass ]]; then
            admin_pass="sigstore-rh"
        fi
    elif [[ $modify_creds == "N" || $modify_creds == "n" ]]; then
        admin_user="sigstore-rh"
        admin_pass="sigstore-rh"
    else
        echo "incorrect input, please try again."
        exit 1
    fi

    check_path=$(ls ./grafana/instance/instance.yaml 2>/dev/null)
    declare path
    if [[ -z $check_path ]]; then
        path="./sigstore-grafana-instance.yaml"
    else 
        path="./grafana/instance/instance.yaml"
    fi

    echo "apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: sigstore-monitoring
  namespace: sigstore-monitoring
spec:
# TODO: this hard-coded image version is necessary until the currently available version
# of grafana operator from OperatorHub (v4.7.1) pulls in a later version of grafana
baseImage: 'docker.io/grafana/grafana:10.1.2'
ingress:
  enabled: true
config:
  auth:
  disable_signout_menu: true
  auth.anonymous:
  enabled: true
  log:
  level: warn
  mode: console
  security:
  admin_password: \"$admin_pass\"
  admin_user: \"$admin_user\"
dashboardLabelSelector:
  - matchExpressions:
      - key: app
        operator: In
        values:
          - grafana" > "${path}"
    oc apply -f "${path}" -n sigstore-monitoring
    grafana_secret_exists=$(oc get secret oc create secret generic grafana-admin-creds -n sigstore-monitoring --ignore-not-found)
    if [[ -n $grafana_secret_exists ]]; then
    echo "Already found an existing \"grafana-admin-creds\" Secret in the \"sigstore-monitoring\" namespace."
    oc create secret generic grafana-admin-creds -n sigstore-monitoring \
        --from-literal=admin_user=$admin_user \
        --from-literal=admin_password=$admin_pass \
        --dry-run=client -o yaml | oc replace -f -
    else
        oc create secret generic grafana-admin-creds -n sigstore-monitoring \
            --from-literal=admin_user=$admin_user \
            --from-literal=admin_password=$admin_pass 
    fi

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

    # setting the password 
    oc -n sigstore-monitoring create secret generic grafana-admin-credentials --from-literal=webhook-secret-key=
}

install_tas_grafana

