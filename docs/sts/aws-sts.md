## Sigstore on OpenShift, with AWS Security Token Service (STS)

[Running pods in OpenShift with AWS IAM Roles for service accounts (IRSA)](https://cloud.redhat.com/blog/running-pods-in-openshift-with-aws-iam-roles-for-service-accounts-aka-irsa)

AWS STS is used to create workload identity tokens for service accounts.
Sigstore is configured to use the service account OIDC Identity Token to pass to Fulcio to authenticate requests.

[Understanding ROSA with STS](https://docs.openshift.com/rosa/rosa_getting_started/rosa-sts-getting-started-workflow.html)

### Configure Fulcio chart with AWS STS OIDC issuer

In your values file the `scaffold.fulcio` section should include the following:

```yaml
fulcio:
  config:
    contents:
      OIDCIssuers:
        # replace
        ? https://rh-oidc.s3.us-east-1.amazonaws.com/.....
        : IssuerURL: https://rh-oidc.s3.us-east-1.amazonaws.com/.......
          ClientID: sigstore
          Type: kubernetes
```

### Create a service account and signer deployment

For this you will need to have an `IAM role` for your AWS Identity Provider with
permissions to list S3 buckets. From the AWS Console, choose
`Roles-> Create Role -> Web Identity`.
Choose your Identity provider from the dropdown list in your account and Audience.
Next, you'll need to add the Policy `AmazonS3ReadOnlyAccess`.
Note the ARN `arn:aws:iam::xxxx:role/xxxxxxxx`, to add to the cosign service account.

Inspect the cosign deployment manifests and make any necessary changes. These are in `./docs/sts`.

Create the `cosign-sts` serviceaccount and deployment

```bash
oc apply -f docs/sts/aws-sts-sa.yaml
oc apply -f docs/sts/cosign-dep.yaml
```

Finally, [cosign](https://github.com/sigstore/cosign) can be used in the pod to sign and verify artifacts.
As a PoC, here we will exec into `cosign-sts-* -n cosign` pod and run the following:

```bash
oc get pods -n cosign | grep cosign-sts
```

### Sign and verify images

First, initialize the TUF roots.

```shell
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign initialize --mirror=$TUF_URL --root=$TUF_URL/root.json'
```

Login to the image repository of your choice using cosign.

```
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign login <repo> -u <username> -p <password>'
```

Sign an image

```
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign sign -y --fulcio-url=$FULCIO_URL --rekor-url=$REKOR_URL --oidc-issuer=$OIDC_ISSUER_URL --identity-token=$AWS_WEB_IDENTITY_TOKEN_FILE <image>'
```

Verify an image

```shell
oc exec -n cosign <pod_name> -- /bin/sh -c 'cosign verify --rekor-url=$REKOR_URL --certificate-identity https://kubernetes.io/namespaces/cosign/serviceaccounts/cosign --certificate-oidc-issuer $OIDC_ISSUER_URL <image>'
```
