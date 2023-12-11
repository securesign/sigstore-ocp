## Fulcio root key injection

Utilize the following commands and configurations to inject Fulcio root secret:

First, generate a root key.
Open [fulcio-create-CA script](../fulcio-create-root-ca-openssl.sh) to check out the commands before running it.
The `openssl` commands are interactive.

```shell
./fulcio-create-root-ca-openssl.sh
```

The command creates keys and cert in `./keys-cert` folder.
Either create a secret in the fulcio-system namespace with:

```bash
# Note replace <PASSWORD> with value of password to decrypt signing key created above.
# if necessary, 'oc create ns fulcio-system'

oc -n fulcio-system create secret generic fulcio-secret-rh --from-file=private=./keys-cert/file_ca_key.pem --from-file=public=./keys-cert/file_ca_pub.pem --from-file=cert=./keys-cert/fulcio-root.pem  --from-literal=password=<PASSWORD> --dry-run=client -o yaml | oc apply -f-
```

Or, add the following to an overriding Values file injecting the public key, private key, and password used for the private key:

```yaml
configs:
  fulcio:
    server:
      secret:
        name: "fulcio-secret-rh"
        password: "<password>"
        public_key_file: "keys-cert/file_ca_pub.pem"
        private_key_file: "keys-cert/file_ca_key.pem"
        root_cert_file: "keys-cert/fulcio-root.pem"
```

## Rekor Signer Key

Open [rekor create signer script](../rekor-create-signer-key.sh) to check out the commands before running it.
Generate a signer key:

```shell
./rekor-create-signer-key.sh
```

Either create a secret in the rekor-system namespace with:

```bash
# if necessary, 'oc create ns rekor-system'
oc -n rekor-system create secret generic rekor-private-key --from-file=private=rekor_key.pem --dry-run=client -o yaml | oc apply -f-
```

Or, add the following to override the values file injecting the signer key:

```yaml
configs:
  rekor:
    signer:
      secret:
        name: rekor-private-key
        private_key_file: "keys-cert/rekor_key.pem"
```

NOTE: The name of the generated secret, `rekor-private-key` can be customized.
Ensure the naming is consistent throughout each of the customization options
