## Signing a Container

Utilize the following steps to sign a container that has been published to an OCI registry

1. Export the following environment variables substituting `base_hostname` with the value used as part of the provisioning

The `base_hostname` can be obtained from

```shell
oc get dnsrecords -o yaml -n openshift-ingress-operator | grep dnsName
# omit the '*.'
export BASE_HOSTNAME=apps.something.something.openshiftapps.com
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
