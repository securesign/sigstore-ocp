{{/*
Expand the name of the chart.
*/}}
{{- define "keycloak.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "keycloak.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create the chart namespace.
*/}}
{{- define "keycloak.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "keycloak.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "keycloak.labels" -}}
helm.sh/chart: {{ include "keycloak.chart" . }}
{{ include "keycloak.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Common PostgreSQL labels
*/}}
{{- define "keycloak.postgresql.labels" -}}
helm.sh/chart: {{ include "keycloak.chart" . }}
{{ include "keycloak.postgresql.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "keycloak.selectorLabels" -}}
app.kubernetes.io/name: {{ include "keycloak.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "keycloak.postgresql.selectorLabels" -}}
app.kubernetes.io/name: {{ include "keycloak.postgresql.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "keycloak.serviceAccountName" -}}
{{- if .serviceAccount.create }}
{{- default .name .serviceAccount.name }}
{{- else }}
{{- default "default" .serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the image path for the passed in image field
*/}}
{{- define "keycloak.image" -}}
{{- if eq (substr 0 7 .version) "sha256:" -}}
{{- printf "%s/%s@%s" .registry .repository .version -}}
{{- else -}}
{{- printf "%s/%s:%s" .registry .repository .version -}}
{{- end -}}
{{- end -}}

{{/*
Name of the PostgreSQL instance
*/}}
{{- define "keycloak.postgresql.name" -}}
{{- (printf "%s-%s" (include "keycloak.name" $) "postgresql") | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Full Name of the PostgreSQL instance
*/}}
{{- define "keycloak.postgresql.fullname" -}}
{{- (printf "%s-%s" (include "keycloak.fullname" $) "postgresql") | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Name of the PostgreSQL Secret
*/}}
{{- define "keycloak.postgresql.secret.name" -}}
{{ default (include "keycloak.postgresql.fullname" .) (.Values.postgresql.secret.existingSecret) }}
{{- end }}

{{/*
Name of the Keycloak Service
*/}}
{{- define "keycloak.service.name" -}}
{{ .Values.openshift | ternary (printf "%s-trusted" (include "keycloak.fullname" .))  (include "keycloak.fullname" .)  }}
{{- end }}

{{/*
Name of the TLS Secret
*/}}
{{- define "keycloak.tls.secret.name" -}}
{{- $defaultTls := printf "%s-tls" (include "keycloak.fullname" .)  }}
{{- if .Values.openshift -}}
{{- default $defaultTls .Values.keycloak.tls.secret -}}
{{- else -}}
{{ len .Values.keycloak.tls.secret }}
{{- end }}
{{- end }}
