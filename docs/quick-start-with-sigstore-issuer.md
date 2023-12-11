## Quick Start with Sigstore Public OIDC Issuer

1. Use the installer's `install` command to install the required signing keys and root certificate for keyless signing and install the sigstore stack.

To build the installer

```
go build -C tas-installer -o ../tas-install
```

The installer expects a `kubeconfig` file at `$HOME/.kube/config`,, or that the flag `--kubeconfig /path/to/kubeconfig` is provided.
By default, the fulcio server is configured to use the upstream public OIDC issuer at `oauth2.sigstore.dev/auth`. An interactive browser
based flow in which you will authenticate with Google, GitHub, or MicroSoft will be initiated when signing artifacts..

First, the user is prompted for information in order to create rekor and fulcio signing keys as well as the fulcio root certificate.
Then, the Trusted Artifact Signer resources will be created. The stack is ready to use when all jobs have been completed. The job
in the `tuf-system` namespace will be the last to complete, and can take several minutes.
 
```shell
./tas-install install
```

 Watch `oc get jobs -A` and when the `tuf-system` job is complete, the TAS stack should be ready to sign & verify artifacts.

Once complete, move to the [Sign & Verify document](sign-verify.md) to test the Sigstore stack.

If there is already a helm release `trusted-artifact-signer` installed, the command `./tas-install install` will perform an upgrade.
In this case, it will reuse the signing keys and certificate secrets from the connected cluster's `fulcio-system` and `rekor-system`
namespaces.

