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
