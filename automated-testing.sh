#!/bin/bash

## Enablement, script meant for linux and OSX

## Self-signed cert check, fix in place for mac, thank you @sabre1041, need one for linux
### Deps: oc, curl
oc_console_route=$(oc get route console -n openshift-console | grep "console-openshift-console" | awk '{print $2}')
https_curl_response=$(curl -X GET https://$oc_console_route &> /dev/null)
https_curl_status=$(echo $?)

if [[ $https_curl_status == "60" ]]; then
    echo "self-signed cert for cluster"
    if [[ $(uname) == "Darwin" ]]; then
        ./scripts/configure-local-env.sh
        echo "certificate should be imported to OSX keychain, trying again"
        https_curl_response=$(curl -X GET https://$oc_console_route &> /dev/null)
        https_curl_status=$(echo $?)
        if [[ $https_curl_status != "0" ]]; then
            echo "Error: \`curl -X GET https://$oc_console_route produced status code $https_curl_status \`"
            exit 1
        fi
    else 
        echo  "currently no option scripted for linux, please add the certificate for your cluster to your trusted store and continue"
        exit 1
    fi
fi

source ./tas-env-variables.sh

## Binary testing
### Deps: jq, yq, podman, oc
echo "{}" > /tmp/tas-report.json
clientserver_namespace=$(cat charts/trusted-artifact-signer/values.yaml | yq .configs.clientserver.namespace)
clientserver_name=$(cat charts/trusted-artifact-signer/values.yaml | yq .configs.clientserver.name)
OS_FAMILY=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

#### Cosign
binary="cosign"
cosign_download_link=""

# Generate cosign entry in report
jq  -c  '.cosign = {}' -i /tmp/tas-report.json

# find correct download link
if [[ $OS_FAMILY == "linux" && $ARCH == "amd64" ]]; then
    cosign_download_link="https://$clientserver_namespace.$BASE_HOSTNAME/clients/$OS_FAMILY/$binary.gz"
else
    cosign_download_options=($(oc get -n $clientserver_namespace consoleclidownloads.console.openshift.io cosign -o json | jq ".spec.links[].href" | cut -d "\"" -f 2 ))
    for cosign_download_option in "${cosign_download_options[@]}"; do
        if [[ $cosign_download_option == "https://$clientserver_name-$clientserver_namespace.$BASE_HOSTNAME/clients/$OS_FAMILY/$binary-$ARCH.gz" ]]; then
            cosign_download_link=$cosign_download_option
        fi
    done
fi

# check cosign download link
if [[ -z $cosign_download_link ]]; then
    echo  "error getting cosign download link"
    jq  --arg OS "$OS_FAMILY" --arg ARCH "$ARCH" '.cosign.download = {"status": "failure", "os": $OS, "arch": $ARCH, "link": ""}' -i /tmp/tas-report.json
else
    echo "download matching OS: $OS_FAMILY and ARCH: $ARCH found:
    $cosign_download_link
    continuing... "
    jq  --arg OS "$OS_FAMILY" --arg ARCH "$ARCH" --arg LINK "$cosign_download_link" '.cosign.download = {"os": $OS, "arch": $ARCH, "link": $LINK}' -i /tmp/tas-report.json
fi

dir=$(pwd)

# idempotency

if [ -d "/tmp/cosign" ]; then
    rm -rf /tmp/cosign
fi

mkdir /tmp/cosign && cd /tmp/cosign

cosign_download=$(curl -sL $cosign_download_link -o /tmp/cosign/cosign-$OS_FAMILY-$ARCH.gz)
cosign_download_status=$(echo $?)
cosign_download_404=$(cat /tmp/cosign/cosign-$OS_FAMILY-$ARCH.gz | grep "<title>404 Not Found</title>")
gzip -d /tmp/cosign/cosign-$OS_FAMILY-$ARCH.gz --force
cosign_unizp_status=$(echo $?)

# checking download status of cosign
if [[ $cosign_download_status == 0 && -z $cosign_download_404 && $cosign_unizp_status == 0 ]]; then
    jq '.cosign.download.status = "success"' -i /tmp/tas-report.json
else
    jq '.cosign.download.status = "failure"' -i /tmp/tas-report.json
fi

chmod +x /tmp/cosign/cosign-$OS_FAMILY-$ARCH

podman pull registry.access.redhat.com/ubi9/s2i-base@sha256:d3838e6e26baa335556eb04f0af128602ddf7b57161d168b21ed6cf997281ddb
/tmp/cosign/cosign-$OS_FAMILY-$ARCH initialize --mirror=$TUF_URL --root=$TUF_URL/root.json
cosign_initialize_status=$(echo $?)
if [[ $cosign_initialize_status == 0 ]]; then
    jq '.cosign.initialize.status = "success"' -i /tmp/tas-report.json
else
    jq '.cosign.initialize.status = "failure"' -i /tmp/tas-report.json

fi

### Cosign keyless flow (no upload)
/tmp/cosign/cosign-$OS_FAMILY-$ARCH sign registry.access.redhat.com/ubi9/s2i-base@sha256:d3838e6e26baa335556eb04f0af128602ddf7b57161d168b21ed6cf997281ddb \
    --yes \
    --rekor-url=$REKOR_URL \
    --fulcio-url=$FULCIO_URL \
    --oidc-issuer=$OIDC_ISSUER_URL \
    --upload=false 
    # --output-file=/tmp/test-output # THIS DOES NOT WORK
    # --timestamp-server-url= \ # THIS HAS YET TO BE INCLUDED IN THE CHARTS
cosign_keyless_signing_status=$(echo $?)

if [[ $cosign_keyless_signing_status == 0 ]]; then
    jq --arg STATUS_CODE "$cosign_keyless_signing_status" '.cosign.sign.keyless = {"result": "success", "status_code": "$STATUS_CODE"}' -i /tmp/tas-report.json
else
    # ADD FAILURE CASE
fi

### Cosign generate-key-pair

export COSIGN_PASSWORD="tmp_cosign_password"
/tmp/cosign/cosign-$OS_FAMILY-$ARCH generate-key-pair --output-key-prefix tas-cosign
cosign_generate_key_statues=$(echo $?)
if [[ $cosign_generate_key_statues == 0 ]]; then
    jq --arg STATUS_CODE "$cosign_generate_key_statues" '.cosign.keyed = {"generate-key-pair": {"result": "success", "status_code": "$STATUS_CODE"}}' -i /tmp/tas-report.json
else
    # ADD FAILURE CASE
fi

## Cosign keyed flow
export COSIGN_PASSWORD="tmp_cosign_password"
tmp/cosign/cosign-$OS_FAMILY-$ARCH sign registry.access.redhat.com/ubi9/s2i-base@sha256:d3838e6e26baa335556eb04f0af128602ddf7b57161d168b21ed6cf997281ddb \
    --key=/tmp/cosign/tas-cosign.key \
    --rekor-url=$REKOR_URL \
    --upload=false 
cosign_keyed_signing_status=$(echo $?)


## COSIGN VERIFY --> this needs some where where we can push attestations
export COSIGN_PASSWORD="tmp_cosign_password"
tmp/cosign/cosign-$OS_FAMILY-$ARCH verify registry.access.redhat.com/ubi9/s2i-base@sha256:d3838e6e26baa335556eb04f0af128602ddf7b57161d168b21ed6cf997281ddb \
    --key=/tmp/cosign/tas-cosign.key \
    --rekor-url=$REKOR_URL
cosign_keyed_signing_status=$(echo $?)
