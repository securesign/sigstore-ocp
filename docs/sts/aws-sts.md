## Sigstore on OpenShift, with AWS Security Token Service (STS)

AWS STS is used to create workload identity tokens for service accounts.
Trusted Artifact Signer is configured to obtain an OIDC Identity Token for Fulcio to authenticate requests. Included in this doc
is an example deployment that contains the cosign binary and a service account that can be used to sign and verify OCI artifacts.

If running with Red Hat OpenShift on AWS (ROSA), refer to
[Understanding ROSA with STS](https://docs.openshift.com/rosa/rosa_getting_started/rosa-sts-getting-started-workflow.html) to create a cluster with an
OIDC identity provider (IdP).
Otherwise, refer to [AWS docs on creating OIDC identity provider](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html).
However the IdP is created, it must have an Audience that is `sigstore`.

Along with an IAM OIDC identity provider, you must create one or more IAM roles. A role is an identity in AWS that doesn't have its own credentials
(as a user does). A role is dynamically assigned to a federated user that is authenticated by your organization's IdP.
The role permits your organization's IdP to request temporary security credentials for access to AWS. The role must be configured to provide
web identity federation. No other permissions are required for use with TAS.

To continue, you should have:

- The OIDC Issuer URL for an IAM OIDC IdP with Audience set to `sigstore`.
    - [AWS docs on creating OIDC identity provider](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html)
    - To find an OpenShift cluster's AWS OIDC Issuer URL:
        - `oc get authentication cluster -o jsonpath='{.spec.serviceAccountIssuer}'` 
- The ARN of a role for web identity federation `arn:aws:iam::xxxx:role/xxxxxxxx` associated with the identity provider
    - [AWS docs for creating roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user.html)

### Configure Fulcio chart with AWS STS OIDC issuer

After you have created the IAM OIDC identity provider with an associated IAM role, update or configure the fulcio server configuration.
The configuration is stored in a configmap named `fulcio-server-config` in the same namespace where the `fulcio-server` is running..

Update fulcio-server-config to add the AWS OIDC identity provider issuer URL:

```yaml
oc get cm fulcio-server-config -n fulcio-system -o yaml
apiVersion: v1
data:
  config.json: |-
    {
      "OIDCIssuers": {
        # AWS identity provider 
        "https://example.s3.us-east-1.amazonaws.com/xxxxxx": {
          "ClientID": "sigstore",
          "IssuerURL": "https://example.s3.us-east-1.amazonaws.com/xxxxxx",
          "Type": "kubernetes"
        },
        "https://other-oidc-issuer-url": {
          "ClientID": "sigstore",
          "IssuerURL": "https://other-oidc-issuer-url",
          "Type": "email"
        },
      }
---
```

### Create a service account and signer deployment

Inspect the cosign deployment manifests and make any necessary changes. These are in `./docs/sts`.
To see the required values to update, look for `UPDATE_ME`

Create the `cosign-sts` serviceaccount and deployment

```bash
oc create ns cosign
oc apply -f docs/sts/aws-sts-sa.yaml
oc apply -f docs/sts/cosign-dep.yaml
```

Finally, [cosign](https://github.com/sigstore/cosign) can be used in the pod to sign and verify artifacts.
As a PoC, here we will exec into the cosigin-sts pod in the cosign namespace.
First, find the pod name.

```bash
POD_NAME=$(oc get pods -n cosign -l app=cosign-sts -o jsonpath='{ .items[0].metadata.name }')
echo ${POD_NAME} # ensure it's populated, pod name is used for below commands
```

### Sign and verify images

Login to the image repository of choice and initialize cosign.

```
oc exec -n cosign ${POD_NAME} -- /bin/sh -c 'cosign login <repo> -u <username> -p <password> && cosign initialize'
```

Sign an image

```
oc exec -n cosign ${POD_NAME} -- /bin/sh -c 'cosign sign -y --identity-token=$AWS_WEB_IDENTITY_TOKEN_FILE <image>'
```

Verify an image

```shell
oc exec -n cosign ${POD_NAME} -- /bin/sh -c 'cosign verify --certificate-identity https://kubernetes.io/namespaces/cosign/serviceaccounts/cosign-sts <image>'
```
