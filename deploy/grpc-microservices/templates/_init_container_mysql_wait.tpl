{{- define "initContainer.mysqlWait" -}}
- name: mysql-wait
  # image: joseluisq/mysql-client
  # command: ['sh', '-c', 'until mysql --host {{ .Release.Name }}-mysql --user root --password={{ .Values.mysql.auth.rootPassword }}']
  # image: busybox:1.28
  # command: ['sh', '-c', 'until nslookup {{ .Release.Name }}-mysql; do echo waiting for mysql; sleep 2; done']
  image: busybox:1.28
  command: ['sh', '-c', 'echo -e "Checking for the availability of MySQL Server deployment"; while ! nc -z {{ .Release.Name }}-mysql 3306; do sleep 1; printf "-"; done; echo -e "  >> MySQL DB Server has started";']
{{- end -}}
