{{/*
Common labels applied to the Config Connector ComputeSSLCertificate resource.
*/}}
{{- define "gcp-compute-sslcertificate.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
