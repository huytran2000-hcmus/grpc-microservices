{{- define "initContainer.mysqlWait" -}}
- name: mysql-wait
  image: joseluisq/mysql-client
  command: ['mysql', '--host', '{{ .Release.Name }}-mysql', '--user', 'root', '--password', '{{ .Values.mysql.auth.rootPassword }}']
{{- end -}}
