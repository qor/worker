# Worker

## Usage

```go
Worker := worker.New()

type ImportProductArgument struct {
    When *time.Time
    File media_library.FileSystem
}

job := Worker.AddJob(worker.Job{
  Name: "import product",
  Handler: func(record interface{}) {
    // record.(*ImportProductArgument)
    // Do something
  },
  Resource: Admin.NewResource(&ImportProductArgument{}),
  Permission: roles.Permission,
})

job := Worker.GetJob("import product")
job.Run(jobID int)
```
