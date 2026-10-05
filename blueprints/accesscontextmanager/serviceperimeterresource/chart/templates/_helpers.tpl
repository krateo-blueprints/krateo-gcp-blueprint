{{/*
Common labels applied to the Config Connector AccessContextManagerServicePerimeterResource resource.
*/}}
{{- define "gcp-accesscontextmanager-serviceperimeterresource.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
