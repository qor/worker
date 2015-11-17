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
  Queue: Queue,
})

Worker.AddJob(QorJob) // queue -> Add

Worker.RunJob(job.ID)
Worker.KillJob(job.ID)
Worker.DeleteJob(job.ID)
```

## Implement Queue

```go
type Cron struct {
}

func (Cron) Add(QorJob) error {
}

func (Cron) Delete(QorJob) error {
}
```
