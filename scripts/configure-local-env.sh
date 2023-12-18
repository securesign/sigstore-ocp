#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

CERT_FILE_PREFIX="tmp-cert"
KUBECTL_TOOL="kubectl"
CERT_ACTION="add"
TEMPDIR=$(mktemp -d -t rhtas-certs -p /tmp)
trap 'rm -r "${TEMPDIR}"' EXIT

function display_help {
  echo "./$(basename "$0") [ -a | --apps-domain APPS_DOMAIN ] [ -gr | --gitops-namespace NAMESPACE ] [ -h | --help ] [ -hr | --helm-revision REVISION ] [ -hr | --helm-repository REPOSITORY ] [ -t | --tool TOOL ]

Deployment of Argo CD Applications to support the managment of SPIFFE/SPIRE on OpenShift

Where:
  -d  | --delete            Delete certificates from OSX Keychain
  -h  | --help              Display this help text
  -t  | --tool              Tool for communicating with OpenShift cluster. Defaults to '${KUBECTL_TOOL}'

"
}


for i in "${@}"
do
case $i in
    -d | --delete )
    CERT_ACTION="delete"
    shift
    ;;
    -t | --tool )
    KUBECTL_TOOL="${1}"
    shift
    ;;
    -h | --help )
    display_help
    exit 0
    ;;
    -*) echo >&2 "Invalid option: " "${@}"
    exit 1
    ;;
esac
done

# Check if split is installed
command -v split >/dev/null 2>&1 || { echo >&2 "split is required but not installed.  Aborting."; exit 1; }

# Check if kubectl or compatible is installed
command -v ${KUBECTL_TOOL} >/dev/null 2>&1 || { echo >&2 "kubectl tool is required but not installed.  Aborting."; exit 1; } 

# Grab the Kube Root Certificates
${KUBECTL_TOOL} get -n default cm kube-root-ca.crt -o jsonpath='{.data.ca\.crt}' > ${TEMPDIR}/ca.crt

# Split Certificates from bundle
split -p "-----BEGIN CERTIFICATE-----" "${TEMPDIR}/ca.crt" ${TEMPDIR}/cert-

# Find the ingress-operator certificte and add/remove it to/from the OSX keystore
for f in ${TEMPDIR}/cert-*; do
    COMMON_NAME=$(openssl x509 -subject -noout -nameopt multiline -in $f | grep commonName | awk '{ print $3 }')
    if echo "${COMMON_NAME}" | grep -q "^ingress-operator"; then
      if [ "${CERT_ACTION}" == "delete" ]; then
        security find-certificate -c "${COMMON_NAME}" -a -Z | sudo awk '/SHA-1/{system("security delete-certificate -Z "$NF)}'
        echo "'${COMMON_NAME}' removed from keychain"
      else
        sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "${f}"
        echo "'${COMMON_NAME}' added to keychain"
      fi
    fi
done
