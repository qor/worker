# Kubernetes Job Backend for QOR Worker

This package provides [QOR worker](http://github.com/qor/worker)'s backend based on [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/)

## Usage

```go
kubernetesBackend, err := kubernetes.New(&kubernetes.Config{})
Worker := worker.New(&worker.Config{
  Queue: kubernetesBackend,
})

Worker.RegisterJob(&worker.Job{
  Name: "Send Newsletter",
  Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
})
```

## Advanced Usage


```go
kubernetesBackend, err := kubernetes.New(&kubernetes.Config{
  Namespace: "namespace-used-to-run-job",
  JobTemplate: `
apiVersion: batch/v1
kind: Job
metadata:
  name: jobname
spec:
  template:
    spec:
      containers:
      - name: app
        image: my_image
        resources:
          limits:
            cpu: "750m"
        env:
          - name: DBHost
            value: postgres.db.svc.cluster.local
          - name: DBUser
            value: qor
          - name: DBPassword
            value: qor
`,
})
```
