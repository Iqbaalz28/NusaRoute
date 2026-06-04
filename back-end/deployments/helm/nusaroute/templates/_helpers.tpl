{{/*
Expand the name of the chart.
*/}}
{{- define "nusaroute.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nusaroute.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nusaroute.labels" -}}
helm.sh/chart: {{ include "nusaroute.chart" . }}
{{ include "nusaroute.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nusaroute.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nusaroute.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
