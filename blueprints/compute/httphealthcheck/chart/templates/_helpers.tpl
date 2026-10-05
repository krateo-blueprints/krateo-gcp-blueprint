{{/*
Common labels applied to the Config Connector ComputeHTTPHealthCheck resource.
*/}}
{{- define "gcp-compute-httphealthcheck.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
