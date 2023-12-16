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
### Deps: jq, yq, 
clientserver_namespace=$(cat charts/trusted-artifact-signer/values.yaml | yq .configs.clientserver.namespace)
OS_FAMILY=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

#### Cosign
binary="cosign"
cosign_download_link=""
if [[ $OS_FAMILY == "linux" && $ARCH == "amd64" ]]; then
    cosign_download_link="https://$clientserver_namespace.$BASE_HOSTNAME/clients/$OS_FAMILY/$binary.gz"
else
    cosign_download_options=($(oc get -n $clientserver_namespace consoleclidownloads.console.openshift.io cosign -o json | jq ".spec.links[].href"))
    for cosign_download_option in "${cosign_download_options[@]}"; do
        if [[ $cosign_download_option == "https://$clientserver_namespace.$BASE_HOSTNAME/clients/$OS_FAMILY/$binary-$ARCH.gz" ]]; then
            cosign_download_link=$cosign_download_option
        fi
    done
fi

if [[ -z $cosign_download_link ]]; then
    echo  "error getting cosign download link"
    exit 1 #THIS IS A TEMPORARY PLACEHOLDER
fi

cosign_download=$(curl -sL $cosign_download_link -o /tmp/cosign-$OS_FAMIL-$ARCH.gz)
not_found_html_string="<head>
<title>404 Not Found</title>
</head>"
if [[ $(cat $cosign_download | grep "$not_found_html_string") ]]


# 2 options for testing cosign, could test by downloading the binary from console-cli-downloads, or we could use the cosign pod with kubectl exec
# 1. download the binary from cluster


# for binary in "${!binaries[@]}"; do
#     oc get consoleclidownloads.console.openshift.io cosign -n openshift-console -o json | jq ".spec.links[].href"

# cosign_options=$(oc get consoleclidownloads.console.openshift.io cosign -n openshift-console -o json | jq ".spec.links")
# 2. kubectl exec (in progress)
    # cosign_pod=$(oc get pods -n cosign | tail -n 1 | awk '{print $1}')1
    # kubectl exec -n cosign $cosign_pod 
    # oc rsh $cosign_pod

# cosign --help