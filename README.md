# Worker

## Basic Usage

```go
Worker := worker.New()

type sendNewsletterArgument struct {
	Subject      string
	Content      string `sql:"size:65532"`
	SendPassword string
}

Worker.RegisterJob(worker.Job{
	Name: "send_newsletter",
	Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
		qorJob.AddLog("Started sending newsletters...")
		qorJob.AddLog(fmt.Sprintf("Argument: %+v", argument.(*sendNewsletterArgument)))

		for i := 1; i <= 100; i++ {
			time.Sleep(100 * time.Millisecond)
			qorJob.AddLog(fmt.Sprintf("Sending newsletter %v...", i))
			qorJob.SetProgress(uint(i))
		}
		qorJob.AddLog("Finished send newsletters")
		return nil
	},
	Resource: Admin.NewResource(&sendNewsletterArgument{}),
})

// Add to qor admin github.com/qor/qor
Admin.AddResource(Worker)
```
