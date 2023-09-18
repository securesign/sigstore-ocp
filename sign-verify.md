## Signing a Container

Utilize the following steps to sign a container that has been published to an OCI registry

1. Export the following environment variables substituting `base_hostname` with the value used as part of the provisioning

The `base_hostname` can be obtained from

```shell
oc get dns cluster -o jsonpath='{ .spec.baseDomain }'
export BASE_HOSTNAME=apps.BASE_DOMAIN
```

The following assumes there exists a Keycloak `keycloak` in namespace `keycloak-system`

```shell
export KEYCLOAK_REALM=sigstore
export FULCIO_URL=https://fulcio.$BASE_HOSTNAME
export KEYCLOAK_URL=https://keycloak-keycloak-system.$BASE_HOSTNAME
export REKOR_URL=https://rekor.$BASE_HOSTNAME
export TUF_URL=https://tuf.$BASE_HOSTNAME
export KEYCLOAK_OIDC_ISSUER=$KEYCLOAK_URL/auth/realms/$KEYCLOAK_REALM
```

2. Initialize the TUF roots

```shell
cosign initialize --mirror=$TUF_URL --root=$TUF_URL/root.json
```

Note: If you have used `cosign` previously, you may need to delete the `~/.sigstore` directory

3. Sign the desired container

```shell
cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$KEYCLOAK_OIDC_ISSUER  <image>
```

Authenticate with the Keycloak instance using the desired credentials.

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

Follow the steps below to sign a container with the cosign pod that has been published to an OCI registry.

If the `BASE_HOSTNAME` environmental variable is not already specified in the Helm chart, make sure to set it in the cosign pod.

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

4. Retrieve `id_token` from the keycloak provider.
```
curl -X POST -H "Content-Type: application/x-www-form-urlencoded" \
-d "client_id=<client_id>" \
-d "username=<username>" \
-d "password=<password>" \
-d "grant_type=password" \
-d "scope=openid" \
<Keycloak_issuer_url>/auth/realms/<client_id>/protocol/openid-connect/token
```

5. Sign the container.
```
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$KEYCLOAK_OIDC_ISSUER --identity-token=<id_token> <image>'
```

6. Verify the signed image.

```shell
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign verify --rekor-url=$REKOR_URL --certificate-identity-regexp sigstore-user --certificate-oidc-issuer-regexp keycloak <image>'
```

If the signature verification did not result in an error, the deployment of Sigstore was successful!