#!/usr/bin/env sh

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

# Function to check Realm Import Status status
check_rhbk_import_status() {
    local namespace="$1"
    local realm_import_name="$2"
    local attempts=0

    while [[ $attempts -lt $max_attempts ]]; do
        realm_import=$(oc get keycloakrealmimport "$realm_import_name" -n "$namespace" --ignore-not-found=true)
        if [ -n "$realm_import" ]; then
            realm_import_complete_status=$(oc get keycloakrealmimport "$realm_import_name" -n "$namespace" -o jsonpath='{.status.conditions[?(@.type=="Done")].status}')
            if [ "$realm_import_complete_status" == "True" ]; then
                echo "$realm_import_name Keycloak Realm Import in namespace $namespace is complete."
                return 0
            fi
        fi

        sleep $sleep_interval
        attempts=$((attempts + 1))
    done

    echo "Timed out. Keycloak Realm Import '$realm_import_name' completed within the specified time."
    return 1
}

# Install Red Hat Build of Keycloak
install_rhbk() {

    # Create keycloak-system namespace
    oc create namespace keycloak-system --dry-run=client -o yaml | oc apply -f-

    # Create PostgreSQL secret if it does not already exist
    if [[ ! -n $(oc get secret -n keycloak-system keycloak-postgresql --ignore-not-found=true) ]]; then
        oc create secret generic -n keycloak-system keycloak-postgresql --from-literal=username=keycloak --from-literal=password=$(echo $(date +%s%N) | sha256sum | head -c 20) --dry-run=client -o yaml | oc apply -f-
    fi

    # Install Keycloak Operator
    helm upgrade -i keycloak-operator charts/keycloak-operator  -n keycloak-system

    check_pod_status "keycloak-system" "rhbk-operator"
    # Check the return value from the function
    if [ $? -ne 0 ]; then
        echo "Pod status check failed. Exiting the script."
        exit 1
    fi

    # Install Keycloak
    helm upgrade -i keycloak charts/keycloak  -n keycloak-system --set keycloak.ingress.host=keycloak-keycloak-system.$common_name --set postgresql.secret.existingSecret=keycloak-postgresql

    check_pod_status "keycloak-system" "keycloak-postgresql-0"
    # Check the return value from the function
    if [ $? -ne 0 ]; then
        echo "Pod status check failed. Exiting the script."
        exit 1
    fi

    check_pod_status "keycloak-system" "keycloak-0"
    # Check the return value from the function
    if [ $? -ne 0 ]; then
        echo "Pod status check failed. Exiting the script."
        exit 1
    fi

    check_rhbk_import_status "keycloak-system" "trusted-artifact-signer"
}

common_name=apps.$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')

# Install Red Hat Build of Keycloak
install_rhbk

mkdir -p keys-cert
pushd keys-cert > /dev/null

read -p "Should this script generate the fulcio certificates? (Y/N): " generate_certs
if [ "$generate_certs" = "Y" ] || [ "$generate_certs" = "y" ]; then
    generate_certs=true
else
    generate_certs=false
fi

if [ "$generate_certs" = true ]; then
    # Prompt for the organizationName
    read -p "Enter the organization name for the certificate: " organization_name
    # Prompt for the email address
    read -p "Enter the email address for the certificate: " email_address
    # Prompt for the password
    read -s -p "Enter the password for the private key: " password

    openssl ecparam -genkey -name prime256v1 -noout -out unenc.key
    openssl ec -in unenc.key -out file_ca_key.pem -des3 -passout pass:"$password"
    openssl ec -in file_ca_key.pem -passin pass:"$password" -pubout -out file_ca_pub.pem
    openssl req -new -x509 -days 365 -key file_ca_key.pem -passin pass:"$password"  -out fulcio-root.pem -passout pass:"$password" -subj "/CN=$common_name/emailAddress=$email_address/O=$organization_name"
    rm unenc.key
fi
openssl ecparam -name prime256v1 -genkey -noout -out rekor_key.pem

segment_backup_job=$(oc get job -n trusted-artifact-signer-monitoring --ignore-not-found=true | tail -n 1 | awk '{print $1}')
if [[ -n $segment_backup_job ]]; then
    oc delete job $segment_backup_job -n trusted-artifact-signer-monitoring
fi

oc new-project trusted-artifact-signer-monitoring > /dev/null 2>&1

pull_secret_exists=$(oc get secret pull-secret -n trusted-artifact-signer-monitoring --ignore-not-found=true)
if [[ -n $pull_secret_exists ]]; then
    read -p "Secret \"pull-secret\" in namespace \"trusted-artifact-signer-monitoring\" already exists. Overwrite it (Y/N)?: " -n1 overwrite_pull_secret
    echo ""
    if [[ $overwrite_pull_secret == "Y" || $overwrite_pull_secret == 'y' ]]; then
        read -p "Please enter the absolute path to the pull-secret.json file:
" pull_secret_path
        file_exists=$(ls $pull_secret_path 2>/dev/null)
        if [[ -n $file_exists ]]; then
            oc create secret generic pull-secret -n trusted-artifact-signer-monitoring --from-file=$pull_secret_path --dry-run=client -o yaml | oc replace -f -
        else
            echo "pull secret was not found based on the path provided: $pull_secret_path"
            exit 0
        fi
    elif [[ $overwrite_pull_secret == "N" || $overwrite_pull_secret == 'n' ]]; then
        echo "Skipping overwriting pull-secret..."
    else 
        echo "Bad input. Skipping this step, using existing pull-secret"
    fi
else
    read -p "Please enter the absolute path to the pull-secret.json file:
" pull_secret_path
    oc create secret generic pull-secret -n trusted-artifact-signer-monitoring --from-file=$pull_secret_path
fi

popd > /dev/null

oc create ns fulcio-system
if [ "$generate_certs" = true ]; then
    oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./keys-cert/file_ca_key.pem --from-file=public=./keys-cert/file_ca_pub.pem --from-file=cert=./keys-cert/fulcio-root.pem --from-literal=password="$password" --dry-run=client -o yaml | oc apply -f-
fi

oc create ns rekor-system
oc -n rekor-system create secret generic rekor-private-key --from-file=private=./keys-cert/rekor_key.pem --dry-run=client -o yaml | oc apply -f-

# TODO: uncomment to install from helm repository, install from the local repo checkout for now
#helm repo add trusted-artifact-signer https://repo-securesign-helm.apps.open-svc-sts.k1wl.p1.openshiftapps.com/helm-charts
#helm repo update
#OPENSHIFT_APPS_SUBDOMAIN=$common_name envsubst < examples/values-sigstore-openshift.yaml | helm install --debug trusted-artifact-signer trusted-artifact-signer/trusted-artifact-signer -n trusted-artifact-signer --create-namespace --values -

if [ "$generate_certs" = true ]; then
    OPENSHIFT_APPS_SUBDOMAIN=$common_name envsubst < examples/values-sigstore-openshift.yaml | helm upgrade -i trusted-artifact-signer --debug charts/trusted-artifact-signer  -n trusted-artifact-signer --create-namespace --values -
else
    OPENSHIFT_APPS_SUBDOMAIN=$common_name envsubst < examples/values-sigstore-openshift.yaml | helm upgrade -i trusted-artifact-signer --debug charts/trusted-artifact-signer  -n trusted-artifact-signer --create-namespace --values - --set scaffold.fulcio.createcerts.enabled=true  
fi

echo "\nTo initialize the environment variables, run 'source ./tas-env-variables.sh' from the terminal."
