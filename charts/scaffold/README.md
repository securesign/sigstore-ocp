# Sigstore OCP Helm Chart

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

A Helm chart for deploying a Sigstore application opinionated for OpenShift

---

`sigstore-rhel/scaffold` chart is an opinionated flavor of the upstream sigstore/scaffold chart located at [sigstore/helm-charts/charts/scaffold](https://github.com/sigstore/helm-charts/tree/main/charts/scaffold). It extends the upstream chart with additional OpenShift specific functionality and provides opinionated values.

## TL;DR

```console
helm repo add sigstore https://sigstore.github.io/helm-charts
helm repo add sigstore/scaffold

helm upgrade -i my-sigstore-scaffold sigstore/scaffold -f values.yaml
```

## Introduction

This chart bootstraps a [Sigstore](https://www.sigstore.dev/) stack on [OpenShift](https://docs.openshift.com/)
using the [Helm](https://helm.sh) package manager. A Sigstore stack includes all or some of the following components:

1. Rekor
2. Trillian
3. Fulcio
4. Certificate Transparency Log
5. TUF
6. Timestamp Authority

## Prerequisites

- OpenShift 4.13+
- Helm 3.2.0+

## Usage

Chart is available in the following formats:

- [Chart Repository](https://helm.sh/docs/topics/chart_repository/)
- [OCI Artifacts](https://helm.sh/docs/topics/registries/)

### Uninstalling the Chart

To uninstall/delete the `my-sigstore-scaffold` deployment:

```console
helm uninstall my-sigstore-scaffold
```

The command removes all the OpenShift components associated with the chart and deletes the release.

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://sigstore.github.io/helm-charts | ctlog | 0.2.40 |
| https://sigstore.github.io/helm-charts | fulcio | 2.2.0 |
| https://sigstore.github.io/helm-charts | rekor | 1.3.1 |
| https://sigstore.github.io/helm-charts | trillian | 0.2.2 |
| https://sigstore.github.io/helm-charts | tsa | 0.1.0 |
| https://sigstore.github.io/helm-charts | tuf | 0.1.4 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| copySecretJob.backoffLimit | int | `6` |  |
| copySecretJob.enabled | bool | `false` |  |
| copySecretJob.imagePullPolicy | string | `"IfNotPresent"` |  |
| copySecretJob.name | string | `"copy-secrets-job"` |  |
| copySecretJob.registry | string | `"docker.io"` |  |
| copySecretJob.repository | string | `"alpine/k8s"` |  |
| copySecretJob.serviceaccount | string | `"tuf-secret-copy-job"` |  |
| copySecretJob.version | string | `"sha256:fb0d2db81fb0f98abb1adf5246d6f0f4d19f34031afe4759cb7ad8e2eb8d2c01"` |  |
| ctlog.createcerts.fullnameOverride | string | `"ctlog-createcerts"` |  |
| ctlog.createtree.displayName | string | `"ctlog-tree"` |  |
| ctlog.createtree.fullnameOverride | string | `"ctlog-createtree"` |  |
| ctlog.enabled | bool | `true` |  |
| ctlog.forceNamespace | string | `"ctlog-system"` |  |
| ctlog.fullnameOverride | string | `"ctlog"` |  |
| ctlog.namespace.create | bool | `true` |  |
| ctlog.namespace.name | string | `"ctlog-system"` |  |
| fulcio.createcerts.fullnameOverride | string | `"fulcio-createcerts"` |  |
| fulcio.ctlog.createctconfig.logPrefix | string | `"sigstorescaffolding"` |  |
| fulcio.ctlog.enabled | bool | `false` |  |
| fulcio.enabled | bool | `true` |  |
| fulcio.forceNamespace | string | `"fulcio-system"` |  |
| fulcio.namespace.create | bool | `true` |  |
| fulcio.namespace.name | string | `"fulcio-system"` |  |
| fulcio.server.fullnameOverride | string | `"fulcio-server"` |  |
| rekor.enabled | bool | `true` |  |
| rekor.forceNamespace | string | `"rekor-system"` |  |
| rekor.fullnameOverride | string | `"rekor"` |  |
| rekor.namespace.create | bool | `true` |  |
| rekor.namespace.name | string | `"rekor-system"` |  |
| rekor.redis.fullnameOverride | string | `"rekor-redis"` |  |
| rekor.server.fullnameOverride | string | `"rekor-server"` |  |
| rekor.trillian.enabled | bool | `false` |  |
| trillian.enabled | bool | `true` |  |
| trillian.forceNamespace | string | `"trillian-system"` |  |
| trillian.fullnameOverride | string | `"trillian"` |  |
| trillian.logServer.fullnameOverride | string | `"trillian-logserver"` |  |
| trillian.logServer.name | string | `"trillian-logserver"` |  |
| trillian.logServer.portHTTP | int | `8090` |  |
| trillian.logServer.portRPC | int | `8091` |  |
| trillian.logSigner.fullnameOverride | string | `"trillian-logsigner"` |  |
| trillian.logSigner.name | string | `"trillian-logsigner"` |  |
| trillian.mysql.fullnameOverride | string | `"trillian-mysql"` |  |
| trillian.namespace.create | bool | `true` |  |
| trillian.namespace.name | string | `"trillian-system"` |  |
| tsa.enabled | bool | `true` |  |
| tsa.forceNamespace | string | `"tsa-system"` |  |
| tsa.namespace.create | bool | `true` |  |
| tsa.namespace.name | string | `"tsa-system"` |  |
| tsa.server.fullnameOverride | string | `"tsa-server"` |  |
| tuf.enabled | bool | `false` |  |
| tuf.forceNamespace | string | `"tuf-system"` |  |
| tuf.fullnameOverride | string | `"tuf"` |  |
| tuf.namespace.create | bool | `true` |  |
| tuf.namespace.name | string | `"tuf-system"` |  |
| tuf.secrets.ctlog.name | string | `"ctlog-public-key"` |  |
| tuf.secrets.ctlog.path | string | `"ctlog-pubkey"` |  |
| tuf.secrets.fulcio.name | string | `"fulcio-server-secret"` |  |
| tuf.secrets.fulcio.path | string | `"fulcio-cert"` |  |
| tuf.secrets.rekor.name | string | `"rekor-public-key"` |  |
| tuf.secrets.rekor.path | string | `"rekor-pubkey"` |  |

## Features

This charts defaults to using the Red Hat images for Sigstore components that are OpenShift compatible:

```
quay.io/something/fulcio:tag
quay.io/something/rekor:tag
quay.io/something/trillian:tag
quay.io/something/ctlog:tag
quay.io/something/tuf:tag
```
