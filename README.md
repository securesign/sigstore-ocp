# Sigstore Helm Chart for OpenShift

**This chart offers an opinionated OpenShift-specific experience.** It is based on and directly depends on an upstream canonical [Sigstore Scaffold Helm chart](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold). For less opinionated experience, consider using the upstream chart directly.

This chart extends all the features in the upstream chart in addition to including OpenShift only features. It is not recommended to use this chart on other platforms.

## Usage

### Installing from the Chart Repository

Information on how to install Sigstore components on OpenShift can be found in the
[quickstart quide](./quick-start-with-keycloak.md)

## Scaffolding Chart

<<<<<<< HEAD
More information can be found by inspecting the [trusted-artifact-signer chart](charts/trusted-artifact-signer).
=======
More information can be found by inspecting the [scaffolding chart](charts/scaffolding).

## Contributing

Install the [pre-commit](https://pre-commit.com/) package and run `pre-commit run --all-files` before pushing changes, or `pre-commit install` to automatically run the pre-commit hooks with every `git commit`. If it fails,
run the `git commit` command again. It's likely the pre-commit hook fixed the issue and you have to bring in the new changes.

### Testing

To set up a `kind` cluster and deploy the charts, run the following from the root of this repository

```bash
./kind/kind-up-test.sh

kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

OPENSHIFT_APPS_SUBDOMAIN=localhost envsubst <  ./examples/values-kind-sigstore.yaml | helm upgrade -i scaffolding --debug ./charts/scaffolding -n sigstore --create-namespace --values -
```

This test setup is to verify that all deployments are healthy and all jobs complete. However, this does not create a working environment to sign artifacts.

>>>>>>> a0f7581 (add kind)
