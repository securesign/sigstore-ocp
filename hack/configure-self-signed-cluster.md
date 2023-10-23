## Local setup when OpenShift cluster is using a self-signed ingress certificate

With local development clusters, often the cluster certificate is self-signed rather than issued from a trusted CA. A few steps must be taken to get a working Sigstore stack.

### Extract the cluster ingress certificate

From OpenShift, the default ingress cert is in `-n openshift-ingress`

```bash
mkdir clustercert && cd clustercert
oc extract secret/default-ingress-cert -n openshift-ingress
cd ../
```

Configure your local terminal to trust the cluster certificate.

```bash
export SSL_CERT_DIR=$(pwd)/clustercert
```

Next, run the install script as usual

```bash
/path/to/tas-easy-install.sh
```

Once the fulcio-system ns exists, create the ingress-cert secret in the fulcio-system namespace

```bash
oc create secret generic -n fulcio-system clustercert --from-file=/path/to/clustercert/tls.crt
```

Finally, patch the fulcio-server deployment in order for
fulcio to trust the ingress certificate for the keycloak OIDC endpoint.

```bash
oc patch deployment/fulcio-server -n fulcio-system --patch-file /path/to/securesign/sigstore-ocp/hack/fulcio-patch-self-signed-oidc.yaml
```

Now wait for all jobs to complete, then sign as usual. Refer to [the sign and verify doc](../sign-verify.md).
