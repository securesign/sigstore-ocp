## Quick Start with Keycloak OIDC

No-Fail steps to get a working sigstore stack with OpenShift

1. Install RHSSO Operator and deploy Sigstore Keycloak

```shell
oc apply --kustomize keycloak/operator
# wait until the keycloak API is ready, check w/ non-erroring 'oc get keycloaks'
oc apply --kustomize keycloak/resources
# wait until keycloak-system pods are healthy/running
```

2. Create the fulcio signing keys & root cert, and then a secret in the fulcio-system namespace. Replace <PASSWORD> in the `oc` command to match the password for decrypting the signing key. The script creates the keys in `./keys-cert` folder.

```shell
oc create ns fulcio-system
./fulcio-create-root-ca-openssl.sh
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./keys-cert/file_ca_key.pem --from-file=public=./keys-cert/file_ca_pub.pem --from-file=cert=./keys-cert/fulcio-root.pem  --from-literal=password=<PASSWORD> --dry-run=client -o yaml | oc apply -f-
```

3. Create the rekor signing keys, and then a secret in the rekor-system namespace. The script creates the key in `./keys-cert` folder.

```shell
oc create ns rekor-system
./rekor-create-signer-key.sh
oc -n rekor-system create secret generic rekor-private-key --from-file=private=./keys-cert/rekor_key.pem --dry-run=client -o yaml | oc apply -f-
```

3.  Run the following:

```shell
OPENSHIFT_APPS_SUBDOMAIN=apps.$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }') envsubst <  examples/values-sigstore-openshift.yaml | helm upgrade -i trusted-artifact-signer --debug charts/trusted-artifact-signer -n trusted-artifact-signer --create-namespace --values -
```

A good way to tell if things are progressing well is to watch `oc get jobs -A` and when the tuf-system job is complete,
things should be ready.

Once complete, move to the [Sign & Verify document](./sign-verify.md) to test the Sigstore stack. 
