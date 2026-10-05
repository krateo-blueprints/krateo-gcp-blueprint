{{/*
Common labels applied to the Config Connector ServiceIdentity resource.
*/}}
{{- define "gcp-serviceusage-serviceidentity.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
