## Trusted Artifact Signer Installer

The `tas-install` command is experimental. To build it, run the following from the repository root.

```
go build -C tas-installer -o ../tas-install
```

### Install 

```
 $ ./tas-install install -h
Installs Trusted Artifact Signer (TAS) on a Kubernetes cluster.

	This command performs a series of actions:
	1. Initializes the Kubernetes client to interact with your cluster
	2. Sets up necessary certificates
	3. Configures secrets
	4. Deploys TAS to openshift

Usage:
  tas-installer install [flags]

Flags:
      --chart-location string    /local/path/to/chart or oci://registry/repo location of Helm chart (default "./charts/trusted-artifact-signer")
      --chart-version string     Version of the Helm chart (default "0.1.29")
  -h, --help                     help for install
      --oidc-client-id string    Specify the OIDC client ID
      --oidc-issuer-url string   Specify the OIDC issuer URL e.g for keycloak: https://[keycloak-domain]/auth/realms/[realm-name]
      --oidc-type string         Specify the OIDC type
      --values string            path to custom values file for chart configuration

Global Flags:
      --kubeconfig string   Specify the kubeconfig path (default "$HOME/.kube/config")
```

### Uninstall

```
$ ./tas-install uninstall -h
Removes installation of Trusted Artifact Signer (TAS) on a Kubernetes cluster.

Usage:
  tas-installer uninstall [flags]

Flags:
  -h, --help   help for uninstall

Global Flags:
      --kubeconfig string   Specify the kubeconfig path (default "$HOME/.kube/config")
```
