{{/*
Expand the name of the chart.
*/}}
{{- define "archimedes-property-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "archimedes-property-operator.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "archimedes-property-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "archimedes-property-operator.labels" -}}
helm.sh/chart: {{ include "archimedes-property-operator.chart" . }}
{{ include "archimedes-property-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "archimedes-property-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "archimedes-property-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Additional pod annotations
*/}}
{{- define "archimedes-property-operator.annotations" -}}
{{- if .Values.podAnnotations }}
{{- toYaml .Values.podAnnotations }}
{{- end }}
{{- end -}}

{{/*
Additional test-connection pod annotations
*/}}
{{- define "archimedes-property-operator.testPodAnnotations" -}}
{{- if .Values.testPodAnnotations }}
{{- toYaml .Values.testPodAnnotations }}
{{- end }}
{{- end }}

{{/*
Additional test-connection pod labels
*/}}
{{- define "archimedes-property-operator.testPodLabels" -}}
{{- if .Values.testPodLabels }}
{{- toYaml .Values.testPodLabels }}
{{- end }}
{{- end }}

{{/*
matchLabels
*/}}
{{- define "archimedes-property-operator.matchLabels" -}}
app.kubernetes.io/name: {{ include "archimedes-property-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}


{{/*
Create the name of the service account to use
*/}}
{{- define "archimedes-property-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "archimedes-property-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Additional containers to add to the deployment
*/}}
{{- define "archimedes-property-operator.additionalContainers" -}}
{{- end -}}
