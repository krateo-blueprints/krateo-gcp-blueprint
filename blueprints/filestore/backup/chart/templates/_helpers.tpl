{{/*
Common labels applied to the Config Connector FilestoreBackup resource.
*/}}
{{- define "gcp-filestore-backup.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
