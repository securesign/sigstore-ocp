# Sigstore Helm Chart for OpenShift

**This chart offers an opinionated OpenShift-specific experience.** It is based on and directly depends on an upstream canonical [Sigstore Scaffold Helm chart](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold). For less opinionated experience, consider using the upstream chart directly.

This chart extends all the features in the upstream chart in addition to including OpenShift only features. It is not recommended to use this chart on other platforms.

## Usage

### Installing from the Chart Repository

Information on how to install Sigstore components on OpenShift can be found in the
[quickstart quide](./quick-start-with-keycloak.md)

## Scaffolding Chart

More information can be found by inspecting the [trusted-artifact-signer chart](charts/trusted-artifact-signer).
