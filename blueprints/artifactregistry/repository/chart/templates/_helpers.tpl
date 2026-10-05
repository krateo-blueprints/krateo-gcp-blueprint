{{/*
Common labels applied to the Config Connector ArtifactRegistryRepository resource.
*/}}
{{- define "gcp-artifactregistry-repository.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
