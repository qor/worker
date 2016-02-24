# Worker

Worker run jobs in background at scheduled time

[![GoDoc](https://godoc.org/github.com/qor/worker?status.svg)](https://godoc.org/github.com/qor/worker)

## Usage

```go
import "github.com/qor/worker"

func main() {
  // Define Worker
  Worker := worker.New()

  // Arguments used to run a job
  type sendNewsletterArgument struct {
    Subject      string
    Content      string `sql:"size:65532"`
    SendPassword string
  }

  // Register Job
  Worker.RegisterJob(&worker.Job{
    Name: "Send Newsletter", // Registerd Job Name
    Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
      // `AddLog` add job log
      qorJob.AddLog("Started sending newsletters...")
      qorJob.AddLog(fmt.Sprintf("Argument: %+v", argument.(*sendNewsletterArgument)))

      for i := 1; i <= 100; i++ {
        time.Sleep(100 * time.Millisecond)
        qorJob.AddLog(fmt.Sprintf("Sending newsletter %v...", i))
        // `SetProgress` set job progress percent, from 0 - 100
        qorJob.SetProgress(uint(i))
      }

      qorJob.AddLog("Finished send newsletters")
      return nil
    },
    // Arguments used to run a job
    Resource: Admin.NewResource(&sendNewsletterArgument{}),
  })

  // Add Worker to qor admin, so you could manage jobs in the admin interface
  Admin.AddResource(Worker)
}
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
