{{- define "prometheus.server.selectorLabels" -}}
{{- with get .Values "kube-prometheus-stack" -}}
{{- with .prometheus.prometheusSpec.podMetadata.labels -}}
{{ toYaml . }}
{{- end -}}
{{- end -}}
{{- end -}}

