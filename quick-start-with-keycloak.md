## No fail quick start in 5 steps

No-Fail steps to get a working sigstore stack with OpenShift
(hopefully)

0. Obtain the base_domain with:

```shell
oc get dns/cluster -o jsonpath='{ .spec.baseDomain }' && echo
```

1. Install RHSSO Operator and deploy Sigstore Keycloak

```shell
oc apply --kustomize keycloak/operator # wait until the keycloak API is ready, check w/ non-erroring 'oc get keycloaks'
oc apply --kustomize keycloak/resources # wait until keycloak-system pods are healthy/running
```

2. Create the keys & root cert, and then a secret in the fulcio-system and rekor-system namespace. Replace <PASSWORD> in the `oc` command to match the password for decrypting the signing key.

```shell
oc create ns fulcio-system
oc create ns rekor-system
./fulcio-create-root-ca-openssl.sh
./rekor-create-signer-key.sh
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./keys-cert/file_ca_key.pem --from-file=public=./keys-cert/file_ca_pub.pem --from-file=cert=./keys-cert/fulcio-root.pem  --from-literal=password=<PASSWORD> --dry-run=client -o yaml | oc apply -f-
oc -n rekor-system create secret generic rekor-private-key --from-file=private=./keys-cert/rekor_key.pem --dry-run=client -o yaml | oc apply -f-
```

3. Substitute base domain found above for `<OPENSHIFT_BASE_DOMAIN> in [examples/values-sigstore-openshift.yaml](./examples/values-sigstore-openshift.yaml)

4.  Run the following:

```shell
helm upgrade -i scaffolding --debug . -n sigstore --create-namespace -f examples/values-sigstore-openshift.yaml
```

A good way to tell if things are progressing well is to watch `oc get jobs -A` and when the tuf-system job is complete,
things should be ready.

Once complete, move to the [Sign & Verify document](./sign-verify.md) to test the Sigstore stack. 
