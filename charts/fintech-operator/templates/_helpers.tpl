{{- define "fintech-operator.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fintech-operator.fullname" -}}
{{- if contains .Chart.Name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "fintech-operator.selectorLabels" -}}
control-plane: controller-manager
app.kubernetes.io/name: {{ include "fintech-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "fintech-operator.labels" -}}
{{ include "fintech-operator.selectorLabels" . }}
app.kubernetes.io/version: {{ .ChartAppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "fintech-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "fintech-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
