# OIDC configuration

While RHTAS uses the RHSSO operator (based on Keycloak) by default, other OIDC options are available. Integrating with these options is described below. These steps should be followed after RHTAS installation.

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
export OIDC_ISSUER_URL=https://accounts.google.com
```
This value overrides what is specified in the [sign-verify documentation](sign-verify.md). Be careful to avoid resetting `OIDC_ISSUER_URL` when using the `sign-verify` documentation steps or sourcing the `tas-env-variables.sh` script. You can check what the environment variable's value is by issuing

```
$ echo $OIDC_ISSUER_URL
```

It should show `https://accounts.google.com`.


### Pass in client ID and secret during signing

Create a client secret file that contains only the client secret:

`echo <my client secret> > <my secret filename>`

When issuing a `cosign sign` command, add the flags to pass in the client secret (as a file path) and the client ID (as a string):

```
cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$OIDC_ISSUER_URL $IMAGE --oidc-client-secret-file=<my secret filename> --oidc-client-id=<my client ID>
```
