#!/bin/bash

# This assumes you are currently running from the context of the namespace where your securesign is created
# Run `oc project <securesign namespace>` to ensure you are working within the correct context

# Initialize Variables
export BASE_HOSTNAME=$(kubectl get cm -n openshift-config-managed  console-public -o go-template="{{ .data.consoleURL }}" | sed 's@https://@@; s/^[^.]*\.//')
export KEYCLOAK_NAMESPACE="${KEYCLOAK_NAMESPACE:=keycloak-system}"

export KEYCLOAK_CLIENT_ID="${KEYCLOAK_CLIENT_ID:=trusted-artifact-signer}"
export KEYCLOAK_REALM="${KEYCLOAK_REALM:=trusted-artifact-signer}"
export KEYCLOAK_HOSTNAME="${KEYCLOAK_HOSTNAME:=https://$(kubectl get keycloak -n ${KEYCLOAK_NAMESPACE} -o jsonpath='{.items[*].spec.hostname.hostname}')}"
export OIDC_ISSUER_URL="${OIDC_ISSUER_URL:=${KEYCLOAK_HOSTNAME}/realms/${KEYCLOAK_REALM}}"

if [[ `kubectl api-resources -o name | grep securesigns.rhtas.redhat.com` ]]; then
    CURRENT_NAMESPACE="${RHTAS_NAMESPACE:=$(kubectl config view --minify -o jsonpath='{..namespace}')}"

    # Ensure Securesign resource has been created
    SECURESIGN_NAME=$(kubectl get securesign -n ${CURRENT_NAMESPACE} -o name | head -1)
    if [[ -z "${SECURESIGN_NAME}" ]]; then
        echo "Error: Securesign resource not created in namespace \"${CURRENT_NAMESPACE}\""
        return
    fi

    TUF_URL=$(kubectl get tuf -o jsonpath='{.items[0].status.url}' -n ${CURRENT_NAMESPACE})
    REKOR_URL=$(kubectl get rekor -o jsonpath='{.items[0].status.url}' -n ${CURRENT_NAMESPACE})
    FULCIO_URL=$(kubectl get fulcio -o jsonpath='{.items[0].status.url}' -n ${CURRENT_NAMESPACE})
else
    # Set Values for Helm Chart deployment
    TUF_URL="${TUF_URL:=https://tuf.${BASE_HOSTNAME}}"
    FULCIO_URL="${FULCIO_URL:=https://fulcio.${BASE_HOSTNAME}}"
    REKOR_URL="${REKOR_URL:=https://rekor.${BASE_HOSTNAME}}"
fi

# Common Variables
export COSIGN_FULCIO_URL="${FULCIO_URL}"
export COSIGN_REKOR_URL="${REKOR_URL}"
export COSIGN_MIRROR="${TUF_URL}"
export COSIGN_ROOT="${TUF_URL}/root.json"
export COSIGN_OIDC_ISSUER="${OIDC_ISSUER_URL}"
export COSIGN_OIDC_CLIENT_ID="${KEYCLOAK_CLIENT_ID}"
export COSIGN_CERTIFICATE_OIDC_ISSUER="${OIDC_ISSUER_URL}"
export COSIGN_YES="true"
export SIGSTORE_FULCIO_URL="${FULCIO_URL}"
export SIGSTORE_OIDC_CLIENT_ID="${KEYCLOAK_CLIENT_ID}"
export SIGSTORE_OIDC_ISSUER="${OIDC_ISSUER_URL}"
export SIGSTORE_REKOR_URL="${REKOR_URL}"
export REKOR_REKOR_SERVER="${REKOR_URL}"
