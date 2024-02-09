#!/bin/bash

export BASE_HOSTNAME=apps.$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')
echo "base hostname = $BASE_HOSTNAME"


export KEYCLOAK_REALM=sigstore
export KEYCLOAK_URL=https://keycloak-keycloak-system.$BASE_HOSTNAME
export TUF_URL=https://tuf.$BASE_HOSTNAME
export COSIGN_FULCIO_URL=https://fulcio.$BASE_HOSTNAME
export COSIGN_REKOR_URL=https://rekor.$BASE_HOSTNAME
export COSIGN_MIRROR=$TUF_URL
export COSIGN_ROOT=$TUF_URL/root.json
export COSIGN_OIDC_ISSUER=$KEYCLOAK_URL/auth/realms/$KEYCLOAK_REALM
export COSIGN_CERTIFICATE_OIDC_ISSUER=$COSIGN_OIDC_ISSUER
export COSIGN_YES="true"

# Gitsign/Sigstore Variables
export SIGSTORE_FULCIO_URL=$COSIGN_FULCIO_URL
export SIGSTORE_OIDC_ISSUER=$COSIGN_OIDC_ISSUER
export SIGSTORE_REKOR_URL=$COSIGN_REKOR_URL

# Rekor CLI Variables
export REKOR_REKOR_SERVER=$COSIGN_REKOR_URL

