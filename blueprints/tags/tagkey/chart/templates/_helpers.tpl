{{/*
Common labels applied to the Config Connector TagsTagKey resource.
*/}}
{{- define "gcp-tags-tagkey.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
