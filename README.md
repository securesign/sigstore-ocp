# sigstore-rhel-openshift

Sigstore helm-charts and build scripts opinionated for running on OCP and RHEL

`sigstore-rhel/scaffold` chart is an opinionated flavor of the upstream sigstore/scaffold chart located at [sigstore/helm-charts/charts/scaffold](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold). It extends the upstream chart with additional OpenShift specific functionality and provides opinionated values.

## Deploy scaffold chart in OpenShift

More information can be found by inspecting the upstream [sigstore/scaffold chart](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold).


### Create custom CA and key in fulcio-system namespace

Open [fulcio-create-CA script](./ROSA/fulcio-create-root-ca-openssl.sh) to check out the commands before running it.
The `openssl` commands are interactive.

*TODO: create external secret & template*

```bash
./fulcio-create-root-ca-openssl.sh
oc create ns fulcio-system
cd fulcio-root
oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=file_ca_key.pem --from-file=public=file_ca_pub.pem --from-file=cert=fulcio-root.pem  --from-literal=password=secure --dry-run=client -o yaml | oc apply -f-
```

Run something like the following to view your cluster's cluster-domain

```bash
oc get dnsrecords -o yaml -n openshift-ingress-operator | grep dnsName
```

### Edit the local [values.yaml](./helm/scaffold/base/values.yaml) at `helm/scaffold/base/values.yaml`


### Install RedHat SSO Operator (Keycloak)

```bash
oc apply -k helm/scaffold/overlays/keycloak-operator

# wait for this command to succeed before going on to be sure the Keycloak CRDs are registered
oc get keycloaks -A
```

### Apply manifests and deploy the helm charts

```bash
oc kustomize --enable-helm ./helm/scaffold/overlays/ocp | oc apply -f -
```

Watch as the stack rolls out.
You might follow the logs from the fulcio-server deployment in `-n fulcio-system`.

#### Add keycloak user and/or credentials

Check out the [user custom resource](https://github.com/redhat-et/sigstore-rhel/blob/main/helm/scaffold/overlays/keycloak/user.yaml)
for how to create a keycloak user. For testing, a user `jdoe@redhat.com` with password: `secure` is created.

You can access the keycloak route and login as the admin user to set credentials in the keycloak admin console. To get the keycloak admin credentials,
run `oc extract secret/credential-keycloak -n keycloak-system`. This will create an `ADMIN_PASSWORD` file with which to login. 

## Sign and/or verify artifacts!

Follow [this](https://github.com/redhat-et/sigstore-rhel/blob/main/sign-verify.md).
