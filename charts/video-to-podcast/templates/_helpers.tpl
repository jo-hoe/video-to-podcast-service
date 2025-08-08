{{/*
Expand the name of the chart.
*/}}
{{- define "video-to-podcast.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "video-to-podcast.fullname" -}}
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
{{- define "video-to-podcast.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "video-to-podcast.labels" -}}
helm.sh/chart: {{ include "video-to-podcast.chart" . }}
{{ include "video-to-podcast.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "video-to-podcast.selectorLabels" -}}
app.kubernetes.io/name: {{ include "video-to-podcast.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
API Service labels
*/}}
{{- define "video-to-podcast.api.labels" -}}
{{ include "video-to-podcast.labels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
API Service selector labels
*/}}
{{- define "video-to-podcast.api.selectorLabels" -}}
{{ include "video-to-podcast.selectorLabels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
UI Service labels
*/}}
{{- define "video-to-podcast.ui.labels" -}}
{{ include "video-to-podcast.labels" . }}
app.kubernetes.io/component: ui
{{- end }}

{{/*
UI Service selector labels
*/}}
{{- define "video-to-podcast.ui.selectorLabels" -}}
{{ include "video-to-podcast.selectorLabels" . }}
app.kubernetes.io/component: ui
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "video-to-podcast.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "video-to-podcast.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
API Service name
*/}}
{{- define "video-to-podcast.api.serviceName" -}}
{{- printf "%s-api" (include "video-to-podcast.fullname" .) }}
{{- end }}

{{/*
UI Service name
*/}}
{{- define "video-to-podcast.ui.serviceName" -}}
{{- printf "%s-ui" (include "video-to-podcast.fullname" .) }}
{{- end }}

{{/*
API Image name
*/}}
{{- define "video-to-podcast.api.image" -}}
{{- $registry := .Values.global.imageRegistry | default "" }}
{{- $repository := .Values.api.image.repository }}
{{- $tag := .Values.api.image.tag | default .Chart.AppVersion }}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end }}

{{/*
UI Image name
*/}}
{{- define "video-to-podcast.ui.image" -}}
{{- $registry := .Values.global.imageRegistry | default "" }}
{{- $repository := .Values.ui.image.repository }}
{{- $tag := .Values.ui.image.tag | default .Chart.AppVersion }}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end }}

{{/*
API Service URL for internal communication
*/}}
{{- define "video-to-podcast.api.url" -}}
{{- printf "http://%s:%d" (include "video-to-podcast.api.serviceName" .) (.Values.api.service.port | int) }}
{{- end }}

{{/*
Create the configuration for the API service
*/}}
{{- define "video-to-podcast.api.config" -}}
api:
  server:
    port: {{ .Values.api.config.server.port | quote }}
    base_url: {{ .Values.api.config.server.baseUrl | quote }}
  database:
    connection_string: {{ .Values.api.config.database.connectionString | quote }}
  storage:
    base_path: {{ .Values.api.config.storage.basePath | quote }}
  external:
    ytdlp_cookies_file: {{ .Values.api.config.external.ytdlpCookiesFile | quote }}
  feed:
    mode: {{ .Values.api.config.feed.mode | quote }}
{{- end }}

{{/*
Create the configuration for the UI service
*/}}
{{- define "video-to-podcast.ui.config" -}}
ui:
  server:
    port: {{ .Values.ui.config.server.port | quote }}
  api:
    base_url: {{ .Values.ui.config.api.baseUrl | default (include "video-to-podcast.api.url" .) | quote }}
    timeout: {{ .Values.ui.config.api.timeout | quote }}
{{- end }}
