#!/bin/bash

check_namespace_exists() {
    namespace=$1
    local sleep_interval=3
    local attempts=0
    local max_attempts=10
    while [[ $attempts -lt $max_attempts ]]; do
        namespace_exists=$( kubectl get namespace $namepsace --ignore-not-found=true)
        if [[ -z $namespace_exists ]]; then
            sleep $sleep_interval
            attempts=$((attempts + 1))
        else
            echo "namespace \`$1\` exists. Proceeding ..."
            return 0
        fi
    done
    echo "Timeout, namespace \`$1\` does not exist."
    exit 1
}

check_secret_exists() {
    namespace=$1
    secret_name=$2
    local sleep_interval=3
    local attempts=0
    local max_attempts=10
    while [[ $attempts -lt $max_attempts ]]; do
        secret_exists=$(kubectl get secret $secret_name -n $namespace --ignore-not-found=true)
        if [[ -z $secret_exists ]]; then
            sleep $sleep_interval
            attempts=$((attempts + 1))
        else
            echo "secret \`$secret_name\` exists."
            return 0
        fi
    done
    echo "Secret \`$secret_name\` does not exist."
    return 1
}


mkdir -p keys-cert
pushd keys-cert > /dev/null

organization_name=$(cat /tmp/tas-initializer-input/organization_name)
email_address=$(cat /tmp/tas-initializer-input/email_address)
password=$(cat /tmp/tas-initializer-input/password)
common_name=$(kubectl get dns cluster -o jsonpath='{ .spec.baseDomain }')

openssl ecparam -genkey -name prime256v1 -noout -out unenc.key
openssl ec -in unenc.key -out file_ca_key.pem -des3 -passout pass:"$password"
openssl ec -in file_ca_key.pem -passin pass:"$password" -pubout -out file_ca_pub.pem
openssl req -new -x509 -days 365 -key file_ca_key.pem -passin pass:"$password"  -out fulcio-root.pem -passout pass:"$password" -subj "/CN=$common_name/emailAddress=$email_address/O=$organization_name"
openssl ecparam -name prime256v1 -genkey -noout -out rekor_key.pem

check_namespace_exists "fulcio-system"
fulcio_secret_exists=$(check_secret_exists "fulcio-system" "fulcio-secret-rh")
if [[ "$fulcio_secret_exists" == "1" ]]; then
    kubectl create secret generic fulcio-secret-rh -n fulcio-system --from-file=/tmp/keys-cert/fulcio-root.pem
else 
    echo "Replacing secret \`fulcio-secret-rh\` in namespace \`fulcio-system\`."
    kubectl create secret generic fulcio-secret-rh -n fulcio-system --from-file=/tmp/keys-cert/fulcio-root.pem -o yaml --dry-run=client | kubectl replace -f -
fi

check_namespace_exists "rekor-system"
rekor_secret_exists=$(check_secret_exists "rekor-system" "rekor-private-key")
if [[ "$rekor_secret_exists" == "1" ]]; then
    kubectl create secret generic rekor-private-key -n rekor-system --from-file=/tmp/keys-cert/rekor_key.pem
else 
    echo "Replacing secret \`rekor-private-key\` in namespace \`rekor-system\`."
    kubectl create secret generic rekor-private-key -n rekor-system --from-file=/tmp/keys-cert/rekor_key.pem -o yaml --dry-run=client  | kubectl replace -f -
fi
