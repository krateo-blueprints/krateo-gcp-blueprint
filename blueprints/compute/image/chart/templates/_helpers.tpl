{{/*
Common labels applied to the Config Connector ComputeImage resource.
*/}}
{{- define "gcp-compute-image.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
