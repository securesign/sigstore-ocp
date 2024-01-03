## Signing a Container From the Local System

Utilize the following steps to sign a container that has been published to an OCI registry, with the cosign client running on your local system and the RHTAS stack running in an OpenShift cluster as documented [here](quick-start-with-keycloak.md).

1. Export the following environment variables substituting `base_hostname` with the value used as part of the provisioning

The OpenShift subdomain can be obtained from

```shell
OPENSHIFT_APPS_SUBDOMAIN=apps.$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')
```

The following assumes there exists a Keycloak `keycloak` in namespace `keycloak-system`

```shell
export OIDC_AUTHENTICATION_REALM=sigstore
export FULCIO_URL=https://fulcio.$OPENSHIFT_APPS_SUBDOMAIN
export OIDC_ISSUER_URL=https://keycloak-keycloak-system.$OPENSHIFT_APPS_SUBDOMAIN/auth/realms/$OIDC_AUTHENTICATION_REALM
export REKOR_URL=https://rekor.$OPENSHIFT_APPS_SUBDOMAIN
export TUF_URL=https://tuf.$OPENSHIFT_APPS_SUBDOMAIN
```

2. Initialize the TUF roots

If you are using a cluster with self-signed certificates, you will first need to download the `kube-root-ca.crt` from the cluster and add it to your
local system's trusted certificate store. On Fedora/RHEL, this will look like:

```shell
oc extract cm/kube-root-ca.crt -n openshift-ingress-operator
mv ca.crt kube-root-ca.crt
sudo mv kube-root-ca.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust extract
```
On other systems, the last two commands may differ.

Then you can initialize cosign, or start here if you are not using a cluster with self-signed certificates.
Note: If you have used `cosign` previously, you may need to first delete the `~/.sigstore` directory

```shell
cosign initialize --mirror=$TUF_URL --root=$TUF_URL/root.json
```

3. Sign the desired container

```shell
cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$OIDC_ISSUER_URL  <image>
```

Authenticate with the OIDC provider (Keycloak, here)  using the desired credentials.

4. Verify the signed image

This example that verifies an image signed with email identity `sigstore-user@email.com` and issuer `https://keycloak-keycloak.apps.com/auth/realms/sigstore`.

```shell
cosign verify \
--rekor-url=$REKOR_URL \
--certificate-identity-regexp sigstore-user \
--certificate-oidc-issuer-regexp keycloak  \
<image>
```

If the signature verification did not result in an error, the deployment of Sigstore was successful!

## Signing a Container Using the Cosign pod.

Follow the steps below to sign an artifact using the cosign pod running in the cosign namespace.

The `OPENSHIFT_APPS_SUBDOMAIN` environmental variable should be specified in the trusted-artifact-signer chart,
with `global.appsSubdomain`. If it isn't, you'll need to set that variable in the cosign
deployment pod specification.

1. Get the name of the pod.

``` 
oc get pods -n cosign 
```

2. Initialize the TUF roots.

```shell
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign initialize --mirror=$TUF_URL --root=$TUF_URL/root.json'
```

3. Login to the image repository of your choice using cosign.
```
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign login <repo> -u <username> -p <password>'
```

4. Retrieve `id_token` from the OIDC provider.
```
curl -X POST -H "Content-Type: application/x-www-form-urlencoded" \
-d "client_id=<client_id>" \
-d "username=<username>" \
-d "password=<password>" \
-d "grant_type=password" \
-d "scope=openid" \
<oidc_issuer_url>/protocol/openid-connect/token
```

5. Sign the container.
```
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$OIDC_ISSUER_URL --identity-token=<id_token> <image>'
```

6. Verify the signed image. Again, this example assumes `Keycloak` is the OIDC provider.

```shell
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign verify --rekor-url=$REKOR_URL --certificate-identity-regexp sigstore-user --certificate-oidc-issuer-regexp keycloak <image>'
```

If the signature verification did not result in an error, the deployment of Sigstore was successful!

## Signing an image on Windows using PowerShell

Use the following steps to sign a container that has been published to an OCI registry, on Windows using `PowerShell`.

1. Define the following environment variables
```
$OPENSHIFT_APPS_SUBDOMAIN = "apps." + $(oc get dns cluster -o jsonpath='{.spec.baseDomain}')
$OIDC_AUTHENTICATION_REALM = "sigstore"
$FULCIO_URL = "https://fulcio." + $OPENSHIFT_APPS_SUBDOMAIN
$OIDC_ISSUER_URL = "https://keycloak-keycloak-system." + $OPENSHIFT_APPS_SUBDOMAIN + "/auth/realms/" + $OIDC_AUTHENTICATION_REALM
$REKOR_URL = "https://rekor." + $OPENSHIFT_APPS_SUBDOMAIN
$TUF_URL = "https://tuf." + $OPENSHIFT_APPS_SUBDOMAIN
```
2. Initialize the TUF roots

Note: If you have used `cosign` previously, you may need to first delete the `C:\Users\<user>\.sigstore` directory

```
cosign initialize --mirror=$TUF_URL --root=$TUF_URL/root.json
```

3. Sign the desired container

    3.1. Signing with Upstream's Public goods instance

    ```
    cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL <image>
    ```

    3.2. Signing with keycloak

    Retrieve the `id_token` from the keycloak instance

    ```
    $body = @{ 
                client_id = "<client_id>" 
                username  = "<username>"
                password  = "<password>"
                grant_type = "password" 
                scope = "openid"
            }

    $id_token = (Invoke-WebRequest -Uri "$OIDC_ISSUER_URL/protocol/openid-connect/token" -Method Post -Body $body -ContentType "application/x-www-form-urlencoded" -UseBasicParsing).Content | ConvertFrom-Json | Select-Object -ExpandProperty id_token 
    ```

    Sign the container

    ```
    cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$OIDC_ISSUER_URL --identity-token=$id_token <image>
    ```

4. Verify the desired container

```
cosign verify --rekor-url=$REKOR_URL --certificate-identity-regexp <identity> --certificate-oidc-issuer-regexp <oidc-issuer> <image>
```