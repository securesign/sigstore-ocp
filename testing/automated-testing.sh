#!/bin/bash

## Enablement, script meant for linux and OSX

# 0. Helper functions

log_step() {
    local step_number="$1"
    local line_length=40

    # Calculate the number of spaces needed for centering
    local spaces_before=$(( (line_length - ${#step_number}) / 2 ))
    local spaces_after=$(( line_length - ${#step_number} - spaces_before ))

    # Print the line of # characters above the centered step number
    printf "%*s\n" "$line_length" | tr ' ' '#'

    # Print the centered step number
    printf "%*s%s%*s\n" "$spaces_before" "" "$step_number" "$spaces_after" ""

    # Print the line of # characters below the centered step number
    printf "%*s\n" "$line_length" | tr ' ' '#'
}

log_y_sub_step() {
    sub_step_number="$1"
    sub_step_name="$2"
    echo "====> $sub_step_number $sub_step_name"
}

log_z_sub_step() {
    sub_step_number="$1"
    sub_step_name="$2"
    echo "=========> $sub_step_number $sub_step_name"
}

wipe_file_if_exists(){
    file_path="$1"
    if [[ -e $file_path ]]; then
        rm -f $file_path
    fi
}

git_root=$(git rev-parse --show-toplevel)

# 1. SETUP SECTION
# ------------------------------------------------------------------------------------------------
## Self-signed cert check, fix in place for mac, thank you @sabre1041, need one for linux
### Deps: oc, curl

log_step "1. Setup"
log_y_sub_step "1.1" "Self-signed cert check and remediation"

oc_console_route=$(oc get route console -n openshift-console | grep "console-openshift-console" | awk '{print $2}')
https_curl_response=$(curl -X GET https://$oc_console_route &> /dev/null)
https_curl_status=$(echo $?)

if [[ $https_curl_status == "60" ]]; then
    echo "self-signed cert for cluster"
    if [[ $(uname) == "Darwin" ]]; then
        $git_root/scripts/configure-local-env.sh
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

log_y_sub_step "1.2" "source env vars"
source $git_root/tas-env-variables.sh &> /dev/null


# 2. BINARY DOWNLOADS AND TESTING
# ----------------------------------------------------------------------------------------------------------
## Binary testing
### Deps: jq, yq, podman, oc, openssl (just for generating unique sha), file, git

log_step "2. Binaries"

if [[ -d "/tmp/tas" ]]; then
    rm -rf /tmp/tas
fi

mkdir /tmp/tas


REPORT_FILE_ABS_PATH=/tmp/tas/tas-report.json
REPORT_TMP_FILE_ABS_PATH=/tmp/tas/tmp-tas-report.json

# idempotency
if [ -e "/tmp/tas/tas-report.json" ]; then
    rm -f /tmp/tas/tas-report.json
fi

if [ -e "/tmp/tas/tmp-tas-report.json" ]; then
    rm -f /tmp/tas/tmp-tas-report.json
fi

random_string=$(LC_ALL=C openssl rand -base64 12 | tr -dc 'a-zA-Z0-9' | head -c 10)
run_sha=$(echo -n "$random_string" | sha256sum | awk '{print $1}')

jq -n '{"run_sha": $ARGS.named["run_sha"], "cosign": {}, "gitsign": {}, "rekor-server": {}, "rekor-cli": {}}' \
    --arg run_sha "$run_sha"  > $REPORT_FILE_ABS_PATH 


jq_update_file() {
     if [[ $? != 0 ]]; then
        echo "jq could not parse file" 
        exit $?
    fi
    mv $REPORT_TMP_FILE_ABS_PATH $REPORT_FILE_ABS_PATH
}



clientserver_namespace=$(cat $git_root/charts/trusted-artifact-signer/values.yaml | yq .configs.clientserver.namespace)
clientserver_name=$(cat $git_root/charts/trusted-artifact-signer/values.yaml | yq .configs.clientserver.name)
OS_FAMILY=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

#### Cosign
log_y_sub_step "2.1" "cosign-cli"

binary="cosign"
cosign_download_link=""

# find correct download link
if [[ $OS_FAMILY == "linux" && $ARCH == "amd64" ]]; then
    cosign_download_link="https://$clientserver_name-$clientserver_namespace.$BASE_HOSTNAME/clients/$OS_FAMILY/$binary.gz"
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
    jq  --arg OS "$OS_FAMILY" --arg ARCH "$ARCH" '.cosign.download = {"status": "failure", "os": $OS, "arch": $ARCH}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
    jq_update_file
else
    # # IF I ADD LOGLEVELS PUT THIS BACK IN
#     echo "download matching OS: $OS_FAMILY and ARCH: $ARCH found:  
#     $cosign_download_link
# continuing... "
    jq  --arg OS "$OS_FAMILY" --arg ARCH "$ARCH" --arg LINK "$cosign_download_link" '.cosign.download = {"os": $OS, "arch": $ARCH, "link": $LINK}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
    jq_update_file
fi

# idempotency
if [ -d "/tmp/tas/cosign" ]; then
    rm -rf /tmp/tas/cosign
fi

mkdir /tmp/tas/cosign

# Cosign Download
cosign_download=$(curl -sL $cosign_download_link -o /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH.gz)
cosign_download_status=$(echo $?)
cosign_download_is_html=$(file /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH.gz | grep "HTML document text")
gzip -d /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH.gz --force
cosign_unizp_status=$(echo $?)

# checking download status of cosign
if [[ $cosign_download_status == 0 && -z $cosign_download_is_html && $cosign_unizp_status == 0 ]]; then
    jq '.cosign.download.result = "success"'  $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
    jq_update_file
else
    # add Additional Identifiers to figure what the error was
    jq '.cosign.download.result = "failure"' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
    jq_update_file
fi

if [[ $(cat $REPORT_FILE_ABS_PATH | jq '.cosign.download.result' | cut -d "\"" -f 2 ) == "success" ]]; then
    log_y_sub_step "2.2" "setup for cosign unit tests"
    
    echo "making binary executable ..."
    chmod +x /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH

    echo "Building base image ..."
    podman build $git_root/testing -f Dockerfile.test -t localhost/tas-infra-test &> /dev/null

    log_y_sub_step "2.3" "cosign unit tests"
    log_z_sub_step "2.3.1" "cosign initialize"

    wipe_file_if_exists "/tmp/tas/cosign/tmp-stdout.log" && wipe_file_if_exists "/tmp/tas/cosign/tmp-stderr.log"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH initialize --mirror=$TUF_URL --root=$TUF_URL/root.json 2>/tmp/tas/cosign/tmp-stderr.log  1>/tmp/tas/cosign/tmp-stdout.log
    cosign_initialize_status=$(echo $?)
    if [[ $cosign_initialize_status == 0 ]]; then
        stdout=$(cat /tmp/tas/cosign/tmp-stdout.log)
        jq '.cosign.initialize = {"result": "success"}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        stderr=$(cat /tmp/tas/cosign/tmp-stderr.log)
        jq --arg STDERR "$stderr" '.cosign.initialize = {"result": "failure", "stderr": $STDERR}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    log_z_sub_step "2.3.2" "cosign sign keyless"

    cosign_keyless_signing_image_tag="ttl.sh/tas-cosign-keyless-sign-$run_sha:1h"
    podman tag localhost/tas-infra-test $cosign_keyless_signing_image_tag
    podman push $cosign_keyless_signing_image_tag &> /dev/null

    ### Cosign keyless flow (no upload)
    wipe_file_if_exists "/tmp/tas/cosign/keyless-sign.log"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH sign $cosign_keyless_signing_image_tag \
        --fulcio-url=$FULCIO_URL \
        --oidc-issuer=$OIDC_ISSUER_URL \
        --rekor-url=$REKOR_URL \
        --upload=true \
        --yes &> /tmp/tas/cosign/keyless-sign.log
    cosign_keyless_signing_status=$(echo $?)

    ################ DEV NOTES ##################

    # COSIGN SIGN OPTIONS THAT DID NOT WORK:
        # --oidc-client-secret-file=jdoe-client-secret.txt \ #THIS DOES NOT WORK
        # --output-file=/tmp/test-output # THIS DOES NOT WORK
    # COSIGN SIGN OPTIONS TO IMPLEMENT LATER:
        # --timestamp-server-url= \ # THIS HAS YET TO BE INCLUDED IN THE CHARTS
    # Issue:
        # I tried redirecting above cosign sign command output like so: `2>/tmp/tas/cosign/tmp-stderr.log  1>/tmp/tas/cosign/tmp-stdout.log`, however everything ended up on stderr, even when it succeeded

    ############### END DEV NOTES ###############

    if [[ $cosign_keyless_signing_status == 0 ]]; then
        tlog_index=$(cat /tmp/tas/cosign/keyless-sign.log | grep "tlog entry created with index: ")
        tlog_index=${tlog_index:31:(( ${#tlog_index} - 31))}
        jq --arg IMAGE "$cosign_keyless_signing_image_tag" --arg TLOG_INDEX "$tlog_index" '.cosign.sign.keyless = {"result": "success", "image": $IMAGE, "tlog_index": $TLOG_INDEX}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        jq --arg STATUS_CODE "$cosign_keyless_signing_status" '.cosign.sign.keyless = {"result": "failure", "status_code": $STATUS_CODE}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    log_z_sub_step "2.3.3" "cosign generate key pair"

    ### Cosign generate-key-pair

    wipe_file_if_exists "/tmp/tas/cosign/generate-key-pair.log"
    cd /tmp/tas/cosign
    export COSIGN_PASSWORD="tmp_cosign_password"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH generate-key-pair --output-key-prefix tas-cosign &> /tmp/tas/cosign/generate-key-pair.log
    cosign_generate_key_statues=$(echo $?)
    cd $git_root/testing
    generate_key_pair_check_string="Private key written to tas-cosign.key
Public key written to tas-cosign.pub"
    if [[ $cosign_generate_key_statues == 0 && "$(cat /tmp/tas/cosign/generate-key-pair.log)" == $generate_key_pair_check_string ]]; then
        jq --arg STATUS_CODE "$cosign_generate_key_statues" '.cosign.generate_key_pair = {"result": "success"}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        jq --arg LOG "$(cat /tmp/tas/cosign/generate-key-pair.log)" --arg STATUS_CODE "$cosign_generate_key_statues" '.cosign.generate_key_pair = {"result": "failure", "status_code": "$STATUS_CODE", "log": $LOG}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    log_z_sub_step "2.3.4" "cosign sign keyed"

    # ## Cosign keyed flow
    cosign_keyed_signing_image_tag="ttl.sh/tas-cosign-keyed-sign-$run_sha:1h"
    # cosign_keyed_signing_image_tag="$quay_repo:cosign-keyed-sign-$run_sha"
    podman tag localhost/tas-infra-test $cosign_keyed_signing_image_tag
    podman push $cosign_keyed_signing_image_tag &> /dev/null

    wipe_file_if_exists "/tmp/tas/cosign/keyed-sign.log"
    export COSIGN_PASSWORD="tmp_cosign_password"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH sign $cosign_keyed_signing_image_tag \
        --key=/tmp/tas/cosign/tas-cosign.key \
        --rekor-url=$REKOR_URL \
        --upload=true \
        --yes &> /tmp/tas/cosign/keyed-sign.log
    cosign_keyed_signing_status=$(echo $?)

    if [[ $cosign_keyed_signing_status == 0 ]]; then
        tlog_index=$(cat /tmp/tas/cosign/keyed-sign.log | grep "tlog entry created with index: ")
        tlog_index=${tlog_index:31:(( ${#tlog_index} - 31))}
        jq --arg IMAGE "$cosign_keyed_signing_image_tag" --arg TLOG_INDEX "$tlog_index" '.cosign.sign.keyed = {"result": "success", "image": $IMAGE, "tlog_index": $TLOG_INDEX}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        jq --arg IMAGE "$cosign_keyed_signing_image_tag" --arg STATUS_CODE "$cosign_keyed_signing_status" '.cosign.sign.keyed = {"result": "failure", "image": $IMAGE, "status_code": $STATUS_CODE}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi
    
    ## COSIGN SIGN BLOB (keyless)

    ## COSIGN PUBLIC KEY

    ## COSIGN DOCKERFILE

    ## COSIGN ATTEST

    ## COSIGN ATTEST BLOB

    ## COSIGN VERIFY 

    ## COSIGN VERIFY BLOB

    ## COSIGN COPY

    ## COSIGN Clean 



    # ## COSIGN VERIFY --> this needs some where where we can push attestations
    # export COSIGN_PASSWORD="tmp_cosign_password"
    # tmp/cosign/cosign-$OS_FAMILY-$ARCH verify registry.access.redhat.com/ubi9/s2i-base@sha256:d3838e6e26baa335556eb04f0af128602ddf7b57161d168b21ed6cf997281ddb \
    #     --key=/tmp/tas/cosign/tas-cosign.key \
    #     --rekor-url=$REKOR_URL
    # cosign_keyed_signing_status=$(echo $?)
fi


