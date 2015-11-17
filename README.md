# Worker

## Usage

```go
Worker := worker.New(db *gorm.DB)
Worker.SetQueue(cron.NewCronQueue())

type ImportProductArgument struct {
    When *time.Time
    File media_library.FileSystem
}

Worker.RegisterJob(worker.Job{
  Name: "import product",
  Handler: func(record interface{}) {
    // record.(*ImportProductArgument)
    // Do something
  },
  Resource: Admin.NewResource(&ImportProductArgument{}),
  Permission: roles.Permission,
  OnKill: func(record interface{}) error {},
})

Worker.AddJob(QorJob)
Worker.RunJob(job.ID)
Worker.KillJob(job.ID)
Worker.PurgeJob(job.ID)
```

## Implement Queue

```go
type Cron struct {
}

func (Cron) Enqueue(QorJob) error {
}

func (Cron) Purge(QorJob) error {
}
```
