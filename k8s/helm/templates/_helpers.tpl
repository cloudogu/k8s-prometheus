{{- define "k8s-prometheus.name" -}}
{{$name := .Chart.Name}}
{{- with get .Values "kube-prometheus-stack" -}}
{{$override := .nameOverride}}
{{- default $name $override | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/* All-in-one labels */}}
{{- define "k8s-prometheus.labels" -}}
app: ces
helm.sh/chart: {{- printf " %s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
k8s.cloudogu.com/component.name: {{ include "k8s-prometheus.name" . }}
k8s.cloudogu.com/component.version: {{ .Chart.AppVersion | quote }}
{{- end -}}

{{- define "prometheus.server.labels" -}}
{{- include "k8s-prometheus.labels" . | nindent 2 }}
{{- include "prometheus.server.selectorLabels" . | nindent 2 }}
{{- end -}}

{{- define "prometheus.server.selectorLabels" -}}
{{- with get .Values "kube-prometheus-stack" -}}
{{- with .prometheus.prometheusSpec.podMetadata.labels -}}
{{ toYaml . }}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "prometheus.node-exporter.labels" -}}
{{- include "k8s-prometheus.labels" . | nindent 2 }}
{{- include "prometheus.node-exporter.selectorLabels" . | nindent 2 }}
{{- end -}}

{{- define "prometheus.node-exporter.selectorLabels" -}}
{{- with get .Values "kube-prometheus-stack" -}}
{{- with get . "prometheus-node-exporter" -}}
{{- with .podLabels -}}
{{ toYaml . }}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "prometheus.kube-state-metrics.labels" -}}
{{- include "k8s-prometheus.labels" . | nindent 2 }}
{{- include "prometheus.kube-state-metrics.selectorLabels" . | nindent 2 }}
{{- end -}}

{{- define "prometheus.kube-state-metrics.selectorLabels" -}}
app.kubernetes.io/name: kube-state-metrics
app.kubernetes.io/instance: k8s-prometheus
{{- end -}}

{{- define "prometheus.operator.labels" -}}
{{ template "prometheus.operator.selectorLabels" . }}
{{ template "k8s-prometheus.labels" . }}
{{- end -}}

{{- define "prometheus.operator.selectorLabels" -}}
app: {{ template "kube-prometheus-stack.name" . }}-operator
app.kubernetes.io/name: {{ template "kube-prometheus-stack.name" . }}-prometheus-operator
app.kubernetes.io/component: prometheus-operator
{{- end -}}




