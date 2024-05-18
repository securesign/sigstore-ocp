# OIDC configuration

While RHTAS uses the RHSSO operator (based on Keycloak) for OIDC by default, other OIDC options are available. Integrating with these options is described below. These steps should be followed after [RHTAS installation](https://access.redhat.com/documentation/en-us/red_hat_trusted_artifact_signer/1/html/deployment_guide/installing-trusted-artifact-signer-using-the-operator-lifecycle-manager_deploy), as they will assume that other components, including the Keycloak client for RHTAS, is present, and that the environment variables for Cosign initialization are [populated](https://github.com/securesign/sigstore-ocp/blob/main/sign-verify.md#signing-a-container-from-the-local-system). You may need to recreate the Keycloak client for RHTAS once the `sigstore` realm has been fully configured as described within this guide.

## Google

### Obtain Client ID and Secret for Google OAuth 2.0

To integrate RHTAS with Google as an identity provider, you must have access to a client ID and client secret for OAuth 2.0 via APIs & Services in Google Cloud Console. If you do not have these available, you can create them by following the directions [here](https://developers.google.com/workspace/guides/create-credentials#oauth-client-id) and specifying the following:

- The type of application required is Web Application 
- Authorized redirect URIs must include `http://localhost/auth/callback`

Once created, the credentials should be available in the Google Cloud Console under APIs & Services -> Credentials -> OAuth 2.0 Client IDs.

### Update Fulcio deployment configuration

Once RHTAS is installed, issue

```
oc edit cm fulcio-server-config -n fulcio-system
```

Update the configuration with the client ID and client secret, as well as the Google IdP issuer URL:

```
apiVersion: v1
data:
  config.json: |-
    {
      "OIDCIssuers": {
        "https://accounts.google.com": {
          "ClientID": "<my client ID>",
          "IssuerURL": "https://accounts.google.com",
          "Type": "email"
        }
      }
    }
kind: ConfigMap
...
```

Restart Fulcio with

```
oc delete $(oc get pods -n fulcio-system -o name) -n fulcio-system
```

You can check that the config updated properly with `oc describe cm fulcio-server-config -n fulcio-system`.

### Set OIDC issuer

The OIDC issuer environment variable must point to Google rather than Keycloak in the terminal where signing and verification with RHTAS are being executed. The correct issuer for Google is https://accounts.google.com.

```
export COSIGN_OIDC_ISSUER=https://accounts.google.com
```
This value overrides what is specified in the [sign-verify documentation](sign-verify.md). Be careful to avoid resetting `COSIGN_OIDC_ISSUER` when using the `sign-verify` documentation steps or sourcing the `tas-env-variables.sh` script. You can check what the environment variable's value is by issuing

```
$ echo $COSIGN_OIDC_ISSUER
```

It should show `https://accounts.google.com`.


### Pass in client ID and secret during signing

Create a client secret file that contains only the client secret:

`echo <my client secret> > <my secret filename>`

When issuing a `cosign sign` command, add the flags to pass in the client secret (as a file path) and the client ID (as a string):

```
cosign sign -y $IMAGE --oidc-client-secret-file=<my secret filename> --oidc-client-id=<my client ID>
```

## GitHub (federated via Keycloak)

GitHub federation via Keycloak can be added either from the CLI or from the Keycloak console. Each option is described below. The client ID and client secret must be obtained from GitHub prior to setting up the identity provider regardless of the option chosen.

### Obtain Client ID and Secret from GitHub OAuth App

To integrate RHTAS with GitHub as an identity provider, you must register your GitHub OAuth App in your Keycloak realm. 

Locate your GitHub OAuth App, or create a new GitHub OAuth App following the directions [here](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app). When creating the OAuth App, it is acceptable to enter placeholder values for the Authorization Callback and Homepage URL fields, as these fields will need to be updated to hold values sourced from Keycloak, which will be described later in this guide.

From the OAuth app's page, locate the client ID and create a new client secret, as these values will be required in later steps.

### Option 1: Create a New KeycloakRealm in OpenShift using the CLI

This option will create a new KeycloakRealm that includes the GitHub Identity Provider and Mapper using the CLI and a Custom Resource file, then add the Redirect URI to the OAuth app.

#### Create a New KeycloakRealm

Create a Custom Resource file called `realm-cr.yaml` from which to create the KeycloakRealm with the following content (note that you must add your own client ID and secret):

```shell
apiVersion: keycloak.org/v1alpha1
kind: KeycloakRealm
metadata:
  labels:
    app: sso
  name: sigstore
spec:
  instanceSelector:
    matchLabels:
      app: sso
  realm:
    displayName: Sigstore
    enabled: true
    id: sigstore
    realm: sigstore
    sslRequired: none
    identityProviders:
      - alias: github
        displayName: GitHub
        enabled: true
        firstBrokerLoginFlowAlias: first broker login
        internalId: github
        providerId: github
        trustEmail: true
        config:
          clientId: <my client id> # add your client ID
          clientSecret: <my client secret> # add your client secret
    identityProviderMappers:
      - name: github
        identityProviderAlias: github
        identityProviderMapper: hardcoded-attribute-idp-mapper
        config:
          attribute.value: "true"
          syncMode: INHERIT
          attribute: emailVerified
```

In the same namespace where Keycloak is deployed (by default `keycloak-system` with TAS), ensure no existing KeycloakRealm with the `metadata.name` specified in your `realm-cr.yaml` file (by default `sigstore` with TAS) is present in your cluster. If it is present, you must first delete it. 
**IMPORTANT:** make sure you are okay with deleting your current `sigstore` KeycloakRealm before issuing the command below.
```shell
$ oc delete keycloakrealm sigstore -n keycloak-system
```

Create the new KeycloakRealm that includes the GitHub Identity Provider using the custom resource file:
```shell
$ oc create -f realm-cr.yaml -n keycloak-system
```

Note: The Sigstore Public Good Instance uses [dex](https://github.com/dexidp/dex/) for federation, which uses this same [hardcoded configuration](https://github.com/dexidp/dex/blob/80d530d9bf0e38dfb0549c847c92e3c697b0402b/connector/github/github.go#L253) for the `emailVerified` value in the `identityProviderMapper`.

#### Add the Homepage and Authorization Callback URLs to the GitHub OAuth app

Obtain the Homepage URL that the GitHub OAuth app requires:
```shell
echo "$(oc get routes -n keycloak-system keycloak -o jsonpath='https://{.spec.host}')/auth/realms/sigstore/"
```

Obtain the Authorization Callback URL that the GitHub OAuth app requires:

```shell
echo "$(oc get routes -n keycloak-system keycloak -o jsonpath='https://{.spec.host}')/auth/realms/sigstore/broker/github/endpoint"
```

Paste these values into the appropriate OAuth app fields, then click `Update Application`:
<p align="center">
  <img src="/images/add_urls.png" />
</p>


### Option 2: Add the GitHub Identity Provider to an existing KeycloakRealm via the Keycloak console

Alternatively, the GitHub Identity Provider and Mapper can be added to an existing KeycloakRealm through the Keycloak console.

#### Log in to Keycloak console

You will need to log in to the Keycloak console to complete some steps of this process. Once RHTAS is installed, this can be achieved by locating the URL for the user interface and admin password.
Note: The command below assumes the name of the route is `keycloak`. 

```shell
$ oc get routes -n keycloak-system keycloak -o jsonpath='https://{.spec.host}'
```

You will need to log in. The username will be `admin`. You can retrieve the password with:
```shell
$ oc get secret/credential-keycloak -n keycloak-system -o jsonpath='{ .data.ADMIN_PASSWORD }' | base64 -d
```

#### Add the GitHub Identity Provider

From the Keycloak console, select `Identity Providers` in the lefthand menu bar, then select `Add provider -> GitHub` from the drop down menu.

You will need to fill in the following fields with these values to add the provider:

| Field              | Value                                                                                                                  |
| -------------------|------------------------------------------------------------------------------------------------------------------------|
| `Client ID`        | the client ID from your GitHub OAuth app's page                                                                        |
| `Client secret`    | the client secret from your GitHub OAuth app's page (note: you may need to generate a new secret to obtain this value) |
| `Enabled`          | `ON`                                                                                                                   |
| `Trust Email`      | `ON`                                                                                                                   |
| `First Login Flow` | `first broker login`                                                                                                   |

#### Add IdentityProviderMapper

From the Keycloak console, navigate to `Identity Providers` in the lefthand menu bar, and select the provider you have just created: `github`. Select the tab `Mappers` and click the `create` button on the right-hand side to bring up the following page:

<p align="center">
  <img src="/images/add_mapper.png" />
</p>

Set the field values as follows:

| Field                  | Value                                                 |
| -----------------------|-------------------------------------------------------|
| `Name`                 | this value can be anything, as it will be overwritten |
| `Sync Mode Override`   | `inherit`                                             |
| `Mapper Type`          | `Hardcoded Attribute`                                 |
| `User Attribute`       | `emailVerified`                                       |
| `User Attribute Value` | `true`                                                |


It should look like this: 

<p align="center">
  <img src="/images/mapper2.png" />
</p>

Make sure to hit `save`.

Note: The Sigstore Public Good Instance uses [dex](https://github.com/dexidp/dex/) for federation, which uses this same [hardcoded configuration](https://github.com/dexidp/dex/blob/80d530d9bf0e38dfb0549c847c92e3c697b0402b/connector/github/github.go#L253) for the `emailVerified` value.

### Copy the Redirect URI to your GitHub OAuth App

From the Keycloak console, navigate to `Identity Providers` in the lefthand menu bar, and select the provider you have just created: `github`. From this page, copy the Redirect URI:

<p align="center">
  <img src="/images/redirect_uri.png" />
</p>

Paste the redirect URI value into the OAuth app Authorization Callback URL field. Then paste it into the Homepage URL field, removing `broker/github/endpoint` from the end of the URL.

The values should look like this, where `CLUSTERVALUE` is the unique value from the URL of your OpenShift cluster:

|Field                       | Value|
|----------------------------|------|
| Homepage URL               | https://keycloak-keycloak-system.apps.rosa.CLUSTERVALUE.p3.openshiftapps.com/auth/realms/sigstore/ |
| Authorization Callback URL | https://keycloak-keycloak-system.apps.rosa.CLUSTERVALUE.p3.openshiftapps.com/auth/realms/sigstore/broker/github/endpoint |


Add these to the appropriate fields marked below:

<p align="center">
  <img src="/images/add_urls.png" />
</p>

Make sure to click `Update Application`.

### Sign the artifact

Follow the commands listed in [`sign-verify.md`](/sign-verify.md) to ensure that your TAS environment variables are populated, then initialize Cosign and sign the artifact. Another way to ensure that the TAS environment variables are populated is running `source ./tas-env-variables.sh` after setting up RHTAS with `./tas-easy-install.sh`.

**Troubleshooting:** If you have initialized Cosign previously, you may need to first delete `~/.sigstore` before running the commands below.

```shell
$ cosign initialize --mirror=$TUF_URL --root=$TUF_URL/root.json
$ cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$OIDC_ISSUER_URL $IMAGE
```

On the login page the pops up in the browser, make sure to select the GitHub login option at the bottom.

<p align="center">
  <img src="/images/github_login_2.png" />
</p>
