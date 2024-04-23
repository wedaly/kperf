{{- $pattern := .Values.pattern }}
{{- $podSizeInBytes := int .Values.podSizeInBytes }}
{{- range $index := (untilStep 0 (int .Values.total) 1) }}
apiVersion: v1
kind: Namespace
metadata:
  name: {{ $pattern }}-{{ $index }}
  labels:
    name: benchmark-testing
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $pattern }}-{{ $index }}
  namespace: {{ $pattern }}-{{ $index }}
  labels:
    app: {{ $pattern }}
spec:
  replicas: 2000
  strategy:
    rollingUpdate:
      maxSurge: 100
    type: RollingUpdate
  selector:
    matchLabels:
      app: {{ $pattern }}
      index: "{{ $index }}"
  template:
    metadata:
      labels:
        app: {{ $pattern }}
        index: "{{ $index }}"
      annotations:
        data: "{{ randAlphaNum $podSizeInBytes | nospace }}"
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: type
                operator: In
                values:
                - kperf-virtualnodes
      tolerations:
      - key: "kperf.io/nodepool"
        operator: "Exists"
        effect: "NoSchedule"
      containers:
      - name: fake-container
        image: fake-image
---
{{- end}}
