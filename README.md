# sigstore-openshift

Wrapper chart to streamline the scaffolding of Sigstore within an OpenShift environment.

## Overview

This wrapper chart builds on top of the [Scaffold](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold) chart from the Sigstore project to both simplify and satisfy the requirements for deployment within an OpenShift

If you have already read this document and want a quick no-fail path to installing a Sigstore stack, follow the section [quick start](#no-fail-quick-start-in-5-steps)

The chart enhances the scaffold chart by taking care of the following:

* Provision Namespaces
* Configure `RoleBindings` to enable access to the `anyuid` SecurityContextConstraint
* Inject Fulcio root and Rekor signing keys

### Chart configurations

The following sections describe how to specifically configure certain features of the chart:

#### Fulcio root key injection

Utilize the following commands and configurations to inject Fulcio root secret:

First, generate a root key.
Open [fulcio-create-CA script](./fulcio-create-root-ca-openssl.sh) to check out the commands before running it.
The `openssl` commands are interactive.

```shell
./fulcio-create-root-ca-openssl.sh
```

Add the following to an overriding Values file injecting the public key, private key, and password used for the private key:

```yaml
configs:
  fulcio:
    server:
      secret:
        name: "fulcio-secret-rh"
        password: "<password>"
        public_key_file: "keys-cert/file_ca_pub.pem"
        private_key_file: "keys-cert/file_ca_key.pem"
        root_cert_file: "keys-cert/fulcio-root.pem"
```

#### Rekor Signer Key

Open [rekor create signer script](./rekor-create-signer-key.sh) to check out the commands before running it.
Generate a signer key:

```shell
./rekor-create-signer-key.sh
```

Add the following to override the values file injecting the signer key:

```yaml
configs:
  rekor:
    signer:
      secret:
        name: rekor-private-key
        private_key_file: "keys-cert/rekor_key.pem"
```

NOTE: The name of the generated secret, `rekor-private-key` can be customized. Ensure the naming is consistent throughout each of the customization options

#### Scaffolding customization

Similar to any Helm dependency, values from the upstream `scaffold` chart can be customized by embedding the properties within the `scaffold` property similar to the following:

```yaml
scaffold:
  fulcio:
    namespace:
      name: fulcio-system
      create: false
...
```

### Sample Implementation

A Helm values file is available in the [examples](examples) directory named [values-sigstore-openshift.yaml](examples/values-sigstore-openshift.yaml) that provides a baseline to work off of. It can be customized based on an individual target environment. 

#### Prerequisites

The following must be satisfied prior to deploying the sample implementation:

* Keycloak/RHSSO Deploy (or compatible OpenID endpoint)
    * A dedicated _Realm_ (Recommended)
    * A _Client_ representing the Sigstore integration
        * Valid _Redirect URI's_. A value of `*` can be used for testing
    * 1 or more `Users`
        * Email Verified

The RHSSO Operator and necessary Keycloak resources can be deployed in OpenShift with the following:

```shell
oc apply --kustomize keycloak/operator

# wait for this command to succeed before going on to be sure the Keycloak CRDs are registered
oc get keycloaks -A

oc apply --kustomize keycloak/resources
# wait for keycloak-system pods to be running before proceeding
```

#### Update the values file

Perform the following modifications to the customized sample files to curate the deployment of the chart:

1. Update all occurrences of `<OPENSHIFT_BASE_DOMAIN>` with the value from the following command:

```shell
oc get dnsrecords -o yaml -n openshift-ingress-operator -o jsonpath='{ .spec.baseDomain }'
```

2. Update all occurrences of `<KEYCLOAK_HOSTNAME>` with the hostname of Keycloak/RHSSO and `<REALM>` with the name of the Keycloak/RHSSO realm. If an alternate OIDC provider other than Keycloak/RHSSO was used, update the values to align to this implementation. 

3. Update all occurrences of `<CLIENT_ID>` with the name of the OIDC client

4. Perform any additional customizations as desired

### Installing the Chart

When logged in as an elevated OpenShift user, execute the following to install the chart referencing the customized values file:

```shell
helm upgrade -i scaffolding --debug . -n sigstore --create-namespace -f <values_file>
```
#### Add keycloak user and/or credentials

Check out the [user custom resource](https://github.com/redhat-et/sigstore-rhel/blob/main/helm/scaffold/overlays/keycloak/user.yaml)
for how to create a keycloak user. For testing, a user `jdoe@redhat.com` with password: `secure` is created.

You can access the keycloak route and login as the admin user to set credentials in the keycloak admin console. To get the keycloak admin credentials,
run `oc extract secret/credential-keycloak -n keycloak-system`. This will create an `ADMIN_PASSWORD` file with which to login. 

### Sign and/or verify artifacts!

Follow [this](https://github.com/redhat-et/sigstore-rhel/blob/main/sign-verify.md).

## No fail quick start in 5 steps

No-Fail steps to get a working sigstore stack with a fresh OpenShift cluster
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

2. Create the keys & root cert. This will populate a directory `./keys-cert`

```shell
./fulcio-create-root-ca-openssl.sh  #interactive, you'll enter "apps.<base_domain>" for the hostname, and enter the same password for all keys
./rekor-create-signer-key.sh
```

3. Substitute `base_domain` from above (rosa.xxxxx.com) in 5 places in [examples/values-ez.yaml](./examples/values-ez.yaml), like so in VIM :) 

```shell 
:%s/rosa.*.com/rosa.p9esn-qgkm3-wkk.o7au.p3.openshiftapps.com/g
```

* *If you intend to use the cosign pod to sign an image, ensure you configure the BASE_HOSTNAME value in the values.yaml file.*
```
cosign:
  BASE_HOSTNAME: apps.<base_domain>
```

4.  Run the following:

```shell
helm upgrade -i scaffolding --debug . -n sigstore --create-namespace -f examples/values-ez.yaml
```

A good way to tell if things are progressing well is to watch `oc get jobs -A` and when the tuf-system job is complete, things should be ready.
Once complete, move to the [Sign & Verify document](./sign-verify.md) to test the Sigstore stack. 
Good Luck!

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| configs.ctlog.create | bool | `true` |  |
| configs.ctlog.namespace | string | `"ctlog-system"` |  |
| configs.ctlog.rolebindings[0] | string | `"ctlog"` |  |
| configs.ctlog.rolebindings[1] | string | `"ctlog-createtree"` |  |
| configs.ctlog.rolebindings[2] | string | `"scaffolding-ctlog-createctconfig"` |  |
| configs.fulcio.create | bool | `true` |  |
| configs.fulcio.namespace | string | `"fulcio-system"` |  |
| configs.fulcio.rolebindings[0] | string | `"fulcio-createcerts"` |  |
| configs.fulcio.rolebindings[1] | string | `"fulcio-server"` |  |
| configs.fulcio.server.secret.name | string | `""` |  |
| configs.fulcio.server.secret.password | string | `""` |  |
| configs.fulcio.server.secret.private_key | string | `""` |  |
| configs.fulcio.server.secret.private_key_file | string | `""` |  |
| configs.fulcio.server.secret.public_key | string | `""` |  |
| configs.fulcio.server.secret.public_key_file | string | `""` |  |
| configs.fulcio.server.secret.root_cert | string | `""` |  |
| configs.fulcio.server.secret.root_cert_file | string | `""` |  |
| configs.rekor.create | bool | `true` |  |
| configs.rekor.namespace | string | `"rekor-system"` |  |
| configs.rekor.rolebindings[0] | string | `"rekor-redis"` |  |
| configs.rekor.rolebindings[1] | string | `"rekor-server"` |  |
| configs.rekor.rolebindings[2] | string | `"scaffolding-rekor-createtree"` |  |
| configs.rekor.signer.secret.name | string | `""` |  |
| configs.rekor.signer.secret.private_key | string | `""` |  |
| configs.rekor.signer.secret.private_key_file | string | `""` |  |
| configs.trillian.create | bool | `true` |  |
| configs.trillian.namespace | string | `"trillian-system"` |  |
| configs.trillian.rolebindings[0] | string | `"trillian-logserver"` |  |
| configs.trillian.rolebindings[1] | string | `"trillian-logsigner"` |  |
| configs.trillian.rolebindings[2] | string | `"trillian-mysql"` |  |
| configs.tuf.create | bool | `true` |  |
| configs.tuf.namespace | string | `"tuf-system"` |  |
| configs.tuf.rolebindings[0] | string | `"tuf"` |  |
| configs.tuf.rolebindings[1] | string | `"tuf-secret-copy-job"` |  |
| rbac.clusterrole | string | `"system:openshift:scc:anyuid"` |  |

----------------------------------------------
