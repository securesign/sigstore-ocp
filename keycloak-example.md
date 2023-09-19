## Deploy RH SSO Operator and Keycloak resources in OpenShift

Follow this to deploy a Keycloak instance with the following:

- A dedicated `Realm` (Recommended)
- A `Client` representing the Sigstore integration
- Valid `Redirect URIs`
    - A value of `*` can be used for testing
- 1 or more `Users`
    - Email Verified

The RHSSO Operator and necessary Keycloak resources are deployed with:

```shell
oc apply --kustomize keycloak/operator

# wait for this command to succeed before going on to be sure the Keycloak CRDs are registered
oc get keycloaks -A

oc apply --kustomize keycloak/resources
# wait for keycloak-system pods to be running before proceeding
```

### Add keycloak user and/or credentials

Check out the [user custom resource](https://github.com/redhat-et/sigstore-rhel/blob/main/helm/scaffold/overlays/keycloak/user.yaml)
for how to create a keycloak user. For testing, a user `jdoe@redhat.com` with password: `secure` is created.

You can access the keycloak route and login as the admin user to set credentials in the keycloak admin console.
To get the keycloak admin credentials, run `oc extract secret/credential-keycloak -n keycloak-system`.
This will create an `ADMIN_PASSWORD` file with which to login.
