{{- define "csi-vsphere-conf" -}}
[Global]
cluster-id = "{{ .Values.clusterID }}"

[VirtualCenter "{{ .Values.serverName }}"]
port = "{{ .Values.serverPort }}"
datacenters = "{{ .Values.datacenters }}"
user = "{{ .Values.username }}"
password = "{{ .Values.password }}"
insecure-flag = "{{ .Values.insecureFlag }}"

[Labels]
{{- if .Values.labelRegion }}
region = "{{ .Values.labelRegion }}"
{{- end }}
{{- if .Values.labelZone }}
zone = "{{ .Values.labelZone }}"
{{- end }}
{{- end -}}
