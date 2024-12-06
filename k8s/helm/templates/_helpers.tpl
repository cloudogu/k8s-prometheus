{{- define "k8s-prometheus.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "prometheus.server.selectorLabels" -}}
{{- with get .Values "kube-prometheus-stack" -}}
{{- with .prometheus.prometheusSpec.podMetadata.labels -}}
{{ toYaml . }}
{{- end -}}
{{- end -}}
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