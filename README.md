# Sigstore Helm Chart for OpenShift

**This chart offers an opinionated OpenShift-specific experience.** It is based on and directly depends on an upstream canonical [Sigstore Scaffold Helm chart](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold). For less opinionated experience, consider using the upstream chart directly.

This chart extends all the features in the upstream chart in addition to including OpenShift only features. It is not recommended to use this chart on other platforms.

## Usage

### Installing from the Chart Repository

For a quickstart on how to install Sigstore components on OpenShift refer to the
[quickstart quide](./quick-start-with-keycloak.md)

## Scaffolding Chart

More information can be found by inspecting the [trusted-artifact-signer chart](charts/trusted-artifact-signer).

## Contributing

Install the [pre-commit](https://pre-commit.com/) package and run `pre-commit run --all-files` before pushing changes, or `pre-commit install` to automatically run the pre-commit hooks with every `git commit`. If it fails,
run the `git commit` command again. It's likely the pre-commit hook fixed the issue and you have to bring in the new changes.

### Testing

To set up a `kind` cluster and deploy the charts, run the following from the root of this repository

```bash
./kind/kind-up-test.sh
<<<<<<< HEAD
=======

oc wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

OPENSHIFT_APPS_SUBDOMAIN=ingress.local envsubst <  ./examples/values-kind-sigstore.yaml | helm upgrade -i trusted-artifact-signer --debug ./charts/trusted-artifact-signer --wait --wait-for-jobs -n trusted-artifact-signer --create-namespace --values -

helm test -n sigstore trusted-artifact-signer
# tests are in charts/trusted-artifact-signer/templates/tests
>>>>>>> b419522 (add doc for using self-signed cluster)
```

This script will setup new KinD cluster for you, deploy the sigstore and execute tests on the deployment.
There are no ingress routes so the sigstore is accessible only from the container. (see https://github.com/securesign/sigstore-ocp/blob/main/sign-verify.md#signing-a-container-using-the-cosign-pod)

To uninstall helm chart:

```bash
helm uninstall trusted-artifact-signer -n sigstore
```

To cleanup the test kind cluster, run:

```bash
sudo kind delete cluster
```

