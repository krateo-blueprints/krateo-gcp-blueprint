{{/*
Common labels applied to the Config Connector ContainerAttachedCluster resource.
*/}}
{{- define "gcp-containerattached-cluster.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: krateo-gcp-blueprint
krateo.io/composition: {{ .Release.Name }}
{{- end -}}
