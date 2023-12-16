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

wipe_dir_if_exists() {
    dir_path="$1"
    if [[ -d $dir_path ]]; then
        rm -rf $dir_path
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

# # TSA is not currently working, add in later
# timestamp_authority_enbaled=$(cat $git_root/charts/trusted-artifact-signer/values.yaml | yq .scaffold.tsa.enabled)
# timestamp_authority_server_url=""
# if [[ "$timestamp_authority_enbaled" == "true" ]]; then
#     tsa_url=$(oc get ingress tsa-server -n tsa-system -o json | jq '.spec.rules[].host' | cut -d "\"" -f 2)
#     timestamp_authority_server_url=" --timestamp-server-url=$tsa_url"
# fi

if [[ $(cat $REPORT_FILE_ABS_PATH | jq '.cosign.download.result' | cut -d "\"" -f 2 ) == "success" ]]; then
    log_y_sub_step "2.2" "setup for cosign unit tests"
    
    echo "making binary executable ..."
    chmod +x /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH

    echo "Building base image ..."
    podman build -q $git_root/testing -f Dockerfile.test -t localhost/tas-infra-test

    log_y_sub_step "2.3" "cosign unit tests"

    ### COSIGN INITIALIZE

    log_z_sub_step "2.3.1" "cosign initialize"

    wipe_dir_if_exists "$HOME/.sigstore"
    wipe_file_if_exists "/tmp/tas/cosign/tmp-stdout.log" && wipe_file_if_exists "/tmp/tas/cosign/tmp-stderr.log"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH initialize --mirror=$TUF_URL --root=$TUF_URL/root.json 2>/tmp/tas/cosign/tmp-stderr.log 1>/tmp/tas/cosign/tmp-stdout.log
    cosign_initialize_status=$(echo $?)
    if [[ $cosign_initialize_status == 0 ]]; then
        sed  '1d' /tmp/tas/cosign/tmp-stdout.log > /tmp/tas/cosign/tmp-two-stdout.log && mv /tmp/tas/cosign/tmp-two-stdout.log /tmp/tas/cosign/tmp-stdout.log
        root_status_b64=$(cat /tmp/tas/cosign/tmp-stdout.log | jq . | base64)
        jq --arg ROOT_STATUS_B64 "$root_status_b64" '.cosign.initialize = {"result": "success", "root_status_b64": $ROOT_STATUS_B64}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        stderr=$(cat /tmp/tas/cosign/tmp-stderr.log)
        jq --arg STDERR "$stderr" '.cosign.initialize = {"result": "failure", "stderr": $STDERR}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    ### COSIGN KEYLESS

    log_z_sub_step "2.3.2" "cosign sign keyless"

    cosign_keyless_signing_image_tag="ttl.sh/tas-cosign-keyless-sign-$run_sha:1h"
    podman tag localhost/tas-infra-test $cosign_keyless_signing_image_tag
    podman push -q $cosign_keyless_signing_image_tag

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

    ## COSGIN VERIFY KEYLESS (offline mode)

    ### Cosign generate-key-pair

    log_z_sub_step "2.3.3" "cosign generate key pair"

    wipe_file_if_exists "/tmp/tas/cosign/generate-key-pair.log"
    cd /tmp/tas/cosign
    export COSIGN_PASSWORD="tmp_cosign_password"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH generate-key-pair --output-key-prefix tas-cosign &> /tmp/tas/cosign/generate-key-pair.log
    cosign_generate_key_status=$(echo $?)
    cd $git_root/testing

    generate_key_pair_check_string=$'Private key written to tas-cosign.key\nPublic key written to tas-cosign.pub'
    
    if [[ $cosign_generate_key_status == 0 && "$(cat /tmp/tas/cosign/generate-key-pair.log)" == $(echo "$generate_key_pair_check_string") ]]; then
        jq --arg STATUS_CODE "$cosign_generate_key_status" '.cosign.generate_key_pair = {"result": "success"}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        jq --arg LOG "$(cat /tmp/tas/cosign/generate-key-pair.log)" --arg STATUS_CODE "$cosign_generate_key_status" '.cosign.generate_key_pair = {"result": "failure", "status_code": "$STATUS_CODE", "log": $LOG}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    ### Cosign keyed flow

    log_z_sub_step "2.3.4" "cosign sign keyed"

    cosign_keyed_signing_image_tag="ttl.sh/tas-cosign-keyed-sign-$run_sha:1h"
    podman tag localhost/tas-infra-test $cosign_keyed_signing_image_tag
    podman push -q $cosign_keyed_signing_image_tag

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


    log_z_sub_step "2.3.5" "cosign verify (keyed)"

    wipe_file_if_exists "/tmp/tas/cosign/tmp-stdout.log" &&  wipe_file_if_exists "/tmp/tas/cosign/tmp-stderr.log"
    export COSIGN_PASSWORD="tmp_cosign_password"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH --output-file /tmp/tas/cosign/tmp-stdout.log verify $cosign_keyed_signing_image_tag \
        --key=/tmp/tas/cosign/tas-cosign.pub \
        --rekor-url=$REKOR_URL \
        -o json 2> /tmp/tas/cosign/tmp-stderr.log
    cosign_keyed_verify_status=$(echo $?)

    if [[ $cosign_keyed_verify_status  == 0 ]]; then
        verification=$(cat /tmp/tas/cosign/tmp-stdout.log | jq '.[0]' | base64)
        docker_reference=$(echo $verification | base64 -d | jq '.critical.identity."docker-reference"' | cut -d "\"" -f 2)
        docker_manifest_digest=$(echo $verification | base64 -d | jq '.critical.image."docker-manifest-digest"' | cut -d "\"" -f 2)
        tlog_index=$(echo $verification | base64 -d | jq '.optional.Bundle.Payload.logIndex')
        payload_body=$(echo $verification | base64 -d | jq '.optional.Bundle.Payload.body' | cut -d "\"" -f 2)
        rekor_uuid=$(echo $payload_body | base64 -d | jq '.spec.data.hash.value' | cut -d "\"" -f 2)
        signature=$(echo $payload_body | base64 -d | jq '.spec.signature.content' | cut -d "\"" -f 2)
        public_key=$(echo $payload_body | base64 -d | jq '.spec.signature.publicKey.content' | cut -d "\"" -f 2 | base64 -d)
        jq --arg PUBLIC_KEY "$public_key" --arg SIG "$signature" --arg SHA "$docker_manifest_digest" --arg IMAGE "$docker_reference" --arg TLOG_INDEX "$tlog_index" --arg REKOR_UUID "$rekor_uuid" '.cosign.verify.keyed = {"result": "success", "image": $IMAGE, "sha": $SHA, "b64_signature": $SIG, "publicKey": $PUBLIC_KEY, "tlog_index": $TLOG_INDEX, "critical.image.docker-manifest-digest": "value", }' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else
        jq --arg STATUS_CODE "$cosign_keyed_verify_status" --arg IMAGE "$cosign_keyed_signing_image_tag" '.cosign.verify.keyed = {"result": "failure", "image": $IMAGE, "status_code":  $STATUS_CODE}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    exit 0

    log_z_sub_step "2.3.6" "cosign clean (keyed)"

    # log_z_sub_step "2.3.7" "cosign attest image (keyed)" ## IN PROGRESS

    # export COSIGN_PASSWORD="tmp_cosign_password"
    # /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH attest $cosign_keyed_signing_image_tag \
    #     --key=/tmp/tas/cosign/tas-cosign.key

    # exit 0
    
    log_z_sub_step "2.3.8" "cosign generate and compare against cosign sign (keyed)"


    wipe_file_if_exists "/tmp/tas/cosign/cosign-generate.log"
    cosign_generate_image_tag="ttl.sh/tas-cosign-generate-$run_sha:1h"
    podman tag localhost/tas-infra-test $cosign_generate_image_tag
    podman push -q $cosign_generate_image_tag
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH generate $cosign_generate_image_tag &> /tmp/tas/cosign/cosign-generate.log
    cosign_generate_status=$(echo $?)
    if [[ $cosign_generate_status == 0 ]]; then
        payload=$(cat /tmp/tas/cosign/cosign-generate.log)
        wipe_file_if_exists "/tmp/tas/cosign/cosign-generate-sign.log" && wipe_file_if_exists "/tmp/tas/cosign/cosign-generate-sign-payload.log"
        /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH sign $cosign_generate_image_tag \
            --fulcio-url=$FULCIO_URL \
            --oidc-issuer=$OIDC_ISSUER_URL \
            --upload=false \
            --yes \
            --output-payload /tmp/tas/cosign/cosign-generate-sign-payload.log &> /dev/null # we dont do anything with this error log, just hits case 3
        cosign_generate_sign_status=$(echo  $?)
        if [[ $cosign_generate_sign_status == 0 ]]; then
            sign_payload=$(cat /tmp/tas/cosign/cosign-generate-sign-payload.log)
            if [[ $payload == $sign_payload ]]; then
                payload_b64=$(cat /tmp/tas/cosign/cosign-generate.log | jq . | base64)
                jq --arg PAYLOAD_B64 "$payload_b64" --arg IMAGE "$cosign_generate_image_tag" --arg STATUS_CODE "$cosign_generate_sign_status" '.cosign.generate = {"result": "success", "image": $IMAGE, "sign": {"result": "success", "match": true, "payload_b64": $PAYLOAD_B64}}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
                jq_update_file
            else
                jq --arg IMAGE "$cosign_generate_image_tag" --arg STATUS_CODE "$cosign_generate_sign_status" '.cosign.generate = {"result": "success", "image": $IMAGE, "sign": {"result": "success", "match": false}}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
                jq_update_file
            fi
        else 

            jq --arg IMAGE "$cosign_generate_image_tag" --arg STATUS_CODE "$cosign_generate_sign_status" '.cosign.generate = {"result": "success", "image": $IMAGE, "sign": {"result": "failure", "status_code": $STATUS_CODE}}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
            jq_update_file
        fi
    else 
        jq --arg IMAGE "$cosign_generate_image_tag" --arg STATUS_CODE "$cosign_generate_status" '.cosign.generate = {"result": "failure", "image": $IMAGE, "status_code": $STATUS_CODE}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi
    
    log_z_sub_step "2.3.9" "cosign sign-blob"

    cp $git_root/testing/Dockerfile.test /tmp/tas/cosign
    ## COSIGN SIGN BLOB (keyed, Dockerfile)
    wipe_file_if_exists "/tmp/tas/cosign/cosign-sign-blob.log"
    export COSIGN_PASSWORD="tmp_cosign_password"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH sign-blob /tmp/tas/cosign/Dockerfile.test \
        --key=/tmp/tas/cosign/tas-cosign.key \
        --rekor-url=$REKOR_URL \
        --bundle=/tmp/tas/cosign/Dockerfile.test.bundle \
        --yes &> /tmp/tas/cosign/cosign-sign-blob.log
    cosign_sign_blob_status=$(echo $?)

    if [[ $cosign_sign_blob_status == 0 ]]; then
        tlog_index=$(cat /tmp/tas/cosign/cosign-sign-blob.log | grep "tlog entry created with index: ")
        tlog_index=${tlog_index:31:(( ${#tlog_index} - 31))}

        jq --arg TLOG_INDEX "$tlog_index" '.cosign.sign.blob = {"result": "success", "tlog_index": $TLOG_INDEX}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else 
        jq --arg STATUS_CODE "$cosign_sign_blob_status" '.cosign.sign.blob = {"result": "failure", "status_code": $STATUS_CODE}' $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    ### Cosign VERIFY-BLOB

    log_z_sub_step "2.3.10" "cosign verify-blob"
    wipe_file_if_exists "/tmp/tas/cosign/cosign-verify-blob.log"
    /tmp/tas/cosign/cosign-$OS_FAMILY-$ARCH verify-blob /tmp/tas/cosign/Dockerfile.test --key=/tmp/tas/cosign/tas-cosign.pub --bundle=/tmp/tas/cosign/Dockerfile.test.bundle &> /tmp/tas/cosign/cosign-verify-blob.log
    cosign_verify_blob_status=$(echo $?)

    if [[ $cosign_verify_blob_status == 0 && $(cat /tmp/tas/cosign/cosign-verify-blob.log) == "Verified OK" ]]; then
        jq '.cosign.verify_blob = {"result":"success"}'  $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    else 
        jq --arg STATUS_CODE "$cosign_verify_blob_status" '.cosign.verify_blob = {"result": "failure", "status_code": $STATUS_CODE}'  $REPORT_FILE_ABS_PATH > $REPORT_TMP_FILE_ABS_PATH
        jq_update_file
    fi

    # exit 0


    ## COSIGN PUBLIC KEY


    ## COSIGN ATTEST

    ## COSIGN ATTEST BLOB

    ## COSIGN VERIFY 

    ## COSIGN COPY

    ## COSIGN Clean 

    exit 0

    # ## COSIGN VERIFY --> this needs some where where we can push attestations
    # export COSIGN_PASSWORD="tmp_cosign_password"
    # tmp/cosign/cosign-$OS_FAMILY-$ARCH verify registry.access.redhat.com/ubi9/s2i-base@sha256:d3838e6e26baa335556eb04f0af128602ddf7b57161d168b21ed6cf997281ddb \
    #     --key=/tmp/tas/cosign/tas-cosign.key \
    #     --rekor-url=$REKOR_URL
    # cosign_keyed_signing_status=$(echo $?)


fi

exit 0

# echo "finished, generated report: "
# cat cat /tmp/tas/tas-report.json | jq
