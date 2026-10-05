{{/*
Common labels applied to the Config Connector IdentityPlatformTenant resource.
*/}}
{{- define "gcp-identityplatform-tenant.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
